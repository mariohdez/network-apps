package main

import (
	"fmt"
	"log"
	"net"
)

const (
	udpNetwork = "udp"
)

// $ ifconfig
// en0: flags=8863<UP,BROADCAST,SMART,RUNNING,SIMPLEX,MULTICAST> mtu 1500
//
//	inet 192.168.1.112 netmask 0xffffff00 broadcast 192.168.1.255
func main() {
	addr := "192.168.1.112:8080"
	svrAddr, err := net.ResolveUDPAddr(udpNetwork, addr)
	if err != nil {
		log.Fatalf("resolve udp address=%v", addr, err)
	}

	conn, err := net.ListenUDP(udpNetwork, svrAddr)
	if err != nil {
		log.Fatalf("listen to udp network: %w", err)
	}
	defer conn.Close()

	fmt.Printf("UDP server listening on %v\n", conn.LocalAddr())

	buf := make([]byte, 1024, 1024)
	for {
		n, clntAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatalf("read from udp:%w", err)
		}

		msgRcvd := string(buf[:n])
		fmt.Printf("message=[%s] recieved from %s\n", msgRcvd, clntAddr.String())

		msgToSnd := fmt.Sprintf("ECHO: %v!", msgRcvd)
		_, err := conn.WriteToUDP([]byte(msgToSnd), udpclntAddr)
		if err != nil {
		}
	}
}
