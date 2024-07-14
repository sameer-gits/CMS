package main

import (
	"net/http"
	"path/filepath"
	"text/template"
)

func renderHtml(w http.ResponseWriter, data interface{}, errs []error, htmlFilename ...string) {
	var htmlFilenames []string
	for _, n := range htmlFilename {
		htmlFilenames = append(htmlFilenames, filepath.Join("../frontend/views/", n))
	}

	tmpl, err := template.ParseFiles(htmlFilenames...)
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	templateData := struct {
		Data   interface{}
		Errors []error
	}{
		Data:   data,
		Errors: errs,
	}

	err = tmpl.Execute(w, templateData)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
