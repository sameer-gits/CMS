package main

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/sameer-gits/CMS/database"
	"golang.org/x/crypto/bcrypt"
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
	mux.HandleFunc("/verify", verifyHandler)
	mux.HandleFunc("GET /404", notFoundHandler)

	mux.HandleFunc("POST /createuser", createUserHandler)
	// mux.HandleFunc("POST /verifyuser", verifyUserHandler)

	log.Println("server running on: http://localhost:" + port)
	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Printf("Server error: %v", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	renderHtml(w, nil, nil, "login.html")
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	renderHtml(w, nil, nil, "verify.html")
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
	var userForm FormUser
	uCtx := context.Background()

	defer func() {
		if len(errs) > 0 {
			renderHtml(w, userForm, errs, "login.html")
		} else if len(errs) == 0 {
			http.Redirect(w, r, "/verify", http.StatusFound)
		}
	}()

	userForm, errs = validateForm(r)
	if errs != nil {
		w.WriteHeader(badCode)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.DefaultCost)
	if err != nil {
		errs = append(errs, errors.New("error hashing password"))
		return
	}

	hashedPassword := base64.URLEncoding.EncodeToString(hash)
	Otpgen := rand.Intn(900000) + 100000

	redisUser := RedisUser{
		Username: userForm.Username,
		Fullname: userForm.Fullname,
		Email:    userForm.Email,
		Otp:      Otpgen,
		Password: hashedPassword,
	}

	tx := database.RedisClient.TxPipeline()

	tx.HSet(uCtx, userForm.Email, redisUser)
	tx.Expire(uCtx, userForm.Email, 2*time.Minute)

	_, err = tx.Exec(uCtx)
	if err != nil {
		database.RedisClient.Del(uCtx, userForm.Email)
		w.WriteHeader(badCode)
		return
	}
}
