package main

import (
	"html/template"
	"net/http"
)

// render template without data
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

// render template with data
func renderTemplData(w http.ResponseWriter, r *http.Request, name string, userID string) {
	user, err := selectUser(userID)
	if err != nil {
		http.Redirect(w, r, "/logout", http.StatusFound)
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
