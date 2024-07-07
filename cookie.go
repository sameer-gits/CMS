package main

import (
	"errors"
	"net/http"
	"time"
)

type Cookie struct {
	UserID string
}

func (c Cookie) createCookie(w http.ResponseWriter) []error {
	var errs []error
	err := errors.New("cookie value too long")

	cookie := http.Cookie{
		Name:     "Cookie",
		Value:    c.UserID,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24 * 30),
		MaxAge:   3600 * 24 * 30,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	if len(cookie.String()) > 4096 {
		errs = append(errs, err)
		return errs
	}

	http.SetCookie(w, &cookie)
	return nil
}
