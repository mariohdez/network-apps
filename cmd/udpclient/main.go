package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	udpNetwork = "udp"
)

func main() {
	clientAddr := "192.168.1.112:7070"
	udpClientAddr, err := net.ResolveUDPAddr(udpNetwork, clientAddr)
	if err != nil {
		log.Fatalf("resolve udp client address=%v: %v", clientAddr, err)
	}

	conn, err := net.ListenUDP(udpNetwork, udpClientAddr)
	if err != nil {
		log.Fatalf("listen to UDP network: %v", err)
	}
	defer conn.Close()

	fmt.Printf("UDP client aquired socket binded to %v address\n", conn.LocalAddr())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		readServerResponse(ctx, conn)
	}()

	srvrAddr := "192.168.1.112:8080"
	udpSrvrAddr, err := net.ResolveUDPAddr(udpNetwork, srvrAddr)
	if err != nil {
		cancel()
		wg.Wait()
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter message to send to server(EOF to exit program): ")
		msg, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				cancel()
				break
			}

			cancel()
			wg.Wait()
			os.Exit(1)
		}

		_, err = conn.WriteToUDP([]byte(strings.TrimSpace(msg)), udpSrvrAddr)
		if err != nil {
			cancel()
			wg.Wait()
			os.Exit(1)
		}
	}

	wg.Wait()
	os.Exit(0)
}

func readServerResponse(ctx context.Context, conn *net.UDPConn) {
	buf := make([]byte, 1024)

	for {
		select {
		case <-ctx.Done():
			return
		default:

			conn.SetReadDeadline(time.Now().Add(time.Second))
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}

				fmt.Printf("read from udp: %w\n", err)
				return
			}

			msgReceived := string(buf[:n])
			fmt.Printf("server[%v] responded with: %s\n", addr.String(), msgReceived)

		}
	}
}
