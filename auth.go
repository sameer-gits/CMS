package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

type CustomError struct {
	StatusCode int
	Message    string
}

func (e *CustomError) Error() string {
	return e.Message
}

func selectUser(userID string, conn *pgx.Conn) (User, error) {
	unauthorized := http.StatusUnauthorized
	serverCode := http.StatusInternalServerError
	selectQuery := `
	SELECT username, fullname, role, email, profile_image FROM users
	WHERE user_id = $1
	`
	var user User
	fmt.Println(userID)

	err := conn.QueryRow(context.Background(),
		selectQuery, userID).Scan(&user.Username, &user.Fullname, &user.Role, &user.Email, &user.ProfileImage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, &CustomError{unauthorized, fmt.Sprintf("error validating user %v", err)}
		}
		return User{}, &CustomError{serverCode, fmt.Sprintf("error validating user %v", err)}
	}

	fmt.Println(user)
	return user, nil
}
