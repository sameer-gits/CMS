package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/sameer-gits/CMS/database"
)

type DbUser struct {
	UserID       uuid.UUID `form:"user_id" json:"user_id" redis:"user_id"`
	Identifier   uuid.UUID `form:"identifier" json:"identifier" redis:"identifier"`
	Username     string    `form:"username" json:"username" redis:"username"`
	Fullname     string    `form:"fullname" json:"fullname" redis:"fullname"`
	Role         rune      `form:"role" json:"role" redis:"role"`
	JoinedAt     time.Time `form:"joined_at" json:"joined_at" redis:"joined_at"`
	Email        string    `form:"email" json:"email" redis:"email"`
	Password     string    `form:"password" json:"password" redis:"password"`
	ProfileImage []byte    `form:"profile_image,omitempty" json:"profile_image,omitempty" redis:"profile_image"`
}

type FormUser struct {
	Username        string `form:"username" json:"username" redis:"username"`
	Fullname        string `form:"fullname" json:"fullname" redis:"fullname"`
	Email           string `form:"email" json:"email" redis:"email"`
	Password        string `form:"password" json:"password" redis:"password"`
	ConfirmPassword string `form:"confirm_password" json:"confirm_password"`
	Message         string `form:"message" json:"message"`
}

type RedisUser struct {
	Username string `form:"username" json:"username" redis:"username"`
	Fullname string `form:"fullname" json:"fullname" redis:"fullname"`
	Email    string `form:"email" json:"email" redis:"email"`
	Otp      int    `form:"otp" json:"otp" redis:"otp"`
	Request  int    `form:"request" json:"request" redis:"request"`
	Blocked  string `form:"blocked" json:"blocked" redis:"blocked"`
	Password string `form:"password" json:"password" redis:"password"`
}

func validateForm(r *http.Request) (FormUser, []error) {
	var errs []error
	var dbexists bool
	ctx := context.Background()

	form := FormUser{
		Username:        r.FormValue("username"),
		Fullname:        r.FormValue("fullname"),
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}

	// check in database if username or email already exists
	checkFormData := `
	SELECT EXISTS (SELECT 1 FROM users WHERE username = $1 OR email = $2);
	`
	err := database.Dbpool.QueryRow(ctx, checkFormData,
		form.Username, form.Email).Scan(&dbexists)
	if err != nil {
		errs = append(errs, errors.New("error checking database for username or email"))
	} else if dbexists {
		errs = append(errs, errors.New("username or email already exists in db"))
	}

	// check in redis if username or email already exists
	var redisexists string
	redisexists, err = database.RedisAllClients.Client0.HGet(ctx, form.Email, "username").Result()
	if err == redis.Nil {
		// user does not exists so continue
	} else if err != nil {
		errs = append(errs, errors.New("redis database error"))
	} else if redisexists != "" {
		errs = append(errs, errors.New("username or email already exists redis"))
	}

	// Username
	if strings.TrimSpace(form.Username) == "" {
		errs = append(errs, errors.New("please provide username"))
	} else if countCharacters(form.Username) < 3 || countCharacters(form.Username) > 66 {
		errs = append(errs, errors.New("usernames must be between 3 to 66 characters long and can only contain letters, numbers, -, _ or max 1 dot in between characters"))
	} else if !isValidUsername(form.Username) {
		errs = append(errs, errors.New("usernames must be between 3 to 66 characters long and can only contain letters, numbers, -, _ or max 1 dot in between characters"))
	}

	// Fullname
	if strings.TrimSpace(form.Fullname) == "" {
		errs = append(errs, errors.New("please provide fullname"))
	} else if countCharacters(form.Fullname) < 2 || countCharacters(form.Fullname) > 66 {
		errs = append(errs, errors.New("fullname must be between 2 and 66 characters"))
	}

	// Email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(form.Email) {
		errs = append(errs, errors.New("please provide valid email"))
	}

	// Password
	if strings.TrimSpace(form.Password) == "" {
		errs = append(errs, errors.New("please provide password"))
	} else if countCharacters(form.Password) < 8 || countCharacters(form.Password) > 18 {
		errs = append(errs, errors.New("password must be between 8 to 18 characters and contain at least one uppercase letter, lowercase letter, number and special character"))
	} else if !hasRequiredPasswordChars(form.Password) {
		errs = append(errs, errors.New("password must be between 8 to 18 characters and contain at least one uppercase letter, lowercase letter, number and special character"))
	}

	//Confirm Password
	if form.ConfirmPassword != form.Password {
		errs = append(errs, errors.New("password and confirm password not matched"))
	}

	return form, errs
}

