package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	readTimeout  = 30 * time.Second // time out, important for resources Usage
	maxLineBytes = 8192             // so attakcer won't send too big data
)

func main() {
	// go run main.go 9090
	if len(os.Args) < 2 {
		fmt.Println("Missing port. Give port after program name.")
		os.Exit(1)
	}

	// note - messageDB is for messages, regular database for users and their passwords
	messageDB := initMessagesDatabase() // creates database for messages
	defer messageDB.Close()             // making sure the connection is closed after function ends
	database := initDatabase()          // creating database for users and passwords
	defer database.Close()

	port := fmt.Sprintf("127.0.0.1:%s", os.Args[1])
	listener, err := net.Listen("tcp", port)
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

		go handleConnection(conn, database, messageDB)
	}
}

// function creating connection.
func handleConnection(conn net.Conn, database *sql.DB, messageDB *sql.DB) {
	defer conn.Close()
	reader := bufio.NewReaderSize(conn, 4096)
	currentUser := ""
	for {
		conn.SetReadDeadline(time.Now().Add(readTimeout)) // idle clients get dropped

		bytes, err := readLimitedLine(reader, maxLineBytes)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Failed to read data. Details:", err)
			}
			return
		}

		fmt.Printf("%srequests: %s", prefix(currentUser), redactForLog(string(bytes)))
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

// redactForLog masks passwords in *LOGIN / *REGISTER commands before they hit stdout.
func redactForLog(line string) string {
	trimmed := strings.TrimRight(line, "\r\n")
	parts := strings.Fields(trimmed)
	if len(parts) >= 2 && (parts[0] == "*LOGIN" || parts[0] == "*REGISTER") {
		return fmt.Sprintf("%s %s [redacted]\n", parts[0], parts[1])
	}
	return line
}

// function creating database
func initDatabase() *sql.DB {
	database, err := sql.Open("sqlite3", "users.db")
	if err != nil {
		fmt.Println("Something went wrong when creating database. Details:", err)
		os.Exit(1)
	}

	// check connection
	err = database.Ping()
	if err != nil {
		fmt.Println("Something went wrong when connecting with database. Details:", err)
		os.Exit(1) // server should close if cannot contact with a database
	}

	createUsersTable := `
CREATE TABLE  IF NOT EXISTS users(
id INTEGER PRIMARY KEY,
username TEXT UNIQUE,
password TEXT NOT NULL,
public_key TEXT NOT NULL
);`

	_, err = database.Exec(createUsersTable)
	if err != nil {
		fmt.Println("Failed when creating users table. Details: ", err)
		os.Exit(1)
	}
	fmt.Println("Database created")
	return database
}

func handleCommand(database *sql.DB, messageDB *sql.DB, line string, currentUser *string, conn net.Conn) {
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
			conn.Write([]byte("Log out to register a new user\n"))
			return
		}
		username, password, publicKey := parts[1], parts[2], parts[3]
		conn.Write([]byte(registerUser(database, username, password, publicKey) + "\n"))
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
			fmt.Println("Something went wrong when logging in:", err)
			conn.Write([]byte("Sorry. Cannot log in at the moment. Please try again later\n"))
			return
		}
		if !ok {
			conn.Write([]byte("Invalid username or password\n"))
			return
		}
		*currentUser = username
		conn.Write([]byte("Logged in successfully\n"))
		return
	case "*GETKEY":
		if len(parts) != 2 {
			conn.Write([]byte("Wrong format. Usage is *GETKEY <username>"))
			return
		}
		username := parts[1]

		key, err := getPublicKey(database, username)
		if err != nil {
			conn.Write([]byte("User not found\n"))
			return
		}
		conn.Write([]byte(key + "\n"))
		return
	case "*LOGOUT":
		if *currentUser == "" {
			conn.Write([]byte("You are not logged in\n"))
			return
		}
		*currentUser = ""
		conn.Write([]byte("Logged out\n"))
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

		if messages == "" {
			conn.Write([]byte("No messages\nEND\n"))
			return
		}

		conn.Write([]byte(messages))
		conn.Write([]byte("END\n"))
		return

	default:
		conn.Write([]byte("Unknown command\n"))
		return
	}
}

func userExists(db *sql.DB, username string) (bool, error) { // username is UNIQUE at the schema level; this gives a user-friendly error path for the client
	var id int
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
