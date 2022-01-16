package routes

import (
	"net/http"
	"os"

	bloghandlers "goblog/internal/handlers"
	mw "goblog/internal/middleware"
	"goblog/internal/util"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

//GetRoutes get the routes for the application
func GetRoutes(hctx *bloghandlers.HTTPHandlerContext, mwctx *mw.MiddleWareContext) *mux.Router {

	staticAssets, err := util.SiteTemplate("/static")
	if err != nil {
		log.Fatal().Err(err).Str("service", "wpengine").Msg("Please ensure the web template directory exists and that you have permissions to access it")
	}

	r := mux.NewRouter()
	r.StrictSlash(true)

	// Index Page
	r.Handle(
		"/",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.HomeHandler))))).Methods("GET")

	// Health check page
	r.Handle(
		"/healthcheck957873",
		http.HandlerFunc(hctx.HealthCheckHandler)).Methods("GET")

	// Sitemap used by search engines
	r.Handle(
		"/sitemap.xml",
		mwctx.MLog(
			http.HandlerFunc(hctx.SiteMap))).Methods("GET")

	// View individual story
	r.Handle(
		"/stories/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.PostViewHandler))))).Methods(("GET"))

	// View individual category
	r.Handle(
		"/category/{category}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.ViewCategoryHandler))))).Methods("GET")

	// View a list of categories
	r.Handle(
		"/categories",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.ViewCategoriesHandler))))).Methods("GET")

	// Photo
	r.Handle(
		"/photo/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.PhotoViewHandler))))).Methods("GET")

	// Image
	r.Handle(
		"/image/{slug}/{type}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.ServerImageHandler))))).Methods("GET")

	// About Page
	r.Handle(
		"/about",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.AboutHandler))))).Methods("GET")

	// Signin
	r.Handle(
		"/signin",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.SignInHandler))))).Methods("GET", "POST")

	// Admin main page
	r.Handle(
		"/admin",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.AdminHomeHandler)))))).Methods("GET", "POST")

	// Media View
	r.Handle(
		"/admin/media",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(http.HandlerFunc(hctx.MediaHandler)))))).Methods("GET")

	// Media View
	r.Handle(
		"/admin/medialist",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(http.HandlerFunc(hctx.MediaListViewHandler)))))).Methods("GET")

	// Media Interface
	r.Handle(
		"/api/v1/putmedia/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(http.HandlerFunc(hctx.PutMediaAPIV2))))))

	// Media Interface
	r.Handle(
		"/api/v1/getmedia/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.GetMediaAPI))))).Methods("GET")

	// Media Interface
	r.Handle(
		"/api/v1/searchmedia",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(http.HandlerFunc(hctx.MediaSearchAPIHandler))))))

	// Media Interface
	r.Handle(
		"/api/v1/change-media-title",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(http.HandlerFunc(hctx.MediaUpdateTitleHandler))))))

	// Media Interface
	r.Handle(
		"/api/v1/editmedia",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(http.HandlerFunc(hctx.EditMediaAPIHandler))))))

	// Add media
	r.Handle(
		"/admin/media/add",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(http.HandlerFunc(hctx.MediaAddHandler)))))).Methods("GET", "POST")

	// Edit Media
	r.Handle(
		"/admin/media/edit/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.MediaEditHandler)))))).Methods("GET", "POST")

	// Admin View Media
	r.Handle(
		"/admin/media/view/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.ViewMediaHandler)))))).Methods("GET")

	// Delete media page
	r.Handle(
		"/admin/media/delete/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.MediaDeleteHandler)))))).Methods("GET")

	r.Handle(
		"/admin/post",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.PostHandler)))))).Methods("GET")

	r.Handle(
		"/admin/post/add",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.PostAddHandler)))))).Methods("GET", "POST")
	r.Handle(
		"/admin/post/view/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.PostAdminViewHandler)))))).Methods("GET")

	r.Handle(
		"/admin/post/edit/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(http.HandlerFunc(hctx.PostEditHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/post/delete/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.PostDeleteHandler)))))).Methods("GET")

	r.Handle(
		"/admin/sessions",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.SessionReportHandler)))))).Methods("GET")

	r.Handle(
		"/admin/session/details/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.SessionDetailsReportHandler)))))).Methods("GET")

	r.Handle(
		"/admin/session/inspector/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.RequestInspectorReportHandler)))))).Methods("GET")

	r.Handle(
		"/admin/filters",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.FilterHandler)))))).Methods("GET")

	r.Handle(
		"/admin/filters/create",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.CreateFilterHandler)))))).Methods("GET")

	r.Handle(
		"/api/v1/filters/create",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.CreateAPIFilterHandler)))))).Methods("GET")

	r.Handle(
		"/admin/affiliates",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.AffiliateHandler)))))).Methods("GET")

	r.Handle(
		"/admin/affiliates/add",
		handlers.LoggingHandler(
			os.Stdout,
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.AffiliateAddHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/affiliates/edit/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.AffiliateEditHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/affiliates/delete/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.AffiliateDeleteHandler)))))).Methods("GET")

	r.Handle(
		"/bouncyhouse/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.AffiliateBouncyHouseHandler)))))).Methods("GET")

	r.Handle(
		"/contact",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					http.HandlerFunc(hctx.ContactHandler))))).Methods("GET")

	r.Handle(
		"/admin/session/delete/{id}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.SessionDeleteHandler)))))).Methods("GET")

	r.Handle(
		"/admin/searchindex",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.SearchIndexListHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/buildmediaindex",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.SearchIndexBuildTagsHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/buildmediaindexgo",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.SearchIndexBuilderMediaHandler)))))).Methods("GET", "POST")

	r.Handle(
		"/admin/api/search-media-tags-by-name/{name}",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				mwctx.GeoFilterMiddleware(
					mwctx.AuthHandlerMiddleware(
						http.HandlerFunc(hctx.SearchIndexMediaTagsAPI)))))).Methods("GET", "POST")

	r.Handle(
		"/api/request/v1",
		mwctx.MLog(
			mwctx.SessionMiddleware(
				http.HandlerFunc(hctx.RequestBotAPI)))).Methods("GET", "POST")

	ServeStatic(r, "./"+staticAssets)

	r.NotFoundHandler = mwctx.SessionMiddleware(
		mwctx.GeoFilterMiddleware(
			http.HandlerFunc(hctx.NotFoundHandler)))

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
