package main

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// go run main.go 9090
	if len(os.Args) < 2 {
		fmt.Println("Error. Give the port the server will listen on after it's name")
		os.Exit(1)
	}

	// note - messageDB is for messages, regular database for users and their passwords
	messageDB := initMessagesDatabase() // creates database for messages
	defer messageDB.Close()             // making sure the connection is closed after function ends
	database := initDatabase()          // creating database for users and passwords
	defer database.Close()

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

		go handleConnection(conn, database, messageDB)
	}
}

// function creating connection. think database param is not needed? - not anymore
func handleConnection(conn net.Conn, database *sql.DB, messageDB *sql.DB) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	currentUser := ""
	for {
		bytes, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err != io.EOF {
				fmt.Println("Failed to read data. Details:", err)
			}
			return
		}
		fmt.Printf("%srequests: %s", prefix(currentUser), bytes)
		response := handleCommand(database, messageDB, string(bytes), &currentUser) // converting bytes to text(string)
		line := fmt.Sprintf("%s\n", response)                                       //line is equal to response
		fmt.Printf("%sresponse is %s", prefix(currentUser), line)                   // print out as a log to server
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Println("failed to write data, err: ", err)
			return
		}
	}
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
		println("Something went wrong when connecting with database. Details:", err)
		os.Exit(1) // server should close if cannot contact with a database
	}

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL
	);` // it's written in SQL

	_, err = database.Exec(createUsersTable)
	if err != nil {
		fmt.Println("Failed when creating users table. Details: ", err)
		os.Exit(1)
	}
	fmt.Println("Database created")
	return database
}

func handleCommand(database *sql.DB, messageDB *sql.DB, line string, currentUser *string) string {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return "Command can't be empty"
	}
	switch parts[0] {
	case "REGISTER":
		if len(parts) != 3 {
			return "ERROR. usage is REGISTER <username> <password>"
		}
		username, password := parts[1], parts[2]
		registerUser(database, username, password)
		return "User registered!"
	case "LOGIN":
		if len(parts) != 3 {
			return "ERROR. usage is LOGIN <username> <password>"
		}
		username, password := parts[1], parts[2]
		ok, err := authenticateUser(database, username, password)
		if err != nil {
			fmt.Println("Something went wrong when logging in")
			return "ERROR. Could not log in"
		}
		if !ok {
			return "ERROR. Invalid username or password"
		}
		*currentUser = username
		return fmt.Sprintf("You are now logged as %s", username)
	default:
		if currentUser != nil && *currentUser != "" {
			err := handleTexts(messageDB, *currentUser, line)
			if err != nil {
				fmt.Println("Failed to save message:", err)
				return "ERROR. Could not save message"
			}
			return "Message sent!"
		}
		return "Unknown command!"
	}
}

func userExists(db *sql.DB, username string) (bool, error) { // i actually secured that username had to be unique when creating table. but this function gives user-friendly error for client
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
