package main

import (
	"errors"
	"net/http"
)

type Cookie struct {
	UserID string
}

func (c Cookie) createCookie(w http.ResponseWriter) error {
	errLong := errors.New("cookie value too long")

	cookie := http.Cookie{
		Name:     "Cookie",
		Value:    c.UserID,
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
