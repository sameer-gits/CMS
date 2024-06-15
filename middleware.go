package main

import (
	"context"
	"fmt"
	"net/http"
)

func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getCookie(r)
		if err != nil {
			// DO Something and then return
			redirectToLogin(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), contextKey, userID)
		fmt.Println("i'm middleware and getCookie has data is confirmed")
		next(w, r.WithContext(ctx))
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
