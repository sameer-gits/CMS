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

const contextKey CtxKey = "key"
const views = "views/"

func routes() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/", middleware(homepage))
	mux.HandleFunc("GET /login", loginPage)

	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		formType := r.FormValue("form_type")
		switch formType {
		case "login":
			err := login(w, r)
			if err != nil {
				http.Error(w, err.Error(), serverCode)
				return
			}
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
	userID := r.Context().Value(contextKey).(Cookie).UserID
	renderTemplData(w, "homepage.html", userID)
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	renderTempl(w, "login.html")
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
