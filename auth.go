package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func login(w http.ResponseWriter, r *http.Request) error {
	username := strings.TrimSpace(r.FormValue("username"))
	password := strings.TrimSpace(r.FormValue("password"))

	if username == "" {
		http.Error(w, "Enter username", badCode)
	}

	if password == "" {
		http.Error(w, "Enter password", badCode)
	}

	userID, hashedPassword, err := checkPassword(username)
	if err != nil {
		http.Error(w, err.Error(), unauthorized)
		return err
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		http.Error(w, "Invalid username or password", unauthorized)
	}
	user := User{UserID: userID}
	fmt.Printf("User logged in: %+v\n", user)

	w.WriteHeader(statusOK)
	return nil
}

func checkPassword(username string) (string, []byte, error) {
	selectQuery := `
    SELECT user_id, password_hash FROM users
    WHERE username = $1
    `

	var userID string
	var hashedPassword []byte

	err := Conn.QueryRow(context.Background(), selectQuery, username).Scan(&userID, &hashedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil, fmt.Errorf("user not found")
		}
		return "", nil, fmt.Errorf("error querying database: %v", err)
	}

	return userID, hashedPassword, nil
}
