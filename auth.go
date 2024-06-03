package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func selectUser(userID string, conn *pgx.Conn) (User, error) {
	unauthorized := "Invalid username or password"
	selectQuery := `
	SELECT username, fullname, role, email, profile_image FROM users
	WHERE user_id = $1
	`
	var user User

	err := conn.QueryRow(context.Background(),
		selectQuery, userID).Scan(&user.Username, &user.Fullname, &user.Role, &user.Email, &user.ProfileImage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, fmt.Errorf(unauthorized)
		}
		return User{}, fmt.Errorf("error validating username: %s", err.Error())
	}

	return user, nil
}
