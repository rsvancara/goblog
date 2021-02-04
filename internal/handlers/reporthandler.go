package handlers

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	requestviewdao "goblog/internal/dao/requestview"
	"goblog/internal/session"
	"goblog/internal/util"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

// SessionReportHandler build a list of current user sessions
func (ctx *HTTPHandlerContext) SessionReportHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	var sessions []session.Session
	sessions, err := session.GetAllSessions()

	template, err := util.SiteTemplate("/admin/sessions.html")
	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Session Report",
		"sessions":  sessions,
		"user":      sess.User,
		"bodyclass": "",
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

	return
}

// SessionDetailsReportHandler build a list of current user sessions
func (ctx *HTTPHandlerContext) SessionDetailsReportHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
		return
	}

	var requestviewDAO requestviewdao.RequestViewDAO
	err := requestviewDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	//var rv models.RequestView
	rvs, err := requestviewDAO.GetRequestViewsBySessionID(vars["id"])
	if err != nil {
		fmt.Printf("Error getting session details for %s with error %s", vars["id"], err)
	}

	template, err := util.SiteTemplate("/admin/sessiondetail.html")
	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":        "Session Report",
		"requestviews": rvs,
		"user":         sess.User,
		"sessionid":    vars["id"],
		"bodyclass":    "",
		"hidetitle":    true,
		"pagekey":      util.GetPageID(r),
		"token":        sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

	return
}

// SessionDeleteHandler build a list of current user sessions
func (ctx *HTTPHandlerContext) SessionDeleteHandler(w http.ResponseWriter, r *http.Request) {

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
		return
	}

	err := session.DeleteSession(vars["id"])
	if err != nil {
		fmt.Printf("error deleting session %s\n", err)
	}

	http.Redirect(w, r, "/admin/sessions", http.StatusSeeOther)

	return
}

// RequestInspectorReportHandler Get details of a request
func (ctx *HTTPHandlerContext) RequestInspectorReportHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
		return
	}

	var requestviewDAO requestviewdao.RequestViewDAO
	err := requestviewDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	//var rv models.RequestView
	rv, err := requestviewDAO.GetRequestViewByPTAG(vars["id"])
	if err != nil {
		fmt.Printf("Error getting session details for %s with error %s", vars["id"], err)
	}

	template, err := util.SiteTemplate("/admin/requestdetail.html")
	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Request View",
		"req":       rv,
		"user":      sess.User,
		"sessionid": vars["id"],
		"bodyclass": "",
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

	return
}
