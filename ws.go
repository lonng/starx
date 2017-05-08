package starx

import (
	"io"
	"net"
	"time"

	"github.com/chrislonng/starx/log"
	"github.com/gorilla/websocket"
)

// wsConn is an adapter to net.Conn, which implements all net.Conn
// interface base on *websocket.Conn
type wsConn struct {
	conn   *websocket.Conn
	typ    int // message type
	reader io.Reader
	writer io.WriteCloser
}

// newWSConn return an initialized *wsConn
func newWSConn(conn *websocket.Conn) (*wsConn, error) {
	c := &wsConn{conn: conn}

	t, r, err := conn.NextReader()
	if err != nil {
		return nil, err
	}

	c.typ = t
	c.reader = r

	w, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return nil, err
	}

	c.writer = w

	return c, nil
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *wsConn) Read(b []byte) (int, error) {
	n, err := c.reader.Read(b)
	if err != nil && err != io.EOF {
		return n, err
	} else if err == io.EOF {
		_, r, err := c.conn.NextReader()
		if err != nil {
			return 0, err
		}
		c.reader = r
	}

	return n, nil
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *wsConn) Write(b []byte) (n int, err error) {
	w, err := c.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}

	c.writer = w

	return w.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *wsConn) Close() error {
	return c.writer.Close()
}

// LocalAddr returns the local network address.
func (c *wsConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *wsConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c *wsConn) SetDeadline(t time.Time) error {
	if err := c.conn.SetReadDeadline(t); err != nil {
		return err
	}

	return c.conn.SetWriteDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *wsConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *wsConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (hs *handlerService) HandleWS(conn *websocket.Conn) {
	c, err := newWSConn(conn)
	if err != nil {
		log.Error(err)
		return
	}
	hs.handle(c)
	/*
		defer conn.Close()

		// message buffer
		packetChan := make(chan *unhandledPacket, packetBufferSize)
		endChan := make(chan bool, 1)

		// all user logic will be handled in single goroutine
		// synchronized in below routine
		go func() {
		loop:
			for {
				select {
				case p := <-packetChan:
					if p != nil {
						hs.processPacket(p.agent, p.packet)
					}
				case <-endChan:
					break loop
				}
			}

		}()

		// register new session when new connection connected in
		agent := defaultNetService.createAgent(conn)
		log.Debug("new agent(%s)", agent.String())
		tmp := make([]byte, 0) // save truncated data
		buf := make([]byte, 512)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Debug("session closed, id: %d, ip: %s", agent.session.Id, agent.socket.RemoteAddr())
				close(packetChan)
				endChan <- true
				agent.close()
				break
			}
			tmp = append(tmp, buf[:n]...)
			var p *packet.Packet // save decoded packet
			for len(tmp) >= packet.HeadLength {
				p, tmp, err = packet.Unpack(tmp)
				if err != nil {
					agent.close()
					break
				}
				packetChan <- &unhandledPacket{agent: agent, packet: p}
			}
		}
	*/
}
