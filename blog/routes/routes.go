package routes

import (
	"blog/blog/util"
	"blog/blog/views"
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

	r := mux.NewRouter()

	// Index Page
	r.Handle(
		"/",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(
						views.HomeHandler))))).Methods("GET")

	// Test Page
	r.Handle(
		"/healthcheck957873",
		handlers.LoggingHandler(
			os.Stdout,
			http.HandlerFunc(
				views.HealthCheck))).Methods("GET")

	// Sitemap
	r.Handle(
		"/sitemap.xml",
		handlers.LoggingHandler(
			os.Stdout,
			http.HandlerFunc(
				views.SiteMap))).Methods("GET")

	// Stories page
	r.Handle(
		"/stories/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.PostView))))).Methods(("GET"))

	// Photo
	r.Handle(
		"/photo/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.PhotoView))))).Methods("GET")

	// Image
	r.Handle(
		"/image/{slug}/{type}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.ServerImage))))).Methods("GET")

	// About Page
	r.Handle(
		"/about",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.AboutHandler))))).Methods("GET")

	// Signin
	r.Handle(
		"/signin",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.Signin))))).Methods("GET", "POST")

	// Admin main page
	r.Handle(
		"/admin",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.AdminHome)))))).Methods("GET", "POST")

	// Media View
	r.Handle(
		"/admin/media",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(views.Media)))))).Methods("GET")

	// Media Interface
	r.Handle(
		"/api/v1/putmedia/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(views.PutMedia))))))

	// Media Interface
	r.Handle(
		"/api/v1/getmedia/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.GetMediaAPI))))).Methods("GET")

	// Add media
	r.Handle(
		"/admin/media/add",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(views.MediaAdd)))))).Methods("GET", "POST")

	// Edit Media
	r.Handle(
		"/admin/media/edit/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.MediaEdit)))))).Methods("GET", "POST")

	// Admin View Media
	r.Handle(
		"/admin/media/view/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.ViewMedia)))))).Methods("GET")

	// Delete media page
	r.Handle(
		"/admin/media/delete/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.MediaDelete)))))).Methods("GET")

	r.Handle(
		"/admin/post",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.Post)))))).Methods("GET")

	r.Handle(
		"/admin/post/add",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.PostAdd)))))).Methods("GET", "POST")
	r.Handle(
		"/admin/post/view/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.PostAdminView)))))).Methods("GET")
	r.Handle(
		"/admin/post/edit/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(views.PostEdit)))))).Methods("GET", "POST")
	r.Handle(
		"/admin/post/delete/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.PostDelete)))))).Methods("GET")
	r.Handle(
		"/admin/sessions",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.SessionReportHandler)))))).Methods("GET")

	r.Handle(
		"/admin/filters",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.FilterHandler)))))).Methods("GET")

	r.Handle(
		"/admin/affiliates",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.AffiliateHandler)))))).Methods("GET")

	r.Handle(
		"/admin/affiliates/add",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.AffiliateAddHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/affiliates/edit/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.AffiliateEditHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/affiliates/delete/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.AffiliateDeleteHandler)))))).Methods("GET")

	r.Handle(
		"/bouncyhouse/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.AffiliateBouncyHouseHandler)))))).Methods("GET")

	r.Handle(
		"/contact",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.ContactHandler))))).Methods("GET")

	r.Handle(
		"/admin/session/delete/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.SessionDeleteHandler)))))).Methods("GET")

	// Fake wordpress routes for detecting and blocking bad bots
	r.Handle(
		"/wp-login.php",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.WPLoginHandler))))).Methods("GET", "POST")

	r.Handle(
		"/wp-admin",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.WPAdminHandler))))).Methods("GET", "POST")

	r.Handle(
		"/request/api/v1",
		handlers.LoggingHandler(
			os.Stdout,
			http.HandlerFunc(views.RequestBotAPI))).Methods("GET", "POST")

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
