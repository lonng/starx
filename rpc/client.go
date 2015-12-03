package rpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// ServerError represents an error that has been returned from
// the remote side of the RPC connection.
type ServerError string

func (e ServerError) Error() string {
	return string(e)
}

var (
	ErrShutdown        = errors.New("connection is shut down")
	ErrRequestOverFlow = errors.New("request too long")
	ErrEmptyBuffer     = errors.New("empty buffer")
)

// Call represents an active RPC.
type Call struct {
	ServiceMethod string        // The name of the service and method to call.
	Args          []interface{} // The argument to the function (*struct).
	Reply         *[]byte       // The reply from the function (*struct).
	Error         error         // After completion, the error status.
	Done          chan *Call    // Strobes when call is complete.
}

// Client represents an RPC Client.
// There may be multiple outstanding Calls associated
// with a single Client, and a Client may be used by
// multiple goroutines simultaneously.
type Client struct {
	codec *clientCodec

	reqMutex sync.Mutex // protects following
	request  Request

	mutex    sync.Mutex // protects following
	seq      uint64
	pending  map[uint64]*Call
	closing  bool // user has called Close
	shutdown bool // server has told us to stop
}

// A ClientCodec implements writing of RPC requests and
// reading of RPC responses for the client side of an RPC session.
// The client calls WriteRequest to write a request to the connection
// and calls ReadResponseHeader and ReadResponseBody in pairs
// to read responses.  The client calls Close when finished with the
// connection. ReadResponseBody may be called with a nil
// argument to force the body of the response to be read and then
// discarded.
type clientCodec struct {
	rw  io.ReadWriteCloser
	buf []byte // save buffer data
}

// TODO
func (codec *clientCodec) writeRequest(request *Request) error {
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	buf := make([]byte, 0)
	length := len(data)
	for {
		b := byte(length % 128)
		length >>= 7
		if length != 0 {
			buf = append(buf, b+128)
		} else {
			buf = append(buf, b)
			break
		}
	}
	buf = append(buf, data...)
	codec.rw.Write(buf)
	return nil
}

// TODO
func (codec *clientCodec) readResponse(response *Response) error {
	if len(codec.buf) == 0 {
		return ErrEmptyBuffer
	}
	return nil
}

func (codec *clientCodec) close() error {
	return codec.rw.Close()
}

func (client *Client) send(ns string, call *Call) {
	client.reqMutex.Lock()
	defer client.reqMutex.Unlock()

	// Register this call.
	client.mutex.Lock()
	if client.shutdown || client.closing {
		call.Error = ErrShutdown
		client.mutex.Unlock()
		call.done()
		return
	}
	seq := client.seq
	client.seq++
	client.pending[seq] = call
	client.mutex.Unlock()

	// Encode and send the request.
	client.request.Seq = seq
	client.request.ServiceMethod = call.ServiceMethod
	client.request.Args = call.Args
	client.request.Namespace = ns
	err := client.codec.writeRequest(&client.request)
	if err != nil {
		client.mutex.Lock()
		call = client.pending[seq]
		delete(client.pending, seq)
		client.mutex.Unlock()
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (client *Client) input() {
	var err error
	var response Response
	var tmp = make([]byte, 512)
	for err == nil {
		n, err := client.codec.rw.Read(tmp)
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		client.codec.buf = append(client.codec.buf, tmp[:n]...)
		for {
			response = Response{}
			err = client.codec.readResponse(&response)
			if err != nil {
				break
			}
			seq := response.Seq
			client.mutex.Lock()
			call := client.pending[seq]
			delete(client.pending, seq)
			client.mutex.Unlock()

			switch {
			case call == nil:
				// We've got no pending call. That usually means that
				// WriteRequest partially failed, and call was already
				// removed; response is a server telling us about an
				// error reading request body. We should still attempt
				// to read error body, but there's no one to give it to.
				err = errors.New("reading error body")
			case response.Error != "":
				// We've got an error response. Give this to the request;
				// any subsequent requests will get the ReadResponseBody
				// error if there is one.
				call.Error = ServerError(response.Error)
				if err != nil {
					err = errors.New("reading error body: " + err.Error())
				}
				call.done()
			default:
				if err != nil {
					call.Error = errors.New("reading body " + err.Error())
				}
				call.done()
			}
		}
	}
	// Terminate pending calls.
	client.reqMutex.Lock()
	client.mutex.Lock()
	client.shutdown = true
	closing := client.closing
	if err == io.EOF {
		if closing {
			err = ErrShutdown
		} else {
			err = io.ErrUnexpectedEOF
		}
	}
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
	client.mutex.Unlock()
	client.reqMutex.Unlock()
	if debugLog && err != io.EOF && !closing {
		log.Println("rpc: client protocol error:", err)
	}
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here.  It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		if debugLog {
			log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
		}
	}
}

// NewClient returns a new Client to handle requests to the
// set of services at the other end of the connection.
// It adds a buffer to the write side of the connection so
// the header and payload are sent as a unit.
func NewClient(conn io.ReadWriteCloser) *Client {
	client := &Client{
		codec:   &clientCodec{conn, make([]byte, 512)},
		pending: make(map[uint64]*Call),
	}
	go client.input()
	return client
}

// Dial connects to an RPC server at the specified network address.
func Dial(network, address string) (*Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

func (client *Client) Close() error {
	client.mutex.Lock()
	if client.closing {
		client.mutex.Unlock()
		return ErrShutdown
	}
	client.closing = true
	client.mutex.Unlock()
	return client.codec.close()
}

// Go invokes the function asynchronously.  It returns the Call structure representing
// the invocation.  The done channel will signal when the call is complete by returning
// the same Call object.  If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *Client) Go(ns string, service string, method string, reply *[]byte, done chan *Call, args ...interface{}) *Call {
	call := new(Call)
	call.ServiceMethod = service + "." + method
	call.Args = args
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel.  If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	client.send(ns, call)
	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(ns string, service string, method string, reply *[]byte, args ...interface{}) error {
	call := <-client.Go(ns, service, method, reply, make(chan *Call, 1), args).Done
	return call.Error
}
