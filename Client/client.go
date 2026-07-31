package main

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/net/proxy"
)

// TODO
// the prototyping stage. Fine for now since the address isn't a secret.
const hardcodedServer = "5jppyjmpqxcqmq4y27lw5ld24kfktlb6tn4rbzhspqyhdw7xeqrfrnid.onion:9090"

func main() {
	server := hardcodedServer
	if len(os.Args) >= 2 {
		server = os.Args[1] // still allow overriding for testing against a different instance
	}

	// SOCKS5 dialer pointed at the local Tor daemon's SocksPort (default 127.0.0.1:9050).
	// This is what routes the connection through the Tor network instead of hitting
	// the network directly - no TLS needed on top, the onion service handles that.
	dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:9050", nil, proxy.Direct)
	if err != nil {
		fmt.Println("Failed to create SOCKS5 dialer. Is Tor running? Details:", err)
		os.Exit(1)
	}

	conn, err := dialer.Dial("tcp", server)
	if err != nil {
		fmt.Println("Connection failed. Details: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Welcome to the Tor Communicator.")
	fmt.Println("Register usage: *REGISTER <username> <password>")
	fmt.Println("Login usage: *LOGIN <username> <password>")
	fmt.Println("Log out usage: *LOGOUT")
	fmt.Println("Texting usage: *MSG <target> <message>")

	reader := bufio.NewReader(os.Stdin)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			return // so it won't stop and keep doing
		}
		_, err = conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Could not write to the server. Details: ", err)
			return
		}
		response := make([]byte, 1024)
		n, err := conn.Read(response)
		if err != nil {
			fmt.Println("Could not display server response. Details:", err)
			return
		}
		reply := string(response[:n])
		fmt.Println("Server: ", reply)
	}
}
