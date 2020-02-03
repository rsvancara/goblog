package routes

import (
	"blog/blog/util"
	"blog/blog/views"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

//GetRoutes get the routes for the application
func GetRoutes() *mux.Router {

	staticAssets, err := util.SiteTemplate("/static")
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(staticAssets)

	r := mux.NewRouter()

	r.HandleFunc("/", views.HomeHandler)
	r.Handle(
		"/",
		handlers.LoggingHandler(
			os.Stdout,
			http.HandlerFunc(
				views.HomeHandler))).Methods("GET")

	r.Handle("/stories/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(http.HandlerFunc(views.PostView))))
	r.Handle("/photo/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(http.HandlerFunc(views.PhotoView)))).Methods("GET")
	r.Handle("/image/{slug}/{type}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(http.HandlerFunc(views.ServerImage)))).Methods("GET")
	r.Handle("/about", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.AboutHandler))))).Methods("GET")
	r.Handle("/signin", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(http.HandlerFunc(views.Signin)))).Methods("GET", "POST")
	r.Handle("/admin", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.AdminHome))))).Methods("GET", "POST")
	r.Handle("/admin/media", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.Media))))).Methods("GET")
	r.Handle("/api/v1/putmedia/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.PutMedia)))))
	r.Handle("/api/v1/getmedia/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(http.HandlerFunc(views.GetMediaAPI)))).Methods("GET")
	r.Handle("/admin/media/add", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.MediaAdd))))).Methods("GET", "POST")
	r.Handle("/admin/media/edit/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.MediaEdit))))).Methods("GET", "POST")
	r.Handle("/admin/media/view/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.ViewMedia))))).Methods("GET")
	r.Handle("/admin/media/delete/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.MediaDelete))))).Methods("GET")
	r.Handle("/admin/post", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.Post))))).Methods("GET")
	r.Handle("/admin/post/add", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.PostAdd))))).Methods("GET", "POST")
	r.Handle("/admin/post/view/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.PostAdminView))))).Methods("GET")
	r.Handle("/admin/post/edit/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.PostEdit))))).Methods("GET", "POST")
	r.Handle("/admin/post/delete/{id}", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.PostDelete))))).Methods("GET")
	r.Handle("/admin/sessions", handlers.LoggingHandler(os.Stdout, views.GeoFilterMiddleware(views.AuthHandler(http.HandlerFunc(views.SessionReportHandler))))).Methods("GET")
	ServeStatic(r, "./"+staticAssets)
	http.Handle("/", r)

	return r

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
