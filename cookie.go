package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	secretKey         = []byte(os.Getenv("SECRET_KEY"))
	errLong           = errors.New("cookie value too long")
	errCookieNotFound = errors.New("cookie not found")
	errInvalid        = errors.New("invalid cookie value")
)

const cookieName = "cookie"

func writeCookie(w http.ResponseWriter, userID string) error {
	data, err := encryptCookieData(userID)
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     cookieName,
		Value:    data,
		Path:     "/",
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	if len(cookie.String()) > 4096 {
		return errLong
	}

	http.SetCookie(w, &cookie)
	return nil
}

func getCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", errCookieNotFound
	}
	userID, err := validateCookieData(cookie.Value)
	if err != nil {
		return "", errInvalid
	}
	return userID, nil
}

func deleteCookie(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:   cookieName,
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func base64Encode(input []byte) string {
	return base64.URLEncoding.EncodeToString(input)
}

func base64Decode(input string) ([]byte, error) {
	value, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return []byte{}, err
	}
	return value, nil
}
func encryptCookieData(userID string) (string, error) {
	hashedsecretKey, err := bcrypt.GenerateFromPassword(secretKey, bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("Error processing cookie")
	}

	hashedCookie, err := bcrypt.GenerateFromPassword([]byte(userID), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("Error processing cookie data")
	}

	secretKeyData := base64Encode(hashedsecretKey)
	cookieData := base64Encode(hashedCookie)
	data := fmt.Sprintf("%s.%s", secretKeyData, cookieData)
	fmt.Println(data)
	return data, nil
}

func validateCookieData(cookieValue string) (string, error) {
	parts := strings.Split(cookieValue, ".")
	if len(parts) != 2 {
		return "", errInvalid
	}
	key := parts[0]
	data := parts[1]

	fmt.Println(data)

	keyByte, keyErr := base64Decode(key)
	if keyErr != nil {
		return "", keyErr
	}
	err := bcrypt.CompareHashAndPassword(keyByte, secretKey)
	if err != nil {
		return "", nil
	}

	return cookieValue, nil
}
