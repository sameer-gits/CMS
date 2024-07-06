package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
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

type CreateUserForm struct {
	Username     string `form:"username" json:"username"`
	Fullname     string `form:"fullname" json:"fullname"`
	Email        string `form:"email" json:"email"`
	Password     string `form:"password" json:"password"`
	ProfileImage []byte `form:"profile_image,omitempty" json:"profile_image,omitempty"`
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	renderhtml(w, nil, nil, "index.html")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var errs []error
	var form CreateUserForm

	defer func() {
		if len(errs) > 0 {
			renderhtml(w, form, errs, "index.html")
		}
	}()

	image, _, err := r.FormFile("profile_image")
	if err != nil {
		if err == http.ErrMissingFile {
			form.ProfileImage = nil
			fmt.Println("No Image")
		} else {
			errs = append(errs, err)
			return
		}
	} else {

		image.Close()

		form.ProfileImage, err = io.ReadAll(image)
		if err != nil {
			errs = append(errs, err)
		}

		contentType := http.DetectContentType(form.ProfileImage)
		if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
			errs = append(errs, errors.New("unsupported image type"))
		}
	}

	form = CreateUserForm{
		Username:     r.FormValue("username"),
		Fullname:     r.FormValue("fullname"),
		Email:        r.FormValue("email"),
		Password:     r.FormValue("password"),
		ProfileImage: form.ProfileImage,
	}

	if strings.TrimSpace(form.Username) == "" {
		err = errors.New("please provide username")
		errs = append(errs, err)
	}

	if strings.TrimSpace(form.Fullname) == "" {
		err = errors.New("please provide fullname")
		errs = append(errs, err)
	}

	emailregex := regexp.MustCompile(`.+@.+\..+`)
	if !emailregex.MatchString(form.Email) {
		err = errors.New("please provide valid email")
		errs = append(errs, err)
	}

	if strings.TrimSpace(form.Password) == "" {
		err = errors.New("please provide password")
		errs = append(errs, err)
		w.WriteHeader(badCode)
		return
	}

	user := User{
		Username:     form.Username,
		Fullname:     form.Fullname,
		Email:        form.Email,
		Password:     form.Password,
		ProfileImage: form.ProfileImage,
	}

	userID, err := user.createUser()
	if err != nil {
		errs = append(errs, err)
		w.WriteHeader(serverCode)
		return
	}

	myCookie := Cookie{
		UserID: userID,
	}

	err = myCookie.createCookie(w)
	if err != nil {
		errs = append(errs, err)
		w.WriteHeader(serverCode)
		return
	}

	if len(errs) > 0 {
		w.WriteHeader(http.StatusCreated)
	}
}
