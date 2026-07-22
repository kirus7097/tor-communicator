package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	// go run main.go 9090
	if len(os.Args) < 2 {
		fmt.Println("Error. Give the port the server will listen on after it's name")
		os.Exit(1)
	}
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		fmt.Println("Failed to create a certificate or load a key. Details: ", err)
		os.Exit(1)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// MinVersion: tls.VersionTLS13, - could use it if i wanted to use only the stronger TLS version, but not every browser or client supports it. TLS 1.2 is still secure
	}

	port := fmt.Sprintf(":%s", os.Args[1])
	listener, err := tls.Listen("tcp", port, config)
	if err != nil {
		fmt.Println("Failed to create listener. Details:", err)
		os.Exit(1)
	}
	defer listener.Close()
	fmt.Printf("Listening on %s\n", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection, but still listening. Details: ", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				fmt.Println("Failed to read data. Details:", err)
			}
			return
		}

		fmt.Printf("requests: %s", bytes)
		line := fmt.Sprintf("Echo: %s", bytes)
		fmt.Printf("response is %s", line)

		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println("failed to write data, err: ", err)
			return
		}
	}
}
