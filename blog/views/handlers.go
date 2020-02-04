package views

import (
	//"context"
	"context"
	"fmt"
	"net/http"

	//"bf.go/blog/mongo"
	//"bf.go/blog/models"

	"blog/blog/models"
	"blog/blog/requestfilter"
	"blog/blog/session"
	"blog/blog/util"

	"github.com/flosch/pongo2"
)

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
		fmt.Printf("IP Address: %s | request: %s\n", ipaddress, r.RequestURI)

		// for testing...we inject an IP Address
		//if ipaddress == "" {
		//	ipaddress = "73.83.74.114"
		//}

		if ipaddress != "" {

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

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HomeHandler Home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s\n", err)
	}

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

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s\n", err)
	}

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
		}

		isAuth, err := sess.Authenticate(creds, r, w)
		if err != nil {
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Printf("error authenticating user %s with error %s", creds.Username, err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
		}
		fmt.Printf("SIGNUP - User is authenticated %s %s\n", sess.SessionToken, sess.User.Username)

		if isAuth == false {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	template, err := util.SiteTemplate("/signin.html")
	//template := "templates/signin.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error loading template with error: %s\n", err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// AdminHome admin home page
func AdminHome(w http.ResponseWriter, r *http.Request) {

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template, err := util.SiteTemplate("/admin/admin.html")
	//template := "templates/admin/admin.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// AboutHandler about page
func AboutHandler(w http.ResponseWriter, r *http.Request) {

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template, err := util.SiteTemplate("/about.html")
	//template := "templates/about.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}
