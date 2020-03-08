package views

import (
	"blog/blog/util"
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
)

// FilterHandler View File
func FilterHandler(w http.ResponseWriter, r *http.Request) {

	sess := GetSession(r)

	template, err := util.SiteTemplate("/admin/filters.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "View Media",
		"user":    sess.User,
		"pagekey": GetPageID(r),
		"token":   sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

}
