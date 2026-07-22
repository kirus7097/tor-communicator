package main

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

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
