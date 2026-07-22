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
	"golang.org/x/crypto/bcrypt"
)

func main() {
	database := initDatabase()
	defer database.Close()

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

		go handleConnection(conn, database)
	}
}

// function creating connection. think database param is not needed?
func handleConnection(conn net.Conn, database *sql.DB) {
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
		response := handleCommand(database, string(bytes))
		line := fmt.Sprintf("%s\n", response)
		fmt.Printf("response is %s", line)

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
	);`

	_, err = database.Exec(createUsersTable)
	if err != nil {
		fmt.Println("Failed when creating users table. Details: ", err)
	}
	fmt.Println("Database created")
	return database
}

func handleCommand(database *sql.DB, line string) string {
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

		err := createUser(database, username, password)
		if err != nil {
			fmt.Println("Something went wrong when registering user data")
			return "ERROR. Could not register user"
		}

		return "User registered successfully!"

	default:
		return "Unknown command!"
	}
}

func createUser(db *sql.DB, username string, password string) error {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}

	_, err = db.Exec(
		"INSERT INTO users(username, password) VALUES (?, ?)",
		username,
		string(hash),
	)

	return err
}
