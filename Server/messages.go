package main

import (
	"database/sql"
	"fmt"
	"os"
)

func handleTexts(database *sql.DB, username string, line string) error {
	_, err := database.Exec(
		"INSERT INTO messages(username, message) VALUES (?, ?)",
		username,
		line,
	)
	if err != nil {
		fmt.Println("Failed to insert message into database: ", err)
	}

	return err
}

// function to create data base to store messages sent by users
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
	message TEXT NOT NULL
	);`

	_, err = database.Exec(createMessagesTable)
	if err != nil {
		fmt.Println("Failed when creating database for messages")
	}
	fmt.Println("Database for messages created")
	return database
}
