package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

type Forum struct {
	ID        uuid.UUID
	Name      string
	Image     []byte
	Public    bool
	CreatedAt time.Time
	CreatedBy string
}

type CustomError struct {
	StatusCode int
	Message    string
}

func (e *CustomError) Error() string {
	return e.Message
}

func InitDB() error {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse database URL: %v\n", err)
		os.Exit(1)
	}

	config.MaxConns = 10
	config.MinConns = 5
	config.MaxConnIdleTime = 60

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connected to database")
	Pool = pool
	return nil
}

func selectUser(userID string) (User, error) {
	selectQuery := `
	SELECT username, fullname, role, email, profile_image FROM users
	WHERE user_id = $1
	`
	var user User

	err := Pool.QueryRow(context.Background(),
		selectQuery, userID).Scan(&user.Username, &user.Fullname, &user.Role, &user.Email, &user.ProfileImage)
	if err != nil {
		if err == pgx.ErrNoRows {
			return User{}, &CustomError{unauthorized, fmt.Sprintf("error validating user: %v", err)}
		}
		return User{}, &CustomError{serverCode, fmt.Sprintf("error validating user %v", err)}
	}

	return user, nil
}

func newForum(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(contextKey).(string)
	forum := Forum{
		Name: r.FormValue("forum_name"),
	}

	forumImage, _, err := r.FormFile("forum_image")
	if err != nil {
		if err == http.ErrMissingFile {
			http.Error(w, "Please upload an JPEG, PNG or GIF", badCode)
		} else {
			http.Error(w, fmt.Sprintf("Error processing image: %v", err), serverCode)
		}
	} else {
		defer forumImage.Close()

		forum.Image, err = io.ReadAll(forumImage)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error reading profile image: %v", err), serverCode)
		}

		filetype := http.DetectContentType(forum.Image)
		if filetype != "image/jpeg" && filetype != "image/png" && filetype != "image/gif" {
			http.Error(w, "Format is not allowed. Please upload a JPEG, PNG or GIF", badCode)
		}
		fmt.Println("Forum Image uploaded")
	}

	insertQuery := `
    INSERT INTO forums (forum_name, forum_image, created_by)
    VALUES ($1, $2, $3)
    RETURNING forum_id
    `

	err = Pool.QueryRow(context.Background(), insertQuery, forum.Name, forum.Image, userID).Scan(
		&forum.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error Creating Forum: %v", err), serverCode)
	}
}

func selectForums() ([]Forum, error) {
	selectQuery := `
    SELECT forum_id, forum_name, forum_image, public, created_at, created_by
    FROM forums
    `
	rows, err := Pool.Query(context.Background(), selectQuery)
	if err != nil {
		return nil, &CustomError{serverCode, fmt.Sprintf("error retrieving forums: %v", err)}
	}
	defer rows.Close()

	var forums []Forum
	for rows.Next() {
		var forum Forum
		err := rows.Scan(&forum.ID, &forum.Name, &forum.Image, &forum.Public, &forum.CreatedAt, &forum.CreatedBy)
		if err != nil {
			return nil, &CustomError{serverCode, fmt.Sprintf("error scanning forum: %v", err)}
		}
		forums = append(forums, forum)
	}

	if rows.Err() != nil {
		return nil, &CustomError{serverCode, fmt.Sprintf("error iterating rows: %v", rows.Err())}
	}
	return forums, nil
}
