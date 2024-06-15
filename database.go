package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

var Conn *pgx.Conn

type CustomError struct {
	StatusCode int
	Message    string
}

func (e *CustomError) Error() string {
	return e.Message
}

func InitDB() error {
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connected to database")
	Conn = conn
	return nil
}

func selectUser(userID string) (User, error) {
	selectQuery := `
	SELECT username, fullname, role, email, profile_image FROM users
	WHERE user_id = $1
	`
	var user User

	err := Conn.QueryRow(context.Background(),
		selectQuery, userID).Scan(&user.Username, &user.Fullname, &user.Role, &user.Email, &user.ProfileImage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, &CustomError{unauthorized, fmt.Sprintf("error validating user %v", err)}
		}
		return User{}, &CustomError{serverCode, fmt.Sprintf("error validating user %v", err)}
	}

	return user, nil
}
