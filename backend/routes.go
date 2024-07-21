package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
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
	mux.HandleFunc("/logout", deleteCookieHandler)
	mux.HandleFunc("/protected", protectedHandler)
	mux.HandleFunc("/404", notFoundHandler)
	mux.HandleFunc("/verify", redirectLoginHandler)
	mux.HandleFunc("/resendotp", redirectLoginHandler)

	mux.HandleFunc("POST /login", createUserHandler)
	mux.HandleFunc("POST /verify", verifyUserHandler)
	mux.HandleFunc("POST /resendotp", resendOtpHandler)

	localFlies := http.FileServer(http.Dir("../frontend/public"))
	mux.Handle("/public/", http.StripPrefix("/public/", localFlies))

	log.Println("server running on: http://localhost:" + port)
	if err := http.ListenAndServe("0.0.0.0:"+port, mux); err != nil {
		log.Printf("Server error: %v", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var formUser FormUser
	renderHtml(w, formUser, nil, "login.html")
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	user, err := userInfoMiddleware(r)
	if err != nil {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}
	renderHtml(w, user, nil, "protected.html")
}

func redirectLoginHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
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
	var redisUser RedisUser
	ctx := context.Background()

	defer func() {
		if len(errs) > 0 {
			w.WriteHeader(badCode)
			renderHtml(w, userForm, errs, "login.html")
		} else if len(errs) == 0 {
			renderHtml(w, userForm, errs, "verify.html")
		}
	}()

	userForm, errs = validateForm(r)
	if errs != nil {
		return
	}

	err := database.RedisAllClients.Client0.HMGet(ctx, userForm.Email, userForm.Email, "request").Scan(&redisUser)
	if err != nil {
		errs = append(errs, errors.New("error creating user try again"))
		return
	}

	if redisUser.Blocked == "true" {
		errs = append(errs, errors.New("max attempts already exceeded: user blocked for 1 day"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.DefaultCost)
	if err != nil {
		errs = append(errs, errors.New("error hashing password"))
		return
	}

	rd := rand.New(rand.NewSource(time.Now().Unix()))
	otpGen := rd.Intn(899999) + 100000
	request := 0
	blocked := "false"
	hashedPassword := base64.URLEncoding.EncodeToString(hash)

	redisUser = RedisUser{
		Username: userForm.Username,
		Fullname: userForm.Fullname,
		Email:    userForm.Email,
		Otp:      otpGen,
		Request:  request,
		Blocked:  blocked,
		Password: hashedPassword,
	}

	tx := database.RedisAllClients.Client0.TxPipeline()

	tx.HSet(ctx, redisUser.Email, redisUser).Err()
	tx.Expire(ctx, redisUser.Email, 2*time.Minute).Err()

	_, err = tx.Exec(ctx)
	if err != nil {
		database.RedisAllClients.Client0.Del(ctx, redisUser.Email)
		return
	}

	headers := "From: " + os.Getenv("EMAIL_HEADER") + "\r\n" +
		"To: " + redisUser.Email + "\r\n" +
		"Subject: Your One-Time Password\r\n" +
		"\r\n"

	body := fmt.Sprintf("Hello your One-Time Password is %d, Valid for 2mins.", redisUser.Otp)

	message := headers + body

	// send OTP to user here
	sendMailTo := mailTo{
		from:        os.Getenv("EMAIL_SMTP"),
		password:    os.Getenv("EMAIL_SMTP_PASSWORD"),
		sendTo:      []string{redisUser.Email},
		smtpHost:    "smtp-relay.brevo.com",
		smtpPort:    "587",
		mailMessage: message,
	}

	err = sendMailTo.sendMail()
	if err != nil {
		errs = append(errs, fmt.Errorf("error sending OTP: %v", err))
	}
}

func verifyUserHandler(w http.ResponseWriter, r *http.Request) {
	var errs []error
	var redisUser RedisUser
	var formUser FormUser

	formUser.Email = r.FormValue("email")
	userOtp := r.FormValue("otp")

	defer func() {
		if len(errs) > 0 {
			renderHtml(w, formUser, errs, "verify.html")
		} else if len(errs) == 0 {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	}()

	ctx := context.Background()
	err := database.RedisAllClients.Client0.HGetAll(ctx, formUser.Email).Scan(&redisUser)
	if err != nil {
		errs = append(errs, errors.New("error matching OTP, try again"))
		return
	}

	if redisUser.Email == "" {
		formUser.Email = ""
		errs = append(errs, errors.New("error no email is register maybe timeout try registering again"))
		return
	}

	if redisUser.Request == 5 {
		tx := database.RedisAllClients.Client0.TxPipeline()

		tx.HSet(ctx, redisUser.Email, "blocked", "true").Err()
		tx.Expire(ctx, redisUser.Email, 24*time.Hour).Err()

		_, err = tx.Exec(ctx)
		if err != nil {
			return
		}

		errs = append(errs, errors.New("max attempts already exceeded: user blocked for 1 day"))
		return
	}

	redisUser.Request++

	if userOtp != strconv.Itoa(redisUser.Otp) {
		err = database.RedisAllClients.Client0.HSet(ctx, redisUser.Email, "request", redisUser.Request).Err()
		if err != nil {
			errs = append(errs, errors.New("wrong OTP, try again"))
			return
		}
		errs = append(errs, errors.New("wrong OTP, try again"))
		return
	}

	userID, errs := redisUser.createUser()
	if errs != nil {
		return
	}

	err = database.RedisAllClients.Client0.Del(ctx, redisUser.Email).Err()
	if err != nil {
		errs = append(errs, errors.New("error removing temporary data"))
	}

	c := Cookie{
		UserID: userID,
	}

	c.createCookie(w)
	if errs != nil {
		errs = append(errs, errors.New("error creating cookie try logging in or try again creating account"))
		return
	}
}

func resendOtpHandler(w http.ResponseWriter, r *http.Request) {
	var errs []error
	var redisUser RedisUser
	var formUser FormUser

	formUser.Email = r.FormValue("email")

	defer func() {
		if len(errs) > 0 {
			renderHtml(w, formUser, errs, "verify.html")
		} else if len(errs) == 0 {
			formUser.Message = "New OTP sent, Please check your email."
			renderHtml(w, formUser, errs, "verify.html")
		}
	}()

	ctx := context.Background()
	err := database.RedisAllClients.Client0.HGetAll(ctx, formUser.Email).Scan(&redisUser)
	if err != nil {
		errs = append(errs, errors.New("error sending new OTP, try again"))
		return
	}

	if redisUser.Request == 5 {
		tx := database.RedisAllClients.Client0.TxPipeline()

		tx.HSet(ctx, redisUser.Email, "blocked", "true").Err()
		tx.Expire(ctx, redisUser.Email, 24*time.Hour).Err()

		_, err = tx.Exec(ctx)
		if err != nil {
			errs = append(errs, errors.New("something went wrong try again"))
			return
		}

		errs = append(errs, errors.New("max attempts already exceeded: user blocked for 1 day"))
		return
	}

	redisUser.Request++
	rd := rand.New(rand.NewSource(time.Now().Unix()))
	otpGen := rd.Intn(899999) + 100000

	tx := database.RedisAllClients.Client0.TxPipeline()
	tx.HSet(ctx, redisUser.Email, "otp", otpGen).Err()
	tx.HSet(ctx, redisUser.Email, "request", redisUser.Request).Err()
	tx.Expire(ctx, redisUser.Email, 2*time.Minute).Err()
	_, err = tx.Exec(ctx)
	if err != nil {
		return
	}

	if redisUser.Email == "" {
		formUser.Email = ""
		errs = append(errs, errors.New("error no email is register maybe timeout try registering again"))
		return
	}

	headers := "From: " + os.Getenv("EMAIL_HEADER") + "\r\n" +
		"To: " + redisUser.Email + "\r\n" +
		"Subject: Your New One-Time Password\r\n" +
		"\r\n"

	body := fmt.Sprintf("Hello your New One-Time Password is %d, Valid for 2mins.", otpGen)

	message := headers + body

	// send OTP to user here
	sendMailTo := mailTo{
		from:        os.Getenv("EMAIL_SMTP"),
		password:    os.Getenv("EMAIL_SMTP_PASSWORD"),
		sendTo:      []string{redisUser.Email},
		smtpHost:    "smtp-relay.brevo.com",
		smtpPort:    "587",
		mailMessage: message,
	}

	err = sendMailTo.sendMail()
	if err != nil {
		fmt.Println(redisUser.Email)
		errs = append(errs, fmt.Errorf("error sending new OTP: %v", err))
	}

}
