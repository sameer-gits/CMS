package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

var rEmail = regexp.MustCompile(`.+@.+\..+`)

func register(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Form error, file might be too large: %v", err), serverCode)
		return
	}
	user := &User{
		Username: r.FormValue("username"),
		Fullname: r.FormValue("fullname"),
		Role:     r.FormValue("role"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}
	profileImage, _, err := r.FormFile("profile_image")
	if err != nil {
		if err == http.ErrMissingFile {
			fmt.Println("No Profile Image")
		} else {
			http.Error(w, fmt.Sprintf("Error processing image: %v", err), serverCode)
			return
		}
	} else {
		defer profileImage.Close()

		user.ProfileImage, err = io.ReadAll(profileImage)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading profile image: %v", err), serverCode)
			return
		}

		filetype := http.DetectContentType(user.ProfileImage)
		if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/gif" {
			http.Error(w, "Format is not allowed. Please upload a JPEG, PNG or GIF", badCode)
			return
		}
		fmt.Println("Profile Image uploaded")
	}

	if !rEmail.MatchString(user.Email) {
		http.Error(w, "Invalid Email", badCode)
		return
	}

	if strings.TrimSpace(user.Username) == "" {
		http.Error(w, "Enter Username", badCode)
		return
	}

	if strings.TrimSpace(user.Fullname) == "" {
		http.Error(w, "Enter Password", badCode)
		return
	}

	if strings.TrimSpace(user.Password) == "" {
		http.Error(w, "Enter Password", badCode)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing password: %v", err), serverCode)
		return
	}

	insertQuery := `
    INSERT INTO users (username, fullname, role, email, profile_image, password_hash)
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING user_id
    `
	err = Conn.QueryRow(context.Background(), insertQuery, user.Username, user.Fullname, user.Role,
		user.Email, user.ProfileImage, hashedPassword).Scan(&user.UserID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error register user: %v", err), serverCode)
		return
	}
	user.Password = ""
	cookieErr := writeCookie(w, user.UserID)
	if cookieErr != nil {
		http.Error(w, cookieErr.Error(), serverCode)
		return
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	username := strings.TrimSpace(r.FormValue("username"))
	if username == "" {
		http.Error(w, "Enter Username", badCode)
		return
	}

	password := strings.TrimSpace(r.FormValue("password"))
	if password == "" {
		http.Error(w, "Enter Password", badCode)
		return
	}

	userID := checkPassword(w, username, password)
	password = ""
	cookieErr := writeCookie(w, userID)
	if cookieErr != nil {
		http.Error(w, cookieErr.Error(), serverCode)
		return
	}
}

func checkPassword(w http.ResponseWriter, username, password string) string {
	selectQuery := `
    SELECT user_id, password_hash FROM users
    WHERE username = $1
    `
	var userID string
	var hashedPassword []byte

	err := Conn.QueryRow(context.Background(), selectQuery, username).Scan(&userID, &hashedPassword)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "User not Found", serverCode)
			return ""
		} else {
			http.Error(w, fmt.Sprintf("Error querying database: %v", err), serverCode)
			return ""
		}
	} else {
		err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
		if err != nil {
			http.Error(w, "Invalid username or password", unauthorized)
			return ""
		}
	}
	return userID
}
