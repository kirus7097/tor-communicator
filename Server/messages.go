package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
)

// function to create database to store messages sent by users
func initMessagesDatabase() *sql.DB {
	database, err := sql.Open("sqlite3", "messages.db")
	if err != nil {
		fmt.Println("Something went wrong when creating database for messages")
		os.Exit(1)
	}

	err = database.Ping()
	if err != nil {
		fmt.Println("Cannot reach database")
		os.Exit(1)
	}

	createMessagesTable := `
CREATE TABLE IF NOT EXISTS messages(
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT NOT NULL,
receiver TEXT NOT NULL,
message TEXT NOT NULL
);`

	_, err = database.Exec(createMessagesTable)
	if err != nil {
		fmt.Println("Failed when creating database for messages")
	}
	fmt.Println("Database for messages created")
	return database
}

func sendMessage(db *sql.DB, sender string, receiver string, message string) error {
	_, err := db.Exec(
		"INSERT INTO messages(username, receiver, message) VALUES (?, ?, ?)",
		sender, receiver, message,
	)
	if err != nil {
		fmt.Println("Failed to insert message into database:", err)
	}
	return err
}

func getMessages(db *sql.DB, username string) (string, error) {
	rows, err := db.Query("SELECT username, message FROM messages WHERE receiver = ?", username)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var sb strings.Builder
	for rows.Next() {
		var sender, message string
		if err := rows.Scan(&sender, &message); err != nil {
			return "", err
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", sender, message))
	}
	return sb.String(), rows.Err()
}
