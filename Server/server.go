package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	readTimeout  = 30 * time.Second // time out, important for resources Usage
	maxLineBytes = 8192             // so attakcer won't send too big data
)

func main() {
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		slog.Error("Couldn't load certificate or server key")
		os.Exit(1)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// go run main.go 9090
	if len(os.Args) < 2 {
		slog.Error("Port not provided")
		os.Exit(1)
	}

	// note - messageDB is for messages, regular database for users and their passwords
	messageDB := initMessagesDatabase() // creates database for messages
	defer messageDB.Close()             // making sure the connection is closed after function ends
	database := initDatabase()          // creating database for users and passwords
	defer database.Close()

	port := fmt.Sprintf("127.0.0.1:%s", os.Args[1])
	listener, err := tls.Listen("tcp", port, config)
	if err != nil {
		slog.Error("Failed to create listener")
		os.Exit(1)
	}
	defer listener.Close()
	slog.Info("Listening", "addr", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Failed to accept connection, but still listening. Details: ", err)
			continue
		}

		go handleConnection(conn, database, messageDB)
	}
}

// TODO: move this function to a different file. someday
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
