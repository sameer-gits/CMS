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
	UserID       string `form:"user_id" json:"user_id"`
	Username     string `form:"username" json:"username"`
	Fullname     string `form:"fullname" json:"fullname"`
	Role         string `form:"role" json:"role"`
	Email        string `form:"email" json:"email"`
	Password     string `form:"password" json:"password"`
	ProfileImage []byte `form:"profile_image,omitempty" json:"profile_image,omitempty"`
}

type CtxKey string

const contextKey CtxKey = "userID"
const views = "views/"

func routes() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/", middleware(homepage))
	mux.HandleFunc("/login", loginPage)
	mux.HandleFunc("/logout", deleteCookie)

	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		formType := r.FormValue("form_type")
		switch formType {
		case "register":
			register(w, r)
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		case "login":
			login(w, r)
			http.Redirect(w, r, "/", http.StatusMovedPermanently)
		default:
			http.Error(w, "Invalid form type", badCode)
			return
		}
	})

	log.Println("server running on: http://localhost:" + port)

	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Fatal("Server error: ", err)
	}
}

func homepage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	userID := r.Context().Value(contextKey).(string)
	renderTemplData(w, "homepage.html", userID)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	renderTempl(w, "login.html")
}

func middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getCookie(r)
		if err != nil {
			http.Error(w, err.Error(), unauthorized)
		}

		ctx := context.WithValue(r.Context(), contextKey, userID)
		fmt.Println("i'm middleware")
		next(w, r.WithContext(ctx))
	}
}

func renderTempl(w http.ResponseWriter, name string) {
	tmpl, err := template.ParseFiles(views + name)
	if err != nil {
		http.Error(w, "Error parsing template", serverCode)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template", serverCode)
		return
	}
}

func renderTemplData(w http.ResponseWriter, name string, userID string) {
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
		Username     string
		Fullname     string
		Role         string
		Email        string
		ProfileImage []byte
	}{
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
