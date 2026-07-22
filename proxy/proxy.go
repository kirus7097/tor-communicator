package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// go run proxy.go :8080 127.0.0.1:9090
	listenAddr := os.Args[1] // adres na ktorym proxy slucha
	targetAddr := os.Args[2] // adres do ktorego proxy przekieruje

	l, err := net.Listen("tcp", listenAddr) // gniazdo TCP. proxy nasluchuje na porcie 8080
	if err != nil {
		panic(err)
	}
	fmt.Println("Proxy listening on", listenAddr, "-> forwarding to", targetAddr)

	for {
		client, err := l.Accept() // kazdy klient, ktory sie polaczy, dostaje osobne polaczenie (client), trafia do funkcji (handle)
		if err != nil {
			continue
		}
		go handle(client, targetAddr) // go routine do obslugi wiele klientow na raz. do powtorki
	}
}

func handle(client net.Conn, targetAddr string) {
	defer client.Close()

	server, err := net.Dial("tcp", targetAddr)
	if err != nil {
		fmt.Println("failed to dial target:", err)
		return
	}
	defer server.Close()

	// client -> server (this is where you intercept/modify requests)
	go func() {
		reader := bufio.NewReader(client)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Println("client read error:", err)
				}
				return
			}
			fmt.Printf("[intercepted request] %s", line)

			// --- modify the message here ---
			modified := tamper(line)

			server.Write(modified)
		}
	}()

	// server -> client (responses pass through untouched, or you can tamper here too)
	io.Copy(client, server)
}

func tamper(line []byte) []byte {
	// Example: replace the content, keep it simple to start
	return []byte("HACKED: zmodyfikowana wiadomosc\n")
}
