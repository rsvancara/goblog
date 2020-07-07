package views

import (
	"blog/blog/models"
	"blog/blog/util"
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

// ViewCategoryHandler View the media
func ViewCategoryHandler(w http.ResponseWriter, r *http.Request) {

	var media models.MediaModel
	var medialist []models.MediaModel
	var err error

	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["category"]; ok {
		// Load Media
		medialist, err = models.GetMediaListByCategory(vars["category"])
		if err != nil {
			fmt.Printf("Error getting media with variable, category %s with error %s", val, err)
		}
	} else {
		fmt.Printf("Error getting url variable, category: %s", val)
	}

	template, err := util.SiteTemplate("/category.html")
	//template := "templates/admin/mediaview.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     fmt.Sprintf("Category - %s", vars["category"]),
		"media":     media,
		"user":      sess.User,
		"bodyclass": "",
		"fluid":     true,
		"hidetitle": true,
		"medialist": medialist,
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

//ViewCategoriesHandler view all categories
func ViewCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	medialist, err := models.AllCategories()
	if err != nil {
		fmt.Printf("Error getting media list with error %s", err)
	}

	fmt.Println(medialist)

	template, err := util.SiteTemplate("/categories.html")
	//template := "templates/admin/mediaview.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     fmt.Sprintf("Media Categories"),
		"user":      sess.User,
		"bodyclass": "",
		"fluid":     true,
		"hidetitle": true,
		"medialist": medialist,
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
