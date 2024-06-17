package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type User struct {
	UserID       uuid.UUID `form:"user_id" json:"user_id"`
	Username     string    `form:"username" json:"username"`
	Fullname     string    `form:"fullname" json:"fullname"`
	Role         string    `form:"role" json:"role"`
	Email        string    `form:"email" json:"email"`
	Password     string    `form:"password" json:"password"`
	ProfileImage []byte    `form:"profile_image,omitempty" json:"profile_image,omitempty"`
}

type CtxKey string

const contextKey CtxKey = "userID"
const views = "views/"

func routes() {
	if port == "" {
		port = "8000"
	}
	mux := http.NewServeMux()

	// all routes here
	mux.HandleFunc("/", notFound(middleware(homepage)))
	mux.HandleFunc("/login", loginPage)
	mux.HandleFunc("/forum", middleware(forum))
	mux.HandleFunc("/forumlist", middleware(forumlist))

	mux.HandleFunc("POST /newforum", middleware(func(w http.ResponseWriter, r *http.Request) {
		newForum(w, r)
		http.Redirect(w, r, "/forum", http.StatusFound)
	}))

	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		formType := r.FormValue("form_type")
		switch formType {
		case "register":
			register(w, r)
			http.Redirect(w, r, "/", http.StatusFound)
		case "login":
			login(w, r)
			http.Redirect(w, r, "/", http.StatusFound)
		default:
			http.Error(w, "Invalid form type", badCode)
			return
		}
	})

	//just for deleting cookie
	mux.HandleFunc("/logout", deleteCookie)

	log.Println("server running on: http://localhost:" + port)

	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Printf("Server error: %v", err)
	}
}

// protected routes here
func homepage(w http.ResponseWriter, r *http.Request) {
	renderTemplData(w, r, "homepage.html")
}

func forum(w http.ResponseWriter, r *http.Request) {
	renderTemplData(w, r, "forum.html", "forumlist.html")
}

func forumlist(w http.ResponseWriter, r *http.Request) {
	renderTemplData(w, r, "forumlist.html")
}

// unprotected routes here
func loginPage(w http.ResponseWriter, r *http.Request) {
	renderTempl(w, "login.html")
}

// not Found Route
func notFound(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			fmt.Printf("not found route: %v\n", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		next(w, r)
	}
}
