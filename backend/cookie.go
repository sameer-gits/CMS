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
	"os"

	"github.com/google/uuid"
)

type Cookie struct {
	UserID uuid.UUID
}

var cookieName = "cookie"

func (c Cookie) createCookie(w http.ResponseWriter) error {
	cookieVal, err := c.encryptCookie()
	if err != nil {
		return err
	}

	cookie := http.Cookie{
		Name:     cookieName,
		Value:    cookieVal,
		Path:     "/",
		MaxAge:   3600 * 24 * 15,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	if len(cookie.String()) > 4096 {
		return errors.New("cookie value max size exceeded")
	}

	http.SetCookie(w, &cookie)
	return nil
}

func deleteCookieHandler(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:   cookieName,
		MaxAge: -1,
		Path:   "/",
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (c Cookie) encryptCookie() (string, error) {
	block, err := aes.NewCipher([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", errors.New("failed to create cipher")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New("failed to create GCM")
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.New("failed to create nonce")
	}

	encryptedData := gcm.Seal(nonce, nonce, []byte(c.UserID.String()), nil)
	return base64.URLEncoding.EncodeToString(encryptedData), nil
}

func decryptCookie(cVal string) (string, error) {
	encryptedData, err := base64Decode(cVal)
	if err != nil {
		return "", errors.New("failed to decode cookie")
	}

	block, err := aes.NewCipher([]byte(os.Getenv("SECRET_KEY")))
	if err != nil {
		return "", errors.New("failed to decrypt cookie block")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New("failed to get decrypt cookie gcm")
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", errors.New("failed to decrypt cookie nonce")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("failed to decrypt cookie as plaintext")
	}

	return string(plaintext), nil
}

// Custom errors are not going to user currently just logging them out for now
func getCookie(r *http.Request) (uuid.UUID, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return uuid.Nil, errors.New("failed to get cookie")
	}
	uID, err := decryptCookie(c.Value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid cookie: %w", err)
	}
	userID, err := uuid.Parse(uID)
	if err != nil {
		return uuid.Nil, errors.New("invalid cookie type")
	}
	return userID, nil
}

func base64Decode(input string) ([]byte, error) {
	value, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return []byte{}, err
	}
	return value, nil
}
