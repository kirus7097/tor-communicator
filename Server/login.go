package main

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func registerUser(database *sql.DB, registerUsername string, registerPassword string) string {
	exists, err := userExists(database, registerUsername)
	if err != nil {
		fmt.Println("Something went wrong when checking if user exists")
		return "Sorry. Cannot reigster user. Try again later"
	}
	if exists {
		return "Username taken"
	}

	hash, err := bcrypt.GenerateFromPassword(
		[]byte(registerPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		fmt.Println("Cannot create user. Details:", err)
		return "Sorry. Cannot register user. Please try again later"
	}

	_, err = database.Exec(
		"INSERT INTO users(username, password) VALUES (?, ?)",
		registerUsername,
		string(hash),
	)
	if err != nil {
		fmt.Println("Cannot reigster user into database. Details:", err)
		return "Sorry. Cannot register user"
	}

	return "User registered"
}

// checks a username/password pair against the database
func authenticateUser(db *sql.DB, username string, password string) (bool, error) {
	var hash string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hash)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, nil // wrong password, not a real error
	}
	return true, nil
}

// builds the "(username) " prefix used for terminal logging once someone is logged in
func prefix(currentUser string) string {
	if currentUser == "" {
		return ""
	}
	return fmt.Sprintf("(%s) ", currentUser)
}
