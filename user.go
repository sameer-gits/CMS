package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
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

type CreateUserForm struct {
	Username        string `form:"username" json:"username"`
	Fullname        string `form:"fullname" json:"fullname"`
	Email           string `form:"email" json:"email"`
	Password        string `form:"password" json:"password"`
	ConfirmPassword string `form:"confirmPassword" json:"confirmPassword"`
	ProfileImage    []byte `form:"profile_image,omitempty" json:"profile_image,omitempty"`
}

func (u CreateUserForm) createUser() (string, []error) {
	var errs []error

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		errs = append(errs, errors.New("error saving password"))
		return "", errs
	}

	var uuid string
	createU := `
	INSERT INTO users (username, fullname, email, profile_image, password_hash)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING user_id
	`

	err = database.Dbpool.QueryRow(context.Background(), createU,
		u.Username, u.Fullname, u.Email, u.ProfileImage, hashedPassword).Scan(&uuid)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "users_username_key" {
					errs = append(errs, errors.New("username already exists please use different username"))
					return "", errs
				}

				if pgErr.ConstraintName == "users_email_key" {
					errs = append(errs, errors.New("email already exists please use different email"))
					return "", errs
				}
			}
		}
		errs = append(errs, errors.New("database error"))
		return "", errs
	}
	fmt.Println(uuid)
	return uuid, nil
}

func validateForm(r *http.Request) (CreateUserForm, []error) {
	var errs []error

	form := CreateUserForm{
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

	// Image
	image, _, err := r.FormFile("profileImage")
	if err != nil {
		if err == http.ErrMissingFile {
			form.ProfileImage = nil
			fmt.Println("No Image")
		} else {
			errs = append(errs, err)
		}
	} else {
		image.Close()

		form.ProfileImage, err = io.ReadAll(image)
		if err != nil {
			errs = append(errs, err)
		}

		contentType := http.DetectContentType(form.ProfileImage)
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
			errs = append(errs, errors.New("unsupported image type"))
		}
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
