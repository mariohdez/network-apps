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
	"os/signal"
	"strings"
	"sync"
	"syscall"
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

	pipelineCtx, pipelineCancel := context.WithCancelCause(context.Background())
	defer pipelineCancel(nil)

	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	msgChan := make(chan string)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(msgChan)

		readUserInput(pipelineCtx, msgChan, errChan)
	}()

	readSrvrResponseChan := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(readSrvrResponseChan)

		writeToServer(pipelineCtx, msgChan, conn, readSrvrResponseChan, errChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		readServerResponse(pipelineCtx, readSrvrResponseChan, conn, errChan)
	}()

	sigChan := make(chan os.Signal)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer signal.Stop(sigChan)

		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		handleSystemInterrupts(pipelineCtx, sigChan, errChan)
	}()

	go func() {
		defer close(errChan)

		wg.Wait()
		fmt.Println("good bye!")
	}()

	errPipelineCanceled := fmt.Errorf("pipeline canceled: %w", context.Canceled)
	for err := range errChan {
		if err == nil || (errors.Is(err, context.Canceled) && errors.Is(context.Cause(pipelineCtx), errPipelineCanceled)) {
			continue
		}

		fmt.Printf("canceling pipeline due to error: %v\n", err)
		pipelineCancel(err)
	}
}

func readUserInput(ctx context.Context, msgChan chan<- string, errChan chan<- error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter message to send to server(EOF to exit program): ")
		msg, err := reader.ReadString('\n')
		if err != nil && errors.Is(err, io.EOF) {
			errChan <- fmt.Errorf("EOF reached.: %w", err)
			return
		}
		if err != nil {
			errChan <- fmt.Errorf("reading user input: %w", err)
			return
		}

		select {
		case <-ctx.Done():
			return
		case msgChan <- strings.TrimSpace(msg):
			// no-op
		}

	}
}

func writeToServer(ctx context.Context, msgChan <-chan string, conn *net.UDPConn, readSrvrResponseChan chan<- bool, errChan chan<- error) {
	srvrAddr := "192.168.1.112:8080"
	udpSrvrAddr, err := net.ResolveUDPAddr(udpNetwork, srvrAddr)
	if err != nil {
		errChan <- fmt.Errorf("resolve udp server address=%s: %w", srvrAddr, err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}

			_, err := conn.WriteToUDP([]byte(msg), udpSrvrAddr)
			if err != nil {
				errChan <- fmt.Errorf("error writing to UDP server: %w", err)
				return
			}

			readSrvrResponseChan <- true
		}
	}
}

func readServerResponse(ctx context.Context, readSrvrResponseChan <-chan bool, conn *net.UDPConn, errChan chan<- error) {
	buf := make([]byte, 1024, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-readSrvrResponseChan:
			if !ok {
				return
			}

			conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			n, addr, err := conn.ReadFromUDP(buf)
			if err != nil {
				errChan <- fmt.Errorf("reading from client UDP: %w", err)
				return
			}

			msgReceived := string(buf[:n])
			fmt.Printf("server[%v] responded with: %s\n", addr.String(), msgReceived)
		}
	}
}

func handleSystemInterrupts(ctx context.Context, sigChan <-chan os.Signal, errChan chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-sigChan:
			errChan <- fmt.Errorf("system interrupt: %v", sig)
			return
		}
	}
}
