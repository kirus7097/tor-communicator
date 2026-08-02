package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
	"time"
)

func handleConnection(conn net.Conn, database *sql.DB, messageDB *sql.DB) {
	defer conn.Close()
	reader := bufio.NewReaderSize(conn, 4096)
	currentUser := ""
	for {
		conn.SetReadDeadline(time.Now().Add(readTimeout)) // idle clients get dropped

		bytes, err := readLimitedLine(reader, maxLineBytes)
		if err != nil {
			if err != io.EOF {
				slog.Error("Failed to read data", "err", err)
			}
			return
		}
		handleCommand(database, messageDB, string(bytes), &currentUser, conn)
	}
}

// readLimitedLine reads up to '\n' but bails out with an error if the line
// exceeds maxBytes, instead of letting bufio grow the buffer unboundedly.
func readLimitedLine(reader *bufio.Reader, maxBytes int) ([]byte, error) {
	var line []byte
	for {
		chunk, err := reader.ReadSlice('\n')
		line = append(line, chunk...)
		if len(line) > maxBytes {
			// drain the rest of the oversized line so the connection isn't left mid-line
			for err == bufio.ErrBufferFull {
				_, err = reader.ReadSlice('\n')
			}
			return nil, fmt.Errorf("line too long (max %d bytes)", maxBytes)
		}
		if err == bufio.ErrBufferFull {
			continue // keep reading, we haven't hit '\n' yet
		}
		return line, err
	}
}

func handleCommand(database *sql.DB, messageDB *sql.DB, line string, currentUser *string, conn net.Conn) {
	slog.Info("Handling Command")
	parts := strings.Fields(line)
	if len(parts) == 0 {
		conn.Write([]byte("Command can't be empty\n"))
		return
	}
	switch parts[0] {
	case "*REGISTER":
		if len(parts) != 4 {
			conn.Write([]byte("Wrong format. Usage is REGISTER <username> <password>\n"))
			return
		}
		if *currentUser != "" {
			slog.Warn("User tried to reggister while logged in")
			conn.Write([]byte("Log out to register a new user\n"))
			return
		}
		username, password, publicKey := parts[1], parts[2], parts[3]
		conn.Write([]byte(registerUser(database, username, password, publicKey) + "\n"))
		slog.Info("User registered")
		return
	case "*LOGIN":
		if len(parts) != 3 {
			conn.Write([]byte("Wrong format. Usage is LOGIN <username> <password>\n"))
			return
		}
		if *currentUser != "" {
			conn.Write([]byte("Already logged in\n"))
			return
		}
		username, password := parts[1], parts[2]
		ok, err := authenticateUser(database, username, password)
		if err != nil {
			slog.Error("Couldn't log user in:", "err", err)
			conn.Write([]byte("Sorry. Cannot log in at the moment. Please try again later\n"))
			return
		}
		if !ok {
			conn.Write([]byte("Invalid username or password\n"))
			return
		}
		*currentUser = username
		conn.Write([]byte("Logged in successfully\n"))
		slog.Info("User logged in succesfully")
		return
	case "*GETKEY":
		slog.Info("Trying to get the key...")
		if len(parts) != 2 {
			conn.Write([]byte("Wrong format. Usage is *GETKEY <username>\n"))
			return
		}
		username := parts[1]

		key, err := getPublicKey(database, username)
		if err != nil {
			conn.Write([]byte("User not found\n"))
			return
		}
		conn.Write([]byte(key + "\n"))
		slog.Info("Key retrieved from the database. Sent back to client")
		return
	case "*LOGOUT":
		if *currentUser == "" {
			conn.Write([]byte("You are not logged in\n"))
			return
		}
		*currentUser = ""
		conn.Write([]byte("Logged out\n"))
		slog.Info("User logged out")
		return

	case "*MSG":
		if *currentUser == "" {
			conn.Write([]byte("Login first!\n"))
			return
		}
		if len(parts) < 3 {
			conn.Write([]byte("Wrong format. Usage is MSG <username> <message>\n"))
			return
		}
		receiver := parts[1]
		message := strings.Join(parts[2:], " ")
		exists, err := userExists(database, receiver)
		if err != nil {
			conn.Write([]byte("Error checking user\n"))
			return
		}
		if !exists {
			conn.Write([]byte("This user doesn't exist\n"))
			return
		}
		err = sendMessage(messageDB, *currentUser, receiver, message)
		if err != nil {
			conn.Write([]byte("Could not send message\n"))
			return
		}
		conn.Write([]byte("Message sent\n"))
		slog.Info("User sent a message to another user")
		return

	case "*READ":
		if *currentUser == "" {
			conn.Write([]byte("Login first!\n"))
			return
		}
		messages, err := getMessages(messageDB, *currentUser)
		if err != nil {
			conn.Write([]byte("Could not fetch messages\n"))
			return
		}
		slog.Info("Fetched messages")

		if messages == "" {
			conn.Write([]byte("No messages\nEND\n"))
			return
		}

		conn.Write([]byte(messages))
		conn.Write([]byte("END\n"))
		conn.Write([]byte("Messages removed\n"))
		removeMessages(messageDB, *currentUser)
		return

	default:
		conn.Write([]byte("Unknown command\n"))
		return
	}
}
