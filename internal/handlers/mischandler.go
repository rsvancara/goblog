//handlers provides misceleneous handlers for view layer
package handlers

import (
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/rs/zerolog/log"
	"github.com/rsvancara/goblog/internal/session"
	"github.com/rsvancara/goblog/internal/util"
)

// Signin Sign into the application
func SignInHandler(w http.ResponseWriter, r *http.Request) {

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
