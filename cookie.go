package main

import (
	"errors"
	"net/http"
)

var (
	// secretKey  string
	errLong = errors.New("cookie value too long")
	// errInvalid = errors.New("invalid cookie value")
)

// func base64Encode(input []byte) string {
// 	return strings.TrimRight(base64.URLEncoding.EncodeToString(input), "=")
// }

const cookieName = "cookie"

func cookieSet(w http.ResponseWriter, userID string) error {
	cookie := http.Cookie{
		Name:     cookieName,
		Value:    userID,
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

func cookieGet(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", errors.New("cookie not found")
	}
	userID := cookie.Value
	return userID, nil
}

// func cookieValidate(w http.ResponseWriter) {
// }

func cookieDelete(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:   cookieName,
		MaxAge: -1,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
