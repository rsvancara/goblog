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

	staticAssets, err := views.SiteTemplate("/static")
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(staticAssets)

	fmt.Println("== Initializing Configuration ==")
	fmt.Printf("Database URI: %s\n", cfg.Dburi)
	fmt.Printf("Cache URI: %s\n", cfg.Cacheuri)

	r := mux.NewRouter()

	r.HandleFunc("/", blog.HomeHandler)
	r.Handle(
		"/",
		handlers.LoggingHandler(
			os.Stdout,
			http.HandlerFunc(
				blog.HomeHandler))).Methods("GET")

	r.Handle("/stories/{id}", handlers.LoggingHandler(os.Stdout, blog.GeoFilterMiddleware(http.HandlerFunc(views.PostView))))
	r.Handle("/photo/{id}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(views.PhotoView))).Methods("GET")
	r.Handle("/image/{slug}/{type}", handlers.LoggingHandler(os.Stdout, http.HandlerFunc(views.ServerImage))).Methods("GET")
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
	r.Handle("/admin/post", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.Post)))).Methods("GET")
	r.Handle("/admin/post/add", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.PostAdd)))).Methods("GET", "POST")
	r.Handle("/admin/post/view/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.PostAdminView)))).Methods("GET")
	r.Handle("/admin/post/edit/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.PostEdit)))).Methods("GET", "POST")
	r.Handle("/admin/post/delete/{id}", handlers.LoggingHandler(os.Stdout, blog.AuthHandler(http.HandlerFunc(views.PostDelete)))).Methods("GET")
	ServeStatic(r, "./"+staticAssets)
	http.Handle("/", r)

	fmt.Println("Now serving requests")

	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:5000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

// ServeStatic  serve static content from the appropriate location
func ServeStatic(router *mux.Router, staticDirectory string) {
	staticPaths := map[string]string{
		"css":     staticDirectory + "/css/",
		"images":  staticDirectory + "/images/",
		"scripts": staticDirectory + "/scripts/",
	}
	for pathName, pathValue := range staticPaths {
		pathPrefix := "/" + pathName + "/"
		router.PathPrefix(pathPrefix).Handler(http.StripPrefix(pathPrefix,
			http.FileServer(http.Dir(pathValue))))
	}
}
