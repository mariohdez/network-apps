package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"network/internal/network"
	"network/internal/response"
	"os"
	"os/signal"
	"slices"
	"syscall"
)

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

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig

		conn.Close()
	}()

	respHandler := response.NewHandler()

	reqHandler(conn, respHandler)

	respHandler.Wait()
	signal.Stop(sig)
	close(sig)
}

func reqHandler(conn *net.UDPConn, respHandler *response.Handler) {
	buf := make([]byte, 1024)
	for {
		n, clntAddr, err := conn.ReadFromUDP(buf)

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			fmt.Printf("read from UDP: %s\n", err)
			return
		}

		respHandler.Respond(conn, slices.Clone(buf[:n]), clntAddr)
	}
}
