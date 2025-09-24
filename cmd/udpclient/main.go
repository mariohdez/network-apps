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

type sender string

const (
	inputSender        sender = "inputSender"
	srvrResponseSender sender = "srvrResponseSender"
)

type output struct {
	sender sender
	msg    string
}

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

	// central error handling and shut down.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error)

	go func() {
		if err := <-errCh; err != nil {
			fmt.Printf("fatal err: %v\n", err)
			cancel()
		}
	}()

	outputCh := make(chan output)
	var consumerWG sync.WaitGroup
	consumerWG.Add(1)
	go func() {
		defer consumerWG.Done()

		outputToStdout(outputCh)
	}()

	var producerWG sync.WaitGroup

	producerWG.Add(1)
	go func() {
		defer producerWG.Done()

		readServerResponse(ctx, conn, outputCh, errCh)
	}()

	srvrAddr := "192.168.1.112:8080"
	udpSrvrAddr, err := net.ResolveUDPAddr(udpNetwork, srvrAddr)
	if err != nil {
		cancel()
		producerWG.Wait()
		os.Exit(1)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	reader := bufio.NewReader(os.Stdin)
	for {
		outputCh <- output{
			sender: inputSender,
			msg:    "Enter message to send to server(EOF to exit program): ",
		}

		msg, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				cancel()
				break
			}

			cancel()
			break
		}

		_, err = conn.WriteToUDP([]byte(strings.TrimSpace(msg)), udpSrvrAddr)
		if err != nil {
			cancel()
			break
		}

		select {
		case <-sigCh:
			cancel()
			producerWG.Wait()
			os.Exit(1)
		default:
			// no-op
		}
	}

	producerWG.Wait()

	close(outputCh)
	consumerWG.Wait()
}

func readServerResponse(ctx context.Context, conn *net.UDPConn, outputCh chan<- output, errCh chan<- error) {
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

				errCh <- fmt.Errorf("read from udp: %w", err)
				return
			}

			msgReceived := string(buf[:n])
			outputCh <- output{
				sender: srvrResponseSender,
				msg: fmt.Sprintf(
					"server[%v] responded with: %s\nEnter message to send to server(EOF to exit program):", addr.String(), msgReceived),
			}
		}
	}
}

func outputToStdout(outputCh <-chan output) {
	for output := range outputCh {

		if output.sender == inputSender {
			fmt.Printf(output.msg)
		} else {
			fmt.Println()
			fmt.Printf(output.msg)
		}
	}
}
