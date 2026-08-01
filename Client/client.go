package main

import (
	"bufio"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"golang.org/x/net/proxy"
)

func sendLine(conn net.Conn, serverReader *bufio.Reader, line string) (string, error) {
	// send command/message to server
	_, err := conn.Write([]byte(line))
	if err != nil {
		return "", err
	}

	// wait for server response
	return serverReader.ReadString('\n')
}

// ask server for someone's public key
// needed before sending encrypted message
//
// every user has their own public key stored on server
// because people need it to encrypt messages to that user
func fetchPublicKey(
	conn net.Conn,
	serverReader *bufio.Reader,
	target string,
) (*[32]byte, error) {
	reply, err := sendLine(
		conn,
		serverReader,
		fmt.Sprintf("*GETKEY %s\n", target),
	)
	if err != nil {
		return nil, err
	}

	// remove \n and spaces from server response
	keyHex := strings.TrimSpace(reply)

	// convert public key from text back to bytes
	keyBytes, err := hex.DecodeString(keyHex)

	if err != nil || len(keyBytes) != 32 {
		return nil, fmt.Errorf(
			"invalid public key from server for %s",
			target,
		)
	}

	var pub [32]byte
	copy(pub[:], keyBytes)

	return &pub, nil
}

// TODO: remove hardcoded server address
// better idea:
// - read from config file
// - ask user during setup
// - maybe store after first connection
// TODO
// the prototyping stage. Fine for now since the address isn't a secret.
const hardcodedServer = "5jppyjmpqxcqmq4y27lw5ld24kfktlb6tn4rbzhspqyhdw7xeqrfrnid.onion:9090"

func main() {
	// load existing identity or create a new one
	// this happens before connecting because we need our public key for registration
	myPub, myPriv, err := LoadOrCreateKeypair()
	if err != nil {
		fmt.Println("Could not load or generate keypair:", err)
		os.Exit(1)
	}

	// use default server unless user provides another one
	server := hardcodedServer

	if len(os.Args) >= 2 {
		server = os.Args[1]
	}

	fmt.Println("Connecting through Tor...")

	// create SOCKS5 connection through local Tor service
	// Tor normally listens on port 9050
	//
	// NOTE:
	// without this, connection would go directly to server
	// which would reveal our IP address
	dialer, err := proxy.SOCKS5(
		"tcp",
		"127.0.0.1:9050",
		nil,
		proxy.Direct,
	)
	if err != nil {
		fmt.Println("Failed to create SOCKS5 proxy:", err)
		os.Exit(1)
	}

	// connect to server using Tor
	rawConn, err := dialer.Dial("tcp", server)
	if err != nil {
		fmt.Println("Connection failed:", err)
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		ServerName: strings.Split(server, ":")[0],

		// Temporary for .onion services without public certificates
		InsecureSkipVerify: true,
	}

	conn := tls.Client(rawConn, tlsConfig)

	err = conn.Handshake()
	if err != nil {
		fmt.Println("TLS handshake failed:", err)
		rawConn.Close()
		os.Exit(1)
	}

	defer conn.Close()
	fmt.Println("Connected.")
	fmt.Println()

	fmt.Println("Tor Communicator")
	fmt.Println("----------------")
	fmt.Println("*REGISTER <username> <password>")
	fmt.Println("*LOGIN <username> <password>")
	fmt.Println("*LOGOUT")
	fmt.Println("*MSG <target> <message>")
	fmt.Println("*READ")
	fmt.Println()

	serverReader := bufio.NewReader(conn)
	inputReader := bufio.NewReader(os.Stdin)

	for {

		fmt.Print("> ")

		// read user command
		message, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println("Input closed:", err)
			return
		}

		// remove newline characters
		trimmed := strings.TrimRight(message, "\r\n")

		// split command into pieces
		// example:
		// *MSG bob hello -> ["*MSG", "bob", "hello"]
		parts := strings.Fields(trimmed)

		if len(parts) == 0 {
			continue
		}

		var reply string

		switch parts[0] {

		case "*REGISTER":

			// username + password required
			if len(parts) != 3 {
				fmt.Println("Usage: *REGISTER <username> <password>")
				continue
			}

			// server needs our public key
			// so other users can encrypt messages to us
			pubHex := hex.EncodeToString(myPub[:])

			reply, err = sendLine(
				conn,
				serverReader,
				fmt.Sprintf(
					"*REGISTER %s %s %s\n",
					parts[1],
					parts[2],
					pubHex,
				),
			)

		case "*MSG":

			if len(parts) < 3 {
				fmt.Println("Usage: *MSG <target> <message>")
				continue
			}

			target := parts[1]

			// message can contain spaces
			plaintext := strings.Join(parts[2:], " ")

			// get receiver public key first
			// encryption requires receiver's public key
			recipPub, keyErr := fetchPublicKey(
				conn,
				serverReader,
				target,
			)

			if keyErr != nil {
				fmt.Println("Could not send message:", keyErr)
				continue
			}

			// encrypt message locally before sending
			// server only sees encrypted data
			ciphertext, encErr := EncryptMessage(
				[]byte(plaintext),
				recipPub,
				myPriv,
			)

			if encErr != nil {
				fmt.Println("Could not encrypt message:", encErr)
				continue
			}

			reply, err = sendLine(
				conn,
				serverReader,
				fmt.Sprintf(
					"*MSG %s %s\n",
					target,
					ciphertext,
				),
			)

		case "*READ":

			// get encrypted messages waiting on server
			messages, readErr := readMessages(
				conn,
				serverReader,
			)

			if readErr != nil {
				fmt.Println("Could not read messages:", readErr)
				continue
			}

			if len(messages) == 0 {
				fmt.Println("No messages")
				continue
			}

			for _, msg := range messages {

				// need sender public key to verify/decrypt message
				senderPub, keyErr := fetchPublicKey(
					conn,
					serverReader,
					msg.Sender,
				)

				if keyErr != nil {
					fmt.Println(
						"Could not get public key for",
						msg.Sender,
						":",
						keyErr,
					)
					continue
				}

				plaintext, decryptErr := decryptMessage(
					msg.Ciphertext,
					senderPub,
					myPriv,
				)

				if decryptErr != nil {
					fmt.Println(
						"Could not decrypt message from",
						msg.Sender,
						":",
						decryptErr,
					)
					continue
				}

				fmt.Printf(
					"%s: %s\n",
					msg.Sender,
					plaintext,
				)
			}

			continue

		default:

			// commands like LOGIN, LOGOUT, etc.
			// are sent directly to server
			reply, err = sendLine(
				conn,
				serverReader,
				message,
			)
		}

		if err != nil {

			if err == io.EOF {
				fmt.Println("Server closed connection.")
			} else {
				fmt.Println(
					"Could not read server response:",
					err,
				)
			}

			return
		}

		fmt.Print("Server: ", reply)
	}
}

func readMessages(
	conn net.Conn,
	serverReader *bufio.Reader,
) ([]Message, error) {
	// ask server for unread messages
	_, err := conn.Write([]byte("*READ\n"))
	if err != nil {
		return nil, err
	}

	var messages []Message

	for {

		// server sends messages line by line
		line, err := serverReader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)

		// server uses END to tell us there are no more messages
		if line == "END" {
			break
		}

		// expected format:
		// username: encrypted_message
		//
		// SplitN because encrypted message itself could theoretically contain ":"
		parts := strings.SplitN(
			line,
			": ",
			2,
		)

		if len(parts) != 2 {
			continue
		}

		messages = append(
			messages,
			Message{
				Sender:     parts[0],
				Ciphertext: parts[1],
			},
		)
	}

	return messages, nil
}
