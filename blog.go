package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"bf.go/blog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/", blog.HomeHandler)
	r.Handle("/", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(blog.HomeHandler)))
	r.Handle("/about", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.AboutHandler)))).Methods("GET")
	r.Handle("/signin", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(blog.Signin))).Methods("GET", "POST")
	r.Handle("/admin", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.AdminHome)))).Methods("GET", "POST")
	r.Handle("/admin/media", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.Media)))).Methods("GET")
	r.Handle("/admin/media/add", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.MediaAdd)))).Methods("GET")
	r.Handle("/admin/post", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.Post)))).Methods("GET")
	r.Handle("/admin/post/add", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.PostAdd)))).Methods("GET", "POST")
	r.Handle("/admin/post/edit/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.PostEdit)))).Methods("GET", "POST")
	r.Handle("/admin/post/delete/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.PostDelete)))).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/", r)

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:5000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	//log.Fatal(http.ListenAndServe("0.0.0.0:5000", n))

	log.Fatal(srv.ListenAndServe())
}
