package starx

import "github.com/gorilla/websocket"

func (hs *handlerService) HandleWS(conn *websocket.Conn) {
	hs.handle(conn)
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
