package views

import (
	//"context"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"

	//"bf.go/blog/mongo"
	//"bf.go/blog/models"

	"github.com/rsvancara/goblog/internal/config"
	"github.com/rsvancara/goblog/internal/models"
	"github.com/rsvancara/goblog/internal/requestfilter"
	"github.com/rsvancara/goblog/internal/session"
	"github.com/rsvancara/goblog/internal/util"

	"github.com/flosch/pongo2"
)

// SessionHandler manage session objects
func SessionHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var sess session.Session
		err := sess.Session(r, w)
		if err != nil {
			fmt.Printf("Session not available %s\n", err)
		}

		var ctxKey util.CtxKey
		ctxKey = "session"
		ctx := context.WithValue(r.Context(), ctxKey, sess)

		//fmt.Printf("session token created %s", sess.SessionToken)

		h.ServeHTTP(w, r.WithContext(ctx))

	})
}

// AuthHandler authorize user
func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var sess session.Session

		err := sess.Session(r, w)

		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if sess.User.IsAuth == false {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// GeoFilterMiddleware Middleware that matches paths to filter rules.
func GeoFilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var geoIP requestfilter.GeoIP

		ipaddress, _ := requestfilter.GetIPAddress(r)
		//fmt.Printf("IP Address: %s | request: %s\n", ipaddress, r.RequestURI)

		// for testing...we inject an IP Address
		//if ipaddress == "" {
		//	ipaddress = "73.83.74.114"
		//}

		geoIP.PageID = models.GenUUID()

		if ipaddress != "" && ipaddress != "[" {

			err := geoIP.Search(ipaddress)
			if err != nil {
				fmt.Printf("Error IP Address not found in the database for IP Address: %s with error %s\n", ipaddress, err)
			}

			if requestfilter.IsPrivateSubnet(geoIP.IPAddress) {
				// Handle situations where we have a private ip address
				// 	1. In development this is ok
				//  2. In production something should be considered wrong
				//  3. Send to capta page?
			}
			if geoIP.IsFound == true {
				// Apply filter rules
				// Filter on IP
				// Filter on City
				// Filter on Country
				// Filter on timezone
				// Filter on EU

				// Filters are based on request path,
				// path is matched to a list of rules in a database
				// and returned to be evaluated.
				// Based on the match condition, action is taken, allow, deny, redirect

			}
		}

		var ctxKey util.CtxKey
		ctxKey = "geoip"
		ctx := context.WithValue(r.Context(), ctxKey, geoIP)

		sess, err := util.SessionContext(r)
		if err != nil {
			fmt.Printf("Error getting session from context: %s\n", err)
		}

		//fmt.Printf("Found a context for session in geo module %s\n", sess.SessionToken)

		// Save a copy of this request for debugging.
		requestDump, err := httputil.DumpRequest(r, false)
		if err != nil {
			fmt.Printf("error getting raw request: %s", err)
		}
		rawRequest := string(requestDump)

		var rv models.RequestView
		rv.IPAddress = geoIP.IPAddress.String()
		rv.HeaderUserAgent = r.Header.Get("User-Agent")
		rv.PTag = geoIP.PageID
		rv.RequestURL = r.RequestURI
		rv.SessionID = sess.SessionToken
		rv.City = geoIP.City
		rv.Country = geoIP.CountryName
		rv.RawRequest = rawRequest
		err = rv.CreateRequestView()
		if err != nil {
			fmt.Printf("error creating requestview: %s\n", err)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HomeHandler Home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	// Get List
	posts, err := models.AllPostsSortedByDate()
	if err != nil {
		fmt.Println(err)
	}

	template, err := util.SiteTemplate("/index.html")
	//template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Index",
		"posts":     posts,
		"user":      sess.User,
		"bodyclass": "frontpage",
		"hidetitle": true,
		"pagekey":   util.GetPageID(r),
		"token":     sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// Signin Sign into the application
func Signin(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		var creds session.Credentials
		creds.Username = r.FormValue("username")
		creds.Password = r.FormValue("password")

		if creds.Username == "" || creds.Password == "" {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		isAuth, err := sess.Authenticate(creds, r, w)
		if err != nil {
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error authenticating user %s with error %s", creds.Username, err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		if isAuth == false {
			fmt.Printf("SIGNUP - User is authenticated %s %s\n", sess.SessionToken, sess.User.Username)
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

// AdminHome admin home page
func AdminHome(w http.ResponseWriter, r *http.Request) {

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
func AboutHandler(w http.ResponseWriter, r *http.Request) {

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
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "healthy")
}

// ContactHandler defines a healthcheck
func ContactHandler(w http.ResponseWriter, r *http.Request) {
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
func SiteMap(w http.ResponseWriter, r *http.Request) {

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("could not get configuration object %s", (err))
		return
	}

	// Get all post records
	posts, err := models.AllPostsSortedByDate()
	if err != nil {
		fmt.Println(err)
	}

	// Get all media records
	media, err := models.AllMediaSortedByDate()
	if err != nil {
		fmt.Println(err)
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

// WPLoginHandler handles fake wordpress login requests.  Log the request
// to process for adding to permenant block list
func WPLoginHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("could not get configuration object %s", (err))
		return
	}

	// Test if we are a POST to capture form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		var f models.FakeRequest

		f.URL = "/wp-login.php"
		f.IPAddress = sess.User.IPAddress
		f.City = sess.User.City
		f.Country = sess.User.Country
		f.TimeZone = sess.User.TimeZone
		f.Username = r.FormValue("log")
		f.Password = r.FormValue("pwd")
		f.UserAgent = r.Header.Get("User-Agent")
		f.SessionID = sess.SessionToken

		//fmt.Printf("Fake WP Login -> login: %s, Password: %s\n", f.Username, f.Password)

		// Create Record
		err = f.InsertFakeRequest()
		if err != nil {
			fmt.Printf("error inserting fakereqest: %s\n", err)
		}

		// Send them to the admin page so they think they logged in.
		http.Redirect(w, r, "/wp-admin", http.StatusSeeOther)
		return

	}

	template, err := util.SiteTemplate("/evil/wp-login.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "WP-Login",
		"site":    cfg.Site,
		"user":    sess.User,
		"pagekey": util.GetPageID(r),
		"token":   sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// WPAdminHandler provides fake admin page
func WPAdminHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	template, err := util.SiteTemplate("/evil/wp-admin.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "WP-Login",
		"sessionid": sess.SessionToken,
		"user":      sess.User,
		"pagekey":   util.GetPageID(r),
		"token":     sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// RequestBotAPI attempts to update additional information about the bot
func RequestBotAPI(w http.ResponseWriter, r *http.Request) {

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\",\"file\":\"error\"}\n"

	sess := util.GetSession(r)

	var d models.RequestView
	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "getting data")
		return
	}

	var rv models.RequestView
	err = rv.GetRequestViewByPTAG(d.PTag)
	if err != nil {
		fmt.Printf("error getting requestview by id %s: %s \n", d.PTag, err)
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

	err = rv.UpdateRequestView()
	if err != nil {
		fmt.Printf("error udating requestview: %s", err)
	}

	//fmt.Println(d)
	// Need to do some database work here and interface with pageview model

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"request recieved %s\"}\n", sess.SessionToken)
	return
}
