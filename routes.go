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
	mux.HandleFunc("/verify", verifyHandler)
	mux.HandleFunc("GET /404", notFoundHandler)

	mux.HandleFunc("POST /createuser", createUserHandler)
	mux.HandleFunc("POST /verifyuser", verifyUserHandler)

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
	var userForm CreateUserForm
	var userEmail string

	defer func() {
		if len(errs) > 0 {
			renderHtml(w, userForm, errs, "login.html")
		} else if len(errs) == 0 {
			http.Redirect(w, r, "/verifyuser", http.StatusFound)
		}
	}()

	userForm, errs = validateForm(r)
	if errs != nil {
		w.WriteHeader(badCode)
		return
	}

	userEmail, errs = userForm.addTmpUser()
	if errs != nil {
		w.WriteHeader(serverCode)
		return
	}

	// send userEmail to /verify endpoint for autofill and otp verification with email autofill with w
}

func verifyUserHandler(email string) {
	var errs []error
	var tmpUser RedisUserTmp
	var userID string

	// defer func() {
	// 	if len(errs) > 0 {
	// 		renderHtml(w, nil, errs, "verify.html")
	// 	} else if len(errs) == 0 {
	// 		http.Redirect(w, r, "/", http.StatusFound)
	// 	}
	// }()

	// give 5 chances with rate limit by incrementing chances by 1 everytime
	tmpUser, errs := verifyTmpUser(email)
	if errs != nil {
		w.WriteHeader(badCode)
		return
	}

	userID, errs = tmpUser.createUser()
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
