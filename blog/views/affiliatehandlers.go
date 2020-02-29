package views

import (
	"blog/blog/models"
	"blog/blog/session"
	"blog/blog/util"
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
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

		if validate == true {

			// Create Record
			err = af.InsertAffiliate()
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

		if validate == true {

			// Create Record
			err = af.EditAffiliate()
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
	//http Session
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		//do something here
		fmt.Println(val)
	}

	// Model
	var af models.Affiliate

	// Load Model
	err = af.GetAffiliate(vars["id"])
	if err != nil {
		fmt.Printf("Error getting affiliate by id %s with error %s\n", vars["id"], err)
		return
	}

	af.DeleteAffiliate()

	http.Redirect(w, r, "/admin/affiliates", http.StatusSeeOther)

}

// AffiliateBouncyHouseHandler view list of affiliates
func AffiliateBouncyHouseHandler(w http.ResponseWriter, r *http.Request) {

	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		//do something here
		fmt.Println(val)
	}

	// Model
	var af models.Affiliate

	// Load Model
	err = af.GetAffiliate(vars["id"])
	if err != nil {
		fmt.Printf("Error getting affiliate by id %s with error %s\n", vars["id"], err)
		return
	}

	//http.Redirect(w, r, af.AffiliateLink, http.StatusSeeOther)

	template, err := util.SiteTemplate("/bouncyhouse.html")
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
	fmt.Fprintf(w, out)

}
