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
		fmt.Println("Missing port. Give port after program name.")
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
		handleCommand(database, messageDB, string(bytes), &currentUser, conn) // converting bytes to text(string)
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
CREATE TABLE IF NOT EXISTS messages(
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL,
receiver TEXT NOT NULL,
message TEXT NOT NULL
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
	if currentUser == nil {
		conn.Write([]byte("Internal server error\n"))
		return
	}
	parts := strings.Fields(line)
	if len(parts) == 0 {
		conn.Write([]byte("Command can't be empty\n"))
		return
	}
	switch parts[0] {
	case "*REGISTER":
		if len(parts) != 3 {
			conn.Write([]byte("Wrong format. Usage is REGISTER <username> <password>\n"))
			return
		}
		if *currentUser != "" {
			conn.Write([]byte("Log out to register a new user\n"))
			return
		}
		username, password := parts[1], parts[2]
		conn.Write([]byte(registerUser(database, username, password) + "\n"))
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

	default:
		if *currentUser == "" {
			conn.Write([]byte("Login first!\n"))
			return
		}
		err := handleTexts(messageDB, *currentUser, line)
		if err != nil {
			fmt.Println("Failed to save message:", err)
			conn.Write([]byte("Could not save message\n"))
			return
		}
		return
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
