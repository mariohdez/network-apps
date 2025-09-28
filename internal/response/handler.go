package response

import (
	"fmt"
	"net"
	"sync"
)

type Handler struct {
	wg sync.WaitGroup
}

func (h *Handler) Respond(conn *net.UDPConn, buf []byte, clntAddr *net.UDPAddr) {
	h.wg.Add(1)

	go func() {
		defer h.wg.Done()

		msgToSnd := fmt.Sprintf("ECHO: %v!", string(buf[:]))
		_, err := conn.WriteToUDP([]byte(msgToSnd), clntAddr)
		if err != nil {
			fmt.Printf("write to client: %v", err)
		}
	}()
}

func (h *Handler) Wait() {
	h.wg.Wait()
}
