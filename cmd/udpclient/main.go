package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"network/internal/network"
	"os"
	"strings"
	"sync"
)

func main() {
	udpClientAddr := resolveUDPAddr(network.UDPNetwork.String(), "192.168.1.112:7070")
	conn, err := net.ListenUDP(network.UDPNetwork.String(), udpClientAddr)
	if err != nil {
		log.Fatalf("listen to UDP network: %v", err)
	}

	errCh := make(chan error)
	go func() {
		for err := range errCh {
			fmt.Printf("handling err: %s\n", err)
		}
	}()

	var wg sync.WaitGroup
	reqMsgCh := make(chan string)
	wg.Add(1)
	go func() {
		defer wg.Done()

		udpSrvrAddr := resolveUDPAddr(network.UDPNetwork.String(), "192.168.1.112:8080")
		makeUDPRequests(conn, udpSrvrAddr, reqMsgCh, errCh)
	}()

	outputCh := make(chan string)

	wg.Add(1)
	go func() {
		defer wg.Done()

		readUDPResponses(conn, outputCh, errCh)
	}()

	go func() {
		outputToStdout(outputCh)
	}()

	getUserInput(reqMsgCh, outputCh, errCh)
	close(reqMsgCh)
	conn.Close()
	wg.Wait()
	close(errCh)
	close(outputCh)
}

func getUserInput(reqMsgCh chan<- string, outputCh chan<- string, errCh chan<- error) {

	reader := bufio.NewReader(os.Stdin)
	for {
		outputCh <- "Enter message to send to server(EOF to exit program)"
		line, err := reader.ReadString('\n')
		if err != nil {
			errCh <- fmt.Errorf("read string: %w", err)
			return
		}

		reqMsgCh <- strings.TrimSpace(line)
	}
}

func makeUDPRequests(conn *net.UDPConn, updSrvrAddr *net.UDPAddr, reqMsgCh <-chan string, errCh chan<- error) {
	for {
		msg, ok := <-reqMsgCh
		if !ok {
			return
		}

		_, err := conn.WriteToUDP([]byte(msg), updSrvrAddr)
		if err != nil {
			errCh <- fmt.Errorf("write to udp: %w", err)
			return
		}
	}
}

func readUDPResponses(conn *net.UDPConn, outputCh chan<- string, errCh chan<- error) {
	buf := make([]byte, 1024)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			errCh <- fmt.Errorf("read udp: %w", err)
			return
		}

		msgReceived := string(buf[:n])
		outputCh <- fmt.Sprintf(
			"server[%v] responded with: %s\nEnter message to send to server(EOF to exit program):", addr.String(), msgReceived)
	}
}

func outputToStdout(outputCh <-chan string) {
	for output := range outputCh {
		fmt.Println(output)
	}
}

func resolveUDPAddr(network string, addr string) *net.UDPAddr {
	udpAddr, err := net.ResolveUDPAddr(network, addr)
	if err != nil {
		log.Fatalf("resolve udp addr: %s", err)
	}

	return udpAddr
}
