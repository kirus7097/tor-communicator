package main

import (
	"database/sql"
	"fmt"
	"strings"
)

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
