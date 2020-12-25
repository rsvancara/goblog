package routes

import (
	"log"
	"net/http"
	"os"

	bloghandlers "github.com/rsvancara/goblog/internal/handlers"
	"github.com/rsvancara/goblog/internal/util"
	"github.com/rsvancara/goblog/internal/views"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

//GetRoutes get the routes for the application
func GetRoutes(hctx *bloghandlers.HTTPHandlerContext) *mux.Router {

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

	// Category
	r.Handle(
		"/category/{category}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.ViewCategoryHandler))))).Methods("GET")

	// Category
	r.Handle(
		"/categories",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(views.ViewCategoriesHandler))))).Methods("GET")

	// Photo
	r.Handle(
		"/photo/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(hctx.PhotoViewHandler))))).Methods("GET")

	// Image
	r.Handle(
		"/image/{slug}/{type}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(hctx.ServerImageHandler))))).Methods("GET")

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
					views.AuthHandler(http.HandlerFunc(hctx.MediaHandler)))))).Methods("GET")

	// Media View
	r.Handle(
		"/admin/medialist",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(hctx.MediaListViewHandler)))))).Methods("GET")

	// Media Interface
	r.Handle(
		"/api/v1/putmedia/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(hctx.PutMediaAPI))))))

	// Media Interface
	r.Handle(
		"/api/v1/getmedia/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					http.HandlerFunc(hctx.GetMediaAPI))))).Methods("GET")

	// Media Interface
	r.Handle(
		"/api/v1/searchmedia",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(hctx.MediaSearchAPIHandler))))))

	// Media Interface
	r.Handle(
		"/api/v1/editmedia",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(hctx.EditMediaAPIHandler))))))

	// Add media
	r.Handle(
		"/admin/media/add",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(http.HandlerFunc(hctx.MediaAddHandler)))))).Methods("GET", "POST")

	// Edit Media
	r.Handle(
		"/admin/media/edit/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(hctx.MediaEditHandler)))))).Methods("GET", "POST")

	// Admin View Media
	r.Handle(
		"/admin/media/view/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(hctx.ViewMediaHandler)))))).Methods("GET")

	// Delete media page
	r.Handle(
		"/admin/media/delete/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(hctx.MediaDeleteHandler)))))).Methods("GET")

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
		"/admin/session/details/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.SessionDetailsReportHandler)))))).Methods("GET")

	r.Handle(
		"/admin/session/inspector/{id}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.RequestInspectorReportHandler)))))).Methods("GET")

	r.Handle(
		"/admin/filters",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.FilterHandler)))))).Methods("GET")

	r.Handle(
		"/admin/filters/create",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.CreateFilterHandler)))))).Methods("GET")

	r.Handle(
		"/api/v1/filters/create",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.CreateAPIFilterHandler)))))).Methods("GET")

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
		"/admin/searchindex",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.SearchIndexListHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/buildmediaindex",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.SearchIndexBuildTagsHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/buildmediaindexgo",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.SearchIndexBuilderMediaHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/api/search-media-tags-by-name/{name}",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				views.GeoFilterMiddleware(
					views.AuthHandler(
						http.HandlerFunc(views.SearchIndexMediaTagsAPI)))))).Methods("GET", "POST")

	r.Handle(
		"/api/request/v1",
		handlers.LoggingHandler(
			os.Stdout,
			views.SessionHandler(
				http.HandlerFunc(views.RequestBotAPI)))).Methods("GET", "POST")

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
