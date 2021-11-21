package handlers

import (
	"goblog/internal/models"

	"github.com/rs/zerolog/log"

	//"blog/blog/session"
	"fmt"
	"net/http"

	affiliatesdao "goblog/internal/dao/affiliates"

	"goblog/internal/util"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

// AffiliateHandler view list of affiliates
func (ctx *HTTPHandlerContext) AffiliateHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	var affiliatesDAO affiliatesdao.AffiliatesDAO
	err := affiliatesDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Create Record
	affiliates, err := affiliatesDAO.GetAllAffiliateOrderByDate()
	if err != nil {
		fmt.Println(err)
	}

	template, err := util.SiteTemplate("/admin/affiliates.html")
	if err != nil {
		log.Error().Err(err)
	}
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"user":       sess.User,
		"title":      "Affiliates",
		"affiliates": affiliates,
		"pagekey":    util.GetPageID(r),
		"token":      sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)

}

// AffiliateAddHandler view list of affiliates
func (ctx *HTTPHandlerContext) AffiliateAddHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	var af models.Affiliate
	// Form Variables
	titleMessage := ""
	titleMessageError := false
	urlMessage := ""
	urlMessageError := false
	categoryMessage := ""
	categoryMessageError := false

	// Test if we are a POST to capture form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		af.AffiliateLink = r.FormValue("inputURL")
		af.AffiliateTitle = r.FormValue("inputTitle")
		af.Category = r.FormValue("inputCategory")
		af.Description = r.FormValue("inputDescription")

		// Do validation here
		validate := true
		if af.AffiliateTitle == "" {
			validate = false
			titleMessage = "Please provide a title"
			titleMessageError = true
		}

		if af.AffiliateTitle == "" {
			validate = false
			urlMessage = "Please provide post content"
			urlMessageError = true
		}

		if af.Category == "" {
			validate = false
			categoryMessage = "Please provide post teaser"
			categoryMessageError = true
		}

		if validate {

			var affiliatesDAO affiliatesdao.AffiliatesDAO
			err := affiliatesDAO.Initialize(ctx.dbClient, ctx.hConfig)
			if err != nil {
				log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
			}

			// Create Record
			err = affiliatesDAO.InsertAffiliate(&af)
			if err != nil {
				fmt.Printf("error inserting post: %s\n", err)
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, "/admin/affiliates", http.StatusSeeOther)
			return
		}
	}

	template, err := util.SiteTemplate("/admin/affiliatesadd.html")
	if err != nil {
		log.Error().Err(err)
	}
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":                "Add Post",
		"affiliate":            af,
		"user":                 sess.User,
		"urlMessage":           urlMessage,
		"urlMessageError":      urlMessageError,
		"titleMessage":         titleMessage,
		"titleMessageError":    titleMessageError,
		"categoryMessage":      categoryMessage,
		"categoryMessageError": categoryMessageError,
		"pagekey":              util.GetPageID(r),
		"token":                sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)

}

// AffiliateEditHandler view list of affiliates
func (ctx *HTTPHandlerContext) AffiliateEditHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		//do something here
		fmt.Println(val)
	}

	var af models.Affiliate

	af.GetAffiliate(vars["id"])

	// Form Variables
	titleMessage := ""
	titleMessageError := false
	urlMessage := ""
	urlMessageError := false
	categoryMessage := ""
	categoryMessageError := false

	// Test if we are a POST to capture form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		af.AffiliateLink = r.FormValue("inputURL")
		af.AffiliateTitle = r.FormValue("inputTitle")
		af.Category = r.FormValue("inputCategory")
		af.Description = r.FormValue("inputDescription")

		// Do validation here
		validate := true
		if af.AffiliateTitle == "" {
			validate = false
			titleMessage = "Please provide a title"
			titleMessageError = true
		}

		if af.AffiliateTitle == "" {
			validate = false
			urlMessage = "Please provide post content"
			urlMessageError = true
		}

		if af.Category == "" {
			validate = false
			categoryMessage = "Please provide post teaser"
			categoryMessageError = true
		}

		if validate {

			var affiliatesDAO affiliatesdao.AffiliatesDAO
			err := affiliatesDAO.Initialize(ctx.dbClient, ctx.hConfig)
			if err != nil {
				log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
			}

			// Create Record
			err = affiliatesDAO.EditAffiliate(&af)
			if err != nil {
				fmt.Printf("error inserting post: %s\n", err)
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, "/admin/affiliates", http.StatusSeeOther)
			return
		}
	}

	template, err := util.SiteTemplate("/admin/affiliatesedit.html")
	if err != nil {
		log.Error().Err(err)
	}

	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":                "Add Post",
		"affiliate":            af,
		"user":                 sess.User,
		"urlMessage":           urlMessage,
		"urlMessageError":      urlMessageError,
		"titleMessage":         titleMessage,
		"titleMessageError":    titleMessageError,
		"categoryMessage":      categoryMessage,
		"categoryMessageError": categoryMessageError,
		"pagekey":              util.GetPageID(r),
		"token":                sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)

}

// AffiliateDeleteHandler view list of affiliates
func (ctx *HTTPHandlerContext) AffiliateDeleteHandler(w http.ResponseWriter, r *http.Request) {
	//sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		//do something here
		fmt.Println(val)
	}

	var affiliatesDAO affiliatesdao.AffiliatesDAO
	err := affiliatesDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Model
	var af models.Affiliate

	// Load Model
	af, err = affiliatesDAO.GetAffiliate(vars["id"])
	if err != nil {
		fmt.Printf("Error getting affiliate by id %s with error %s\n", vars["id"], err)
		return
	}

	affiliatesDAO.DeleteAffiliate(&af)

	http.Redirect(w, r, "/admin/affiliates", http.StatusSeeOther)

}

// AffiliateBouncyHouseHandler view list of affiliates
func (ctx *HTTPHandlerContext) AffiliateBouncyHouseHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		//do something here
		fmt.Println(val)
	}

	var affiliatesDAO affiliatesdao.AffiliatesDAO
	err := affiliatesDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Load Model
	af, err := affiliatesDAO.GetAffiliate(vars["id"])
	if err != nil {
		fmt.Printf("Error getting affiliate by id %s with error %s\n", vars["id"], err)
		return
	}

	//http.Redirect(w, r, af.AffiliateLink, http.StatusSeeOther)

	template, err := util.SiteTemplate("/bouncyhouse.html")
	if err != nil {
		log.Error().Err(err)
	}

	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Add Post",
		"affiliate": af,
		"user":      sess.User,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)

}
