//handlers provides misceleneous handlers for view layer
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/rs/zerolog/log"
	"github.com/rsvancara/goblog/internal/config"
	mediadao "github.com/rsvancara/goblog/internal/dao/media"
	postsdao "github.com/rsvancara/goblog/internal/dao/posts"
	requestviewdao "github.com/rsvancara/goblog/internal/dao/requestview"
	"github.com/rsvancara/goblog/internal/models"
	"github.com/rsvancara/goblog/internal/session"
	"github.com/rsvancara/goblog/internal/util"
)

// Signin Sign into the application
func (ctx *HTTPHandlerContext) SignInHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {

			log.Error().Err(err).Str("service", "authentication").Msg("Error parsing sign in form values")
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		var creds session.Credentials
		creds.Username = r.FormValue("username")
		creds.Password = r.FormValue("password")

		if creds.Username == "" || creds.Password == "" {
			log.Info().Str("service", "authentication").Msg("Error parsing sign in form values, blank values provided")
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		isAuth, err := sess.Authenticate(creds, r, w)
		if err != nil {
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Error().Err(err).Str("service", "authentication").Msgf("Error authenticating user %s", creds.Username)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if isAuth == false {
			log.Info().Msgf("SIGNUP - User is authenticated %s %s\n", sess.SessionToken, sess.User.Username)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	template, err := util.SiteTemplate("/signin.html")
	//template := "templates/signin.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":    "Index",
		"greating": "Hello",
		"user":     sess.User,
		"pagekey":  util.GetPageID(r),
		"token":    sess.SessionToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error loading template with error: %s\n", err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// AdminHomeHandler admin home page
func (ctx *HTTPHandlerContext) AdminHomeHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	geoIP, err := util.GeoIPContext(r)
	if err != nil {
		fmt.Printf("error obtaining geoip context: %s", err)
	}

	template, err := util.SiteTemplate("/admin/admin.html")
	//template := "templates/admin/admin.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":    "Index",
		"greating": "Hello",
		"user":     sess.User,
		"pagekey":  geoIP.PageID,
		"token":    sess.SessionToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// AboutHandler about page
func (ctx *HTTPHandlerContext) AboutHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	template, err := util.SiteTemplate("/about.html")
	//template := "templates/about.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":    "Index",
		"greating": "Hello",
		"user":     sess.User,
		"pagekey":  util.GetPageID(r),
		"token":    sess.SessionToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// HealthCheck defines a healthcheck
func (ctx *HTTPHandlerContext) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "healthy")
}

// ContactHandler defines a healthcheck
func (ctx *HTTPHandlerContext) ContactHandler(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template, err := util.SiteTemplate("/contact.html")
	//template := "templates/about.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":    "Index",
		"greating": "Hello",
		"user":     sess.User,
		"pagekey":  util.GetPageID(r),
		"token":    sess.SessionToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// SiteMap generate a sitemap.xml
func (ctx *HTTPHandlerContext) SiteMap(w http.ResponseWriter, r *http.Request) {

	cfg, err := config.GetConfig()
	if err != nil {
		log.Error().Err(err).Str("service", "authentication").Msg("could not get configuration object %s")
		return
	}
	var postDAO postsdao.PostsDAO
	err = postDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Get all post records
	posts, err := postDAO.AllPostsSortedByDate()
	if err != nil {
		log.Error().Err(err).Str("service", "sitemap").Msg("Error getting all posts sorted by date in sitemap handler")
	}

	var mediaDAO mediadao.MediaDAO
	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Get all media records
	media, err := mediaDAO.AllMediaSortedByDate()
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	var b strings.Builder

	fmt.Fprintf(&b, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>")
	fmt.Fprintf(&b, "<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">")

	fmt.Fprintf(&b, "<url>")

	fmt.Fprintf(&b, fmt.Sprintf("<loc>https://%s/</loc>", cfg.GetSite()))

	fmt.Fprintf(&b, "<lastmod>2020-01-01</lastmod>")

	fmt.Fprintf(&b, "<changefreq>monthly</changefreq>")

	fmt.Fprintf(&b, "<priority>1.0</priority>")

	fmt.Fprintf(&b, "</url>")

	fmt.Fprintf(&b, "<url>")

	fmt.Fprintf(&b, fmt.Sprintf("<loc>https://%s</loc>", cfg.GetSite()))

	fmt.Fprintf(&b, "<lastmod>2020-01-01</lastmod>")

	fmt.Fprintf(&b, "<changefreq>monthly</changefreq>")

	fmt.Fprintf(&b, "<priority>1.0</priority>")

	fmt.Fprintf(&b, "</url>")

	for _, p := range posts {
		fmt.Fprintf(&b, "<url>")

		fmt.Fprintf(&b, fmt.Sprintf("<loc>https://%s/stories/%s</loc>", cfg.GetSite(), p.Slug))

		fmt.Fprintf(&b, fmt.Sprintf("<lastmod>%s</lastmod>", p.CreatedAt.Format("2006-01-02")))

		fmt.Fprintf(&b, "<changefreq>monthly</changefreq>")

		fmt.Fprintf(&b, "<priority>0.8</priority>")

		fmt.Fprintf(&b, "</url>")
	}

	for _, m := range media {
		fmt.Fprintf(&b, "<url>")

		fmt.Fprintf(&b, fmt.Sprintf("<loc>https://%s/photo/%s</loc>", cfg.GetSite(), m.Slug))

		fmt.Fprintf(&b, fmt.Sprintf("<lastmod>%s</lastmod>", m.CreatedAt.Format("2006-01-02")))

		fmt.Fprintf(&b, "<changefreq>monthly</changefreq>")

		fmt.Fprintf(&b, "<priority>0.8</priority>")

		fmt.Fprintf(&b, "</url>")
	}

	fmt.Fprintf(&b, "</urlset>")

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/xml")

	fmt.Fprintf(w, b.String())

}

// RequestBotAPI attempts to update additional information about the bot
func (ctx *HTTPHandlerContext) RequestBotAPI(w http.ResponseWriter, r *http.Request) {

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\",\"file\":\"error\"}\n"

	sess := util.GetSession(r)

	var requestviewDAO requestviewdao.RequestViewDAO

	err := requestviewDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "requestviewdao").Msg("Error initialzing media data access object ")
	}

	var d models.RequestView
	err = json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "getting data")
		return
	}

	rv, err := requestviewDAO.GetRequestViewByPTAG(d.PTag)
	if err != nil {
		log.Error().Err(err).Str("service", "requestviewdao").Msgf("Error getting requestview by id %s", d.PTag)
	}

	rv.BrowserVersion = d.BrowserVersion
	rv.FunctionalBrowser = d.FunctionalBrowser
	rv.SessionID = d.SessionID
	rv.NavAppVersion = d.NavAppVersion
	rv.NavBrowser = d.NavBrowser
	rv.NavPlatform = d.NavPlatform
	rv.OS = d.OS
	rv.OSVersion = d.OSVersion
	rv.UserAgent = d.UserAgent

	err = requestviewDAO.UpdateRequestView(&rv)
	if err != nil {
		log.Error().Err(err).Str("service", "requestviewdao").Msg("error udating requestview")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"request recieved %s\"}\n", sess.SessionToken)
	return
}