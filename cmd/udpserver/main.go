package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"network/internal/network"
	"os"
	"os/signal"
	"slices"
	"sync"
)

type srvrReq struct {
	n        int
	clntAddr *net.UDPAddr
	err      error
}

// $ ifconfig
// en0: flags=8863<UP,BROADCAST,SMART,RUNNING,SIMPLEX,MULTICAST> mtu 1500
//
//	inet 192.168.1.112 netmask 0xffffff00 broadcast 192.168.1.255
func main() {
	addr := "192.168.1.112:8080"
	svrAddr, err := net.ResolveUDPAddr(network.UDPNetwork.String(), addr)
	if err != nil {
		log.Fatalf("resolve udp address=%v:%s", addr, err)
	}

	conn, err := net.ListenUDP(network.UDPNetwork.String(), svrAddr)
	if err != nil {
		log.Fatalf("listen to udp network: %s", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig)

	go func() {
		<-sig

		cancel()
	}()

	srvrReqCh := make(chan srvrReq)
	var respHandler ResponseHandler

	reqHandler(ctx, conn, srvrReqCh, &respHandler)

	respHandler.Wait()
	signal.Stop(sig)
}

func reqHandler(ctx context.Context, conn *net.UDPConn, srvrReqCh chan srvrReq, respHandler *ResponseHandler) {
	for {
		buf := make([]byte, 1024)
		go func() {

			n, clntAddr, err := conn.ReadFromUDP(buf)
			srvrReqCh <- srvrReq{
				n:        n,
				clntAddr: clntAddr,
				err:      err,
			}
		}()

		select {
		case <-ctx.Done():
			return
		case srvrReq := <-srvrReqCh:
			if srvrReq.err != nil {
				fmt.Printf("read from udp: %v\n", srvrReq.err)
				return
			}

			bufCpy := slices.Clone(buf[:srvrReq.n])
			respHandler.Respond(conn, bufCpy, srvrReq.clntAddr)
		}
	}
}

type ResponseHandler struct {
	wg sync.WaitGroup
}

func (h *ResponseHandler) Respond(conn *net.UDPConn, buf []byte, clntAddr *net.UDPAddr) {
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

func (h *ResponseHandler) Wait() {
	h.wg.Wait()
}
