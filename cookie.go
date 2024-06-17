package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

var (
	errLong           = errors.New("cookie value too long")
	errCookieNotFound = errors.New("cookie not found")
	errInvalid        = errors.New("invalid cookie value")
)

const cookieName = "cookie"

func writeCookie(w http.ResponseWriter, userID uuid.UUID) error {
	data, err := encryptCookieData(userID)
	if err != nil {
		return fmt.Errorf("failed to encrypt cookie data: %w", err)
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
		Path:   "/",
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/login", http.StatusFound)
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

func encryptCookieData(data uuid.UUID) (string, error) {
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	encryptedData := gcm.Seal(nonce, nonce, []byte(data.String()), nil)
	return base64Encode(encryptedData), nil
}

func validateCookieData(data string) (string, error) {
	encryptedData, err := base64Decode(data)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", errInvalid
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	return string(plaintext), nil
}
