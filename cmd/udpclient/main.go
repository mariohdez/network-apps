package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const (
	udpNetwork = "udp"
)

func main() {
	clientAddr := "192.168.1.112:7070"
	udpClientAddr, err := net.ResolveUDPAddr(udpNetwork, clientAddr)
	if err != nil {
		log.Fatalf("resolve udp client address=%v: %w", clientAddr, err)
	}

	conn, err := net.ListenUDP(udpNetwork, udpClientAddr)
	if err != nil {
		log.Fatalf("listen to UDP network: %w", err)
	}
	defer conn.Close()

	fmt.Printf("UDP client aquired socket binded to %v address\n", conn.LocalAddr())

	srvrAddr := "192.168.1.112:8080"
	udpSrvrAddr, err := net.ResolveUDPAddr(udpNetwork, srvrAddr)
	if err != nil {
		log.Fatalf("resolve udp server address: %w", err)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter message to send to server: ")
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("read string %w", err)
		}

		conn.WriteToUDP([]byte(msg), udpSrvrAddr)
	}
}