func isValidUsername(s string) bool {
	periodCount := 0

	for i, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.' && r != '-' {
			return false
		}

		if r == '.' {
			if i == 0 || i == len(s)-1 {
				return false // "." cannot be first or last character
			}
			periodCount++
			if periodCount > 1 {
				return false // more than one "." not allowed
			}
		}
	}
	return true
}

func hasRequiredPasswordChars(password string) bool {
	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasNumber && hasSpecial
}

func (u RedisUser) createUser() (uuid.UUID, []error) {
	var errs []error

	var userID uuid.UUID
	createU := `
	INSERT INTO users (username, fullname, email, password_hash)
    VALUES ($1, $2, $3, $4)
    RETURNING user_id
	`

	err := database.Dbpool.QueryRow(context.Background(), createU,
		u.Username, u.Fullname, u.Email, u.Password).Scan(&userID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "users_username_key" {
					errs = append(errs, errors.New("username already exists please use different username"))
					return uuid.Nil, errs
				}

				if pgErr.ConstraintName == "users_email_key" {
					errs = append(errs, errors.New("email already exists please use different email"))
					return uuid.Nil, errs
				}
			}
		}
		errs = append(errs, errors.New("database error"))
		return uuid.Nil, errs
	}
	return userID, nil
}

func userInfoMiddleware(r *http.Request) (DbUser, error) {
	var user DbUser
	ctx := context.Background()
	userID, err := getCookie(r)
	if err != nil {
		return DbUser{}, err
	}

	// Check Redis 1 if the user is there
	err = database.RedisAllClients.Client1.HGetAll(ctx, userID.String()).Scan(&user)
	if user.Email != "" {
		// if user found in Redis, return
		_ = err
		return user, nil
	}

	// Check if user exists in main DB
	getUser := `
	SELECT username, user_identifier, fullname, role, joined_at, email, profile_image
	FROM users WHERE user_id = $1;
	`
	err = database.Dbpool.QueryRow(ctx, getUser, userID).Scan(
		&user.Username,
		&user.Identifier,
		&user.Fullname,
		&user.Role,
		&user.JoinedAt,
		&user.Email,
		&user.ProfileImage,
	)

	if err != nil {
		return DbUser{}, err
	}

	// add user details in Redis 1 for future
	tx := database.RedisAllClients.Client1.TxPipeline()
	tmpUser := map[string]interface{}{
		"username":      user.Username,
		"fullname":      user.Fullname,
		"role":          user.Role,
		"joined_at":     user.JoinedAt,
		"email":         user.Email,
		"profile_image": user.ProfileImage,
	}

	tx.HSet(ctx, userID.String(), tmpUser).Err()
	tx.Expire(ctx, userID.String(), 15*time.Minute).Err()

	_, err = tx.Exec(ctx)
	if err != nil {
		database.RedisAllClients.Client1.Del(ctx, userID.String())
		log.Println("err creating temp user: %w", err)
		return DbUser{}, err
	}
	user.UserID = uuid.Nil
	return user, nil
}

func getUserIdentifier(ctx context.Context, userIdentifier uuid.UUID) (bool, error) {
	var uIdentifier uuid.UUID
	getUser := `
	SELECT user_identifier
	FROM users WHERE user_identifier = $1;
	`
	err := database.Dbpool.QueryRow(ctx, getUser, userIdentifier).Scan(&uIdentifier)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func getInTableId(ctx context.Context, inTableID uuid.UUID, tableType string) (bool, error) {
	var tableID uuid.UUID
	inTableType := pq.QuoteIdentifier(tableType)
	get := fmt.Sprintf(`
	SELECT %s_id
	FROM %ss WHERE %s_id = $1;
	`, inTableType, inTableType, inTableType)

	err := database.Dbpool.QueryRow(ctx, get, inTableID).Scan(&tableID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func countCharacters(s string) int {
	return utf8.RuneCountInString(s)
}
