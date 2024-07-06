package main

import (
	"context"

	"github.com/google/uuid"
	"github.com/sameer-gits/CMS/database"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserID       uuid.UUID `form:"user_id" json:"user_id"`
	Username     string    `form:"username" json:"username"`
	Fullname     string    `form:"fullname" json:"fullname"`
	Role         string    `form:"role" json:"role"`
	Email        string    `form:"email" json:"email"`
	Password     string    `form:"password" json:"password"`
	ProfileImage []byte    `form:"profile_image,omitempty" json:"profile_image,omitempty"`
}

func (u User) createUser() (string, error) {
	createU := `
	INSERT INTO users (username, fullname, email, profile_image, password_hash)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING user_id
	`

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	var uuid string
	database.Dbpool.QueryRow(context.Background(), createU,
		u.Username, u.Fullname, u.Email, hashedPassword, u.ProfileImage).Scan(&uuid)
	return uuid, nil
}
