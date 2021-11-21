package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"goblog/internal/models"
	"goblog/internal/util"

	"github.com/flosch/pongo2"
	"github.com/rs/zerolog/log"
)

// FilterHandler View File
func (ctx *HTTPHandlerContext) FilterHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	template, err := util.SiteTemplate("/admin/filters.html")
	if err != nil {
		log.Error().Err(err)
	}

	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "View Media",
		"user":    sess.User,
		"pagekey": util.GetPageID(r),
		"token":   sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)

}

// CreateFilterHandler View File
func (ctx *HTTPHandlerContext) CreateFilterHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	template, err := util.SiteTemplate("/admin/filtersadd.html")
	if err != nil {
		log.Error().Err(err)
	}
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "View Media",
		"user":    sess.User,
		"pagekey": util.GetPageID(r),
		"token":   sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)

}

// CreateAPIFilterHandler handles creation of filters via AJAX
func (ctx *HTTPHandlerContext) CreateAPIFilterHandler(w http.ResponseWriter, r *http.Request) {

	//sess := util.GetSession(r)

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\"}\n"

	//vars := mux.Vars(r)

	var filter models.Filter

	err := json.NewDecoder(r.Body).Decode(&filter)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "getting data")
		return
	}

}
