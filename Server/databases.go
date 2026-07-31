package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const databaseDir = "Databases"

func createDatabaseDirectory() {
	err := os.MkdirAll(databaseDir, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to create database directory. Details:", err)
		os.Exit(1)
	}
}

func initDatabase() *sql.DB {
	createDatabaseDirectory()

	database, err := sql.Open("sqlite3", databaseDir+"/users.db")
	if err != nil {
		fmt.Println("Something went wrong when opening users database. Details:", err)
		os.Exit(1)
	}

	err = database.Ping()
	if err != nil {
		fmt.Println("Something went wrong when connecting to users database. Details:", err)
		os.Exit(1)
	}

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		public_key TEXT NOT NULL
	);`

	_, err = database.Exec(createUsersTable)
	if err != nil {
		fmt.Println("Failed when creating users table. Details:", err)
		database.Close()
		os.Exit(1)
	}

	fmt.Println("Users database initialized successfully")
	return database
}

func initMessagesDatabase() *sql.DB {
	createDatabaseDirectory()

	database, err := sql.Open("sqlite3", databaseDir+"/messages.db")
	if err != nil {
		fmt.Println("Something went wrong when opening messages database. Details:", err)
		os.Exit(1)
	}

	err = database.Ping()
	if err != nil {
		fmt.Println("Something went wrong when connecting to messages database. Details:", err)
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
		fmt.Println("Failed when creating messages table. Details:", err)
		database.Close()
		os.Exit(1)
	}

	fmt.Println("Messages database initialized successfully")
	return database
}
