package views

import (
	"blog/blog/models"
	"blog/blog/session"
	"blog/blog/util"
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
)

// AffiliateHandler view list of affiliates
func AffiliateHandler(w http.ResponseWriter, r *http.Request) {

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	// Create Record
	affiliates, err := models.GetAllAffiliateOrderByDate()
	if err != nil {
		fmt.Println(err)
	}

	template, err := util.SiteTemplate("/admin/affiliates.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"user":       sess.User,
		"title":      "Affiliates",
		"affiliates": affiliates,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

}

// AffiliateAddHandler view list of affiliates
func AffiliateAddHandler(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template, err := util.SiteTemplate("/admin/affiliatesadd.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"user":  sess.User,
		"title": "Affiliates",
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

}

// AffiliateEditHandler view list of affiliates
func AffiliateEditHandler(w http.ResponseWriter, r *http.Request) {

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template, err := util.SiteTemplate("/admin/affiliatesedit.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title": "Affiliates",
		"user":  sess.User,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

}

// AffiliateDeleteHandler view list of affiliates
func AffiliateDeleteHandler(w http.ResponseWriter, r *http.Request) {

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

}
