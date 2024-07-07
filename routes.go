package main

import (
	"fmt"
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

	mux.HandleFunc("POST /createuser", createUserHandler)
	mux.HandleFunc("/", homeHandler)

	log.Println("server running on: http://localhost:" + port)
	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Printf("Server error: %v", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderHtml(w, nil, nil, "index.html")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var errs []error
	var userForm CreateUserForm

	defer func() {
		if len(errs) > 0 {
			renderHtml(w, userForm, errs, "index.html")
		}
	}()

	userForm, errs = validateForm(r)
	if len(errs) > 0 {
		w.WriteHeader(badCode)
		return
	}

	fmt.Println(userForm)

	userID, errs := userForm.createUser()
	if errs != nil {
		w.WriteHeader(serverCode)
		return
	}

	myCookie := Cookie{
		UserID: userID,
	}

	err := myCookie.createCookie(w)
	if err != nil {
		errs = append(errs, err)
		w.WriteHeader(serverCode)
		return
	}

	if len(errs) < 1 {
		w.WriteHeader(http.StatusCreated)
	}
}
