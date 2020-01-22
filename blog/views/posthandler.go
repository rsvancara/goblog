package views

import (
	"blog/blog/models"
	"blog/blog/session"
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

// PhotoView View File
func PhotoView(w http.ResponseWriter, r *http.Request) {
	var media models.MediaModel

	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	// Load Media
	err = media.GetMediaBySlug(vars["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	template := "templates/admin/mediaview.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":           "View Media",
		"media":           media,
		"user":            sess.User.Username,
		"bodyclass":       "",
		"fluid":           true,
		"hidetitle":       true,
		"exposureprogram": media.GetExposureProgramTranslated(),
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// GetMediaAPI View File
func GetMediaAPI(w http.ResponseWriter, r *http.Request) {

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\",\"image\":\"error\",\"url\":\"/static/no-image.svg\",\"refurl\":\"#\"}\n"

	//http Session
	var sess session.Session
	sess.Session(r)
	//if err != nil {
	//fmt.Printf("Session not available %s\n", err)
	//}

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if _, ok := vars["id"]; ok {

	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, "Error find ID", "ID was not available inthe URL or could not be parsed")
		return
	}

	// Media Object
	var media models.MediaModel

	err := media.GetMedia(vars["id"])
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "could not get media object from database")
		return
	}

	s3URL := "https://vi-goblog.s3-us-west-2.amazonaws.com" + media.S3LargeView
	refURL := "/photo/" + media.Slug

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"media found\",\"url\":\"%s\",\"refurl\":\"%s\"}\n", s3URL, refURL)

	return
}
