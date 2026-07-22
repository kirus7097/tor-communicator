package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("you didn't provide server ip and the port")
		os.Exit(1)
	}

	server := os.Args[1]
	config := &tls.Config{
		InsecureSkipVerify: true, // it is not secure as the client doesn't really check if the cert is genuine. not for production, but for development okay
	}

	conn, err := tls.Dial("tcp", server, config)
	if err != nil {
		fmt.Println("Connection failed. Details: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Securely connected. Data should be encrypted from now on and not possible to eavesdrop through Wireshark")
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("> ")
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
		fmt.Println("Server: ", string(response[:n]))
	}
}
