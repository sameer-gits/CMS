package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

type User struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Fullname     string `json:"fullname"`
	Role         string `json:"role"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	ProfileImage []byte `json:"profile_image"`
}

type Cookie struct {
	UserID string
}

type CtxKey string

const serverCode = http.StatusInternalServerError
const contextKey CtxKey = "key"
const views = "views/"

func routes() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/", middleware(homepage))

	log.Println("server running on: http://localhost:" + port)

	if err := http.ListenAndServe("0.0.0.0:"+port, nil); err != nil {
		log.Fatal("Server error: ", err)
	}
}

func homepage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(contextKey).(Cookie).UserID
	renderTempl(w, "homepage.html", userID)
}

func middleware(next http.HandlerFunc) http.HandlerFunc {
	userID := "1aedddb7-259e-48f3-ad45-82436df3b074"
	// userID fetch here using JWT
	return func(w http.ResponseWriter, r *http.Request) {

		cookie := Cookie{
			UserID: userID,
		}

		ctx := context.WithValue(r.Context(), contextKey, cookie)
		fmt.Println("i'm middleware")
		next(w, r.WithContext(ctx))
	}
}

func renderTempl(w http.ResponseWriter, name string, userID string) {
	user, err := selectUser(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching user details: %v", err), serverCode)
		return
	}

	tmpl, err := template.ParseFiles(views + name)
	if err != nil {
		http.Error(w, "Error parsing template", serverCode)
		return
	}

	data := struct {
		UserID       string
		Username     string
		Fullname     string
		Role         string
		Email        string
		ProfileImage []byte
	}{
		UserID:       userID,
		Username:     user.Username,
		Fullname:     user.Fullname,
		Role:         user.Role,
		Email:        user.Email,
		ProfileImage: user.ProfileImage,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", serverCode)
		return
	}
}
