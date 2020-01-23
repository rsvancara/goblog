package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"blog/blog"
	"blog/blog/config"
	"blog/blog/views"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {

	fmt.Println("== Starting Service ==")

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println("== Initializing Configuration ==")
	fmt.Printf("Database URI: %s\n", cfg.Dburi)
	fmt.Printf("Cache URI: %s\n", cfg.Cacheuri)

	r := mux.NewRouter()

	r.HandleFunc("/", blog.HomeHandler)
	r.Handle("/", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(blog.HomeHandler))).Methods("GET")
	r.Handle("/stories/{id}", handlers.LoggingHandler(os.Stdout, blog.GeoFilterMiddleware(http.HandlerFunc(blog.PostView))))
	r.Handle("/photo/{id}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(views.PhotoView))).Methods("GET")
	r.Handle("/about", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.AboutHandler)))).Methods("GET")
	r.Handle("/signin", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(blog.Signin))).Methods("GET", "POST")
	r.Handle("/admin", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.AdminHome)))).Methods("GET", "POST")
	r.Handle("/admin/media", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.Media)))).Methods("GET")
	r.Handle("/api/v1/putmedia/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.PutMedia))))
	r.Handle("/api/v1/getmedia/{id}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(views.GetMediaAPI))).Methods("GET")
	r.Handle("/admin/media/add", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.MediaAdd)))).Methods("GET", "POST")
	r.Handle("/admin/media/edit/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.MediaEdit)))).Methods("GET", "POST")
	r.Handle("/admin/media/view/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.ViewMedia)))).Methods("GET")
	r.Handle("/admin/media/delete/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.MediaDelete)))).Methods("GET")
	r.Handle("/admin/post", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.Post)))).Methods("GET")
	r.Handle("/admin/post/add", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.PostAdd)))).Methods("GET", "POST")
	r.Handle("/admin/post/view/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.PostAdminView)))).Methods("GET")
	r.Handle("/admin/post/edit/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.PostEdit)))).Methods("GET", "POST")
	r.Handle("/admin/post/delete/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(blog.PostDelete)))).Methods("GET")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/", r)

	fmt.Println("Now serving requests")

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
