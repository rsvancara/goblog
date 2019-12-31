package views

import (
	"blog/blog/session"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

type jsonErrorMessage struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Media media
func Media(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s\n", err)
	}

	template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// MediaAdd add media
func MediaAdd(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template := "templates/admin/mediaadd.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// PutMedia Upload file to server
func PutMedia(w http.ResponseWriter, r *http.Request) {

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s\"}\n"

	vars := mux.Vars(r)

	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}

	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}
	defer file.Close() // Close the file when we finish

	// This is path which we want to store the file
	f, err := os.OpenFile("temp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err)
		return
	}

	// Copy the file to the destination path
	io.Copy(f, file)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"file %s uploaded\"}\n", vars["id"])

	return
}

func jsonError(err error, w http.ResponseWriter) {

	var jerror jsonErrorMessage

	jerror.Message = err.Error()
	jerror.Status = "error"

	byteError, err := json.Marshal(jerror)
	if err != nil {
		fmt.Printf("Could not marshal error into json string with error %s\n", err)
	}

	errorString := string(byteError)

	w.WriteHeader(http.StatusOK)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, errorString)

	return
}
