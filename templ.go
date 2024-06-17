package main

import (
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
)

// render template without data
func renderTempl(w http.ResponseWriter, name ...string) {
	var templateFiles []string
	for _, n := range name {
		templateFiles = append(templateFiles, filepath.Join(views, n))
	}

	tmpl, err := template.ParseFiles(templateFiles...)
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
func renderTemplData(w http.ResponseWriter, r *http.Request, name ...string) {
	userID, err := getCookie(r)
	if err != nil {
		http.Error(w, err.Error(), unauthorized)
		return
	}

	user, err := selectUser(userID)
	if err != nil {
		// there is problem not sure!
		fmt.Println("Hello")
		// http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}
	forums, err := selectForums()
	if err != nil {
		// there is problem not sure!
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}
	var templateFiles []string
	for _, n := range name {
		templateFiles = append(templateFiles, filepath.Join(views, n))
	}

	tmpl, err := template.ParseFiles(templateFiles...)
	if err != nil {
		http.Error(w, "Error parsing template", serverCode)
		return
	}

	data := struct {
		User   User
		Forums []Forum
	}{
		User:   user,
		Forums: forums,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error executing template", serverCode)
		return
	}
}
