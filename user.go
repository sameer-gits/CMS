package main

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type DbUser struct {
	UserID       uuid.UUID `form:"user_id" json:"user_id"`
	Username     string    `form:"username" json:"username"`
	Fullname     string    `form:"fullname" json:"fullname"`
	Role         string    `form:"role" json:"role"`
	JoinedAt     time.Time `form:"joined_at" json:"joined_at"`
	Email        string    `form:"email" json:"email"`
	Password     string    `form:"password" json:"password"`
	ProfileImage []byte    `form:"profile_image,omitempty" json:"profile_image,omitempty"`
}

type FormUser struct {
	Username        string `form:"username" json:"username"`
	Fullname        string `form:"fullname" json:"fullname"`
	Email           string `form:"email" json:"email"`
	Password        string `form:"password" json:"password"`
	ConfirmPassword string `form:"confirmPassword" json:"confirmPassword"`
}

type RedisUser struct {
	Username string `form:"username" json:"username" redis:"username"`
	Fullname string `form:"fullname" json:"fullname" redis:"fullname"`
	Email    string `form:"email" json:"email" redis:"email"`
	Otp      int    `form:"otp" json:"otp" redis:"otp"`
	Password string `form:"password" json:"password" redis:"password"`
}

func validateForm(r *http.Request) (FormUser, []error) {
	var errs []error

	form := FormUser{
		Username:        r.FormValue("username"),
		Fullname:        r.FormValue("fullname"),
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirmPassword"),
	}

	// Username
	if strings.TrimSpace(form.Username) == "" {
		errs = append(errs, errors.New("please provide username"))
	} else if len(form.Username) < 3 || len(form.Username) > 66 {
		errs = append(errs, errors.New("usernames must be between 3 to 66 characters long and can only contain letters, numbers, -, _ or max 1 dot in between characters"))
	} else if !isValidUsername(form.Username) {
		errs = append(errs, errors.New("usernames must be between 3 to 66 characters long and can only contain letters, numbers, -, _ or max 1 dot in between characters"))
	}

	// Fullname
	if strings.TrimSpace(form.Fullname) == "" {
		errs = append(errs, errors.New("please provide fullname"))
	} else if len(form.Fullname) < 2 || len(form.Fullname) > 66 {
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
	} else if len(form.Password) < 8 {
		errs = append(errs, errors.New("password must be at least 8 characters long and contain at least one uppercase letter, lowercase letter, number and special character"))
	} else if !hasRequiredPasswordChars(form.Password) {
		errs = append(errs, errors.New("password must be at least 8 characters long and contain at least one uppercase letter, lowercase letter, number and special character"))
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

// func (u RedisUser) createUser() (string, []error) {
// 	var errs []error

// 	var uuid string
// 	createU := `
// 	INSERT INTO users (username, fullname, email, password_hash)
//     VALUES ($1, $2, $3, $4, $5)
//     RETURNING user_id
// 	`

// 	err := database.Dbpool.QueryRow(context.Background(), createU,
// 		u.Username, u.Fullname, u.Email, u.Password).Scan(&uuid)

// 	if err != nil {
// 		var pgErr *pgconn.PgError
// 		if errors.As(err, &pgErr) {
// 			if pgErr.Code == "23505" {
// 				if pgErr.ConstraintName == "users_username_key" {
// 					errs = append(errs, errors.New("username already exists please use different username"))
// 					return "", errs
// 				}

// 				if pgErr.ConstraintName == "users_email_key" {
// 					errs = append(errs, errors.New("email already exists please use different email"))
// 					return "", errs
// 				}
// 			}
// 		}
// 		errs = append(errs, errors.New("database error"))
// 		return "", errs
// 	}
// 	fmt.Println(uuid)
// 	return uuid, nil
// }
