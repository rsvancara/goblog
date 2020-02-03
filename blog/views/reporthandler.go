package views

import (
	"blog/blog/session"
	"blog/blog/util"
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
)

// SessionReportHandler build a list of current user sessions
func SessionReportHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Build Session List")

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s\n", err)
	}

	// Get List
	var sessions []session.Session

	sessions, err = session.GetAllSessions()

	template, err := util.SiteTemplate("/admin/sessions.html")
	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Session Report",
		"sessions":  sessions,
		"user":      sess.User,
		"bodyclass": "",
		"hidetitle": true,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

	return
}
