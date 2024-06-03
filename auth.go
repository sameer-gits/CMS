package main

import (
	"context"
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

	err := conn.QueryRow(context.Background(),
		selectQuery, userID).Scan(&user.Username, &user.Fullname, &user.Role, &user.Email, &user.ProfileImage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, &CustomError{unauthorized, "invalid user details"}
		}
		return User{}, &CustomError{serverCode, "error validating user"}
	}

	return user, nil
}
