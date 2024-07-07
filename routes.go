package main

import (
	"log"
	"net/http"
	"os"
)

const (
	serverCode   = http.StatusInternalServerError
	unauthorized = http.StatusUnauthorized
	statusOK     = http.StatusOK
	badCode      = http.StatusBadRequest
)

func routes() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("GET /404", notFoundHandler)
	mux.HandleFunc("POST /createuser", createUserHandler)

	log.Println("server running on: http://localhost:" + port)
	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Printf("Server error: %v", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	renderHtml(w, nil, nil, "login.html")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	renderHtml(w, nil, nil, "notFound.html")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" || r.Method != "GET" {
		notFoundHandler(w, r)
		return
	}
	renderHtml(w, nil, nil, "index.html")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var errs []error
	var userForm CreateUserForm
	var userID string

	defer func() {
		if len(errs) > 0 {
			renderHtml(w, userForm, errs, "login.html")
		} else if len(errs) == 0 {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}()

	userForm, errs = validateForm(r)
	if errs != nil {
		w.WriteHeader(badCode)
		return
	}

	userID, errs = userForm.createUser()
	if errs != nil {
		w.WriteHeader(serverCode)
		return
	}

	myCookie := Cookie{
		UserID: userID,
	}

	errs = myCookie.createCookie(w)
	if errs != nil {
		w.WriteHeader(serverCode)
		return
	}
}
