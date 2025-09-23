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
)

const (
	udpNetwork = "udp"
)

func main() {
	//	sigChan := make(chan os.Signal)
	//	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	//	quitChan := make(chan struct{})

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

	srvrAddr := "192.168.1.112:8080"
	udpSrvrAddr, err := net.ResolveUDPAddr(udpNetwork, srvrAddr)
	if err != nil {
		log.Fatalf("resolve udp server address=%s: %v", srvrAddr, err)
	}

	pipelineCtx, pipelineCancel := context.WithCancelCause(context.Background())
	defer pipelineCancel(nil)

	errChan := make(chan error)
	var wg sync.WaitGroup

	msgChan := make(chan string)
	wg.Add(1)
	go func() {
		defer wg.Done()
		readUserInput(pipelineCtx, msgChan, errChan)
	}()

	wg.Add(1)
	go func() {
		writeToServer(pipelineCtx, msgChan, conn, udpSrvrAddr, errChan)
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

		fmt.Println("canceling pipeline due to error: %v", err)
	}
}

func readUserInput(ctx context.Context, msgChan chan<- string, errChan chan<- error) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter message to send to server(EOF to exit program): ")
		msg, err := reader.ReadString('\n')
		if err != nil && errors.Is(err, io.EOF) {
			fmt.Println("EOF reached.")
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

func writeToServer(ctx context.Context, msgChan <-chan string, conn *net.UDPConn, udpSrvrAddr *net.UDPAddr, errChan chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				fmt.Println("message channel closed")
				return

			}
			_, err := conn.WriteToUDP([]byte(strings.TrimSpace(msg)), udpSrvrAddr)
			if err != nil {
				errChan <- fmt.Errorf("error writing to UDP server", err)
				return
			}
		}

	}
}
