package main

import (
	"database/sql"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
)

func registerUser(database *sql.DB, registerUsername string, registerPassword string, publicKey string) string {
	exists, err := userExists(database, registerUsername)
	if err != nil {
		slog.Warn("Couldn't check if the user exists")
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
		slog.Error("Couldn't save encrypted password", "err:", err)
		return "Sorry. Cannot register user. Please try again later"
	}

	_, err = database.Exec(
		"INSERT INTO users(username, password, public_key) VALUES (?, ?, ?)",
		registerUsername,
		string(hash),
		publicKey,
	)
	if err != nil {
		slog.Error("Cannot reigster user into database", "err:", err)
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
	slog.Info("User logged in")
	return true, nil
}

func getPublicKey(db *sql.DB, username string) (string, error) {
	var publicKey string

	err := db.QueryRow(
		"SELECT public_key FROM users WHERE username = ?",
		username,
	).Scan(&publicKey)

	if err == sql.ErrNoRows {
		return "", fmt.Errorf("user not found")
	}

	if err != nil {
		return "", err
	}

	return publicKey, nil
}
