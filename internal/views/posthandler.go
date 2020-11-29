package views

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/rsvancara/goblog/internal/models"
	"github.com/rsvancara/goblog/internal/session"
	"github.com/rsvancara/goblog/internal/util"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// PhotoView View File
func PhotoView(w http.ResponseWriter, r *http.Request) {

	var media models.MediaModel

	sess := util.GetSession(r)

	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	// Load Media
	err := media.GetMediaBySlug(vars["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	template, err := util.SiteTemplate("/mediaview.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":           "View Media",
		"media":           media,
		"user":            sess.User,
		"bodyclass":       "",
		"fluid":           true,
		"hidetitle":       true,
		"exposureprogram": media.GetExposureProgramTranslated(),
		"pagekey":         util.GetPageID(r),
		"token":           sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// GetMediaAPI View File
func GetMediaAPI(w http.ResponseWriter, r *http.Request) {

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\",\"image\":\"error\",\"url\":\"/static/no-image.svg\",\"refurl\":\"#\"}\n"

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if _, ok := vars["id"]; ok {

	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, "Error find ID", "ID was not available inthe URL or could not be parsed")
		return
	}

	// Media Object
	var media models.MediaModel

	err := media.GetMedia(vars["id"])
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "could not get media object from database")
		return
	}

	s3URL := "/image/" + media.Slug + "/large"
	refURL := "/photo/" + media.Slug

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"media found\",\"url\":\"%s\",\"refurl\":\"%s\",\"title\":\"%s\",\"slug\":\"%s\",\"category\":\"%s\"}\n", s3URL, refURL, media.Title, media.Slug, media.Category)

	return
}

// PostView Home page
func PostView(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	// Model
	var pm models.PostModel

	// Load Model
	err := pm.GetPostBySlug(vars["id"])
	if err != nil {
		fmt.Printf("Error getting object from database: %s", err)
	}

	md := []byte(pm.Post)
	var buf bytes.Buffer

	gm := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	err = gm.Convert(md, &buf)
	if err != nil {
		fmt.Printf("Error rendering markdown: %s", err)
	}

	template, err := util.SiteTemplate("/post.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "Index",
		"post":    pm,
		"content": buf.String(),
		"user":    sess.User,
		"pagekey": util.GetPageID(r),
		"token":   sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// Post post
func Post(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	// Create Record
	posts, err := models.AllPostsSortedByDate()
	if err != nil {
		fmt.Println(err)
	}

	template, err := util.SiteTemplate("/admin/post.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "Index",
		"posts":   posts,
		"user":    sess.User,
		"pagekey": util.GetPageID(r),
		"token":   sess.SessionToken,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// PostEdit edit the post
func PostEdit(w http.ResponseWriter, r *http.Request) {

	// Form Management Variables
	titleMessage := ""
	titleMessageError := false
	postMessage := ""
	postMessageError := false
	statusMessage := ""
	statusMessageError := false
	featuredMessage := ""
	featuredMessageError := false
	postTeaserMessage := ""
	postTeaserMessageError := false
	postKeywordsMessage := ""
	postKeywordsMessageError := false

	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		//do something here
		fmt.Println(val)
	}

	// Model
	var pm models.PostModel

	// Load Model
	err := pm.GetPost(vars["id"])
	if err != nil {
		fmt.Printf("error getting post from database for id %s with error %s", vars["id"], err)
	}

	var teaserImageURL string
	// Get image URL if teaser image is present
	if pm.TeaserImage != "" {
		var mm models.MediaModel
		err = mm.GetMedia(pm.TeaserImage)
		if err != nil {
			fmt.Printf("error getting image from database for id %s with error %s", pm.TeaserImage, err)
		}
		teaserImageURL = mm.Slug
	}

	// Test if we are a POST to capture form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		pm.Post = r.FormValue("inputPost")
		pm.Title = r.FormValue("inputTitle")
		pm.Status = r.FormValue("inputStatus")
		pm.Featured = r.FormValue("inputFeatured")
		pm.PostTeaser = r.FormValue("inputPostTeaser")
		pm.Keywords = r.FormValue("inputKeywords")
		pm.TeaserImage = r.FormValue("inputTeaserImage")

		var mm models.MediaModel
		if pm.TeaserImage != "" {
			mm.GetMedia(pm.TeaserImage)
			pm.TeaserImageSlug = mm.Slug
		}

		// Do validation here
		validate := true
		if pm.Title == "" {
			validate = false
			titleMessage = "Please provide a title"
			titleMessageError = true
		}

		if pm.Post == "" {
			validate = false
			postMessage = "Please provide post content"
			postMessageError = true
		}

		if pm.PostTeaser == "" {
			validate = false
			postTeaserMessage = "Please provide post teaser"
			postTeaserMessageError = true
		}

		if pm.Status == "enabled" || pm.Status == "disabled" {

		} else {
			statusMessage = "Invalid status code"
			statusMessageError = true
		}

		if pm.Featured == "yes" || pm.Featured == "no" {

		} else {
			featuredMessage = "Invalid status code"
			featuredMessageError = true
		}

		if pm.Keywords == "" {
			validate = false
			postKeywordsMessage = "Please provide post keywords"
			postKeywordsMessageError = true
		}

		if validate == true {

			// Create Record
			err = pm.UpdatePost()
			if err != nil {
				fmt.Println(err)
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, "/admin/post", http.StatusSeeOther)
			return
		}
	}

	// HTTP Template
	template, err := util.SiteTemplate("/admin/postedit.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":                    "Edit Post",
		"post":                     pm,
		"user":                     sess.User,
		"postMessage":              postMessage,
		"postMessageError":         postMessageError,
		"titleMessage":             titleMessage,
		"titleMessageError":        titleMessageError,
		"statusMessage":            statusMessage,
		"statusMessageError":       statusMessageError,
		"featuredMessage":          featuredMessage,
		"featuredMessageError":     featuredMessageError,
		"postTeaserMessage":        postTeaserMessage,
		"postTeaserMessageError":   postTeaserMessageError,
		"postKeywordsMessage":      postKeywordsMessage,
		"postKeywordsMessageError": postKeywordsMessageError,
		"pagekey":                  util.GetPageID(r),
		"token":                    sess.SessionToken,
		"teaserImageUrl":           teaserImageURL,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// PostAdminView view the post
func PostAdminView(w http.ResponseWriter, r *http.Request) {

	//http Session
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		fmt.Printf("Error no id was found for post: %s", val)
	}

	// Model
	var pm models.PostModel

	// Load Model
	pm.GetPost(vars["id"])

	md := []byte(pm.Post)
	var buf bytes.Buffer

	gm := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithUnsafe(),
		),
	)

	err = gm.Convert(md, &buf)
	if err != nil {
		fmt.Printf("Error rendering markdown: %s", err)
	}

	// HTTP Template
	template, err := util.SiteTemplate("/admin/postview.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "Edit Post",
		"post":    pm,
		"content": buf.String,
		"user":    sess.User,
		"pagekey": util.GetPageID(r),
		"token":   sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Printf("error rendering template: %s", err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// PostAdd add post
func PostAdd(w http.ResponseWriter, r *http.Request) {

	var teaserImageURL string
	var pm models.PostModel

	// Form Variables
	titleMessage := ""
	titleMessageError := false
	postMessage := ""
	postMessageError := false
	statusMessage := ""
	statusMessageError := false
	featuredMessage := ""
	featuredMessageError := false
	postTeaserMessage := ""
	postTeaserMessageError := false
	postKeywordsMessage := ""
	postKeywordsMessageError := false

	// HTTP Session
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	// Test if we are a POST to capture form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		pm.Post = r.FormValue("inputPost")
		pm.Title = r.FormValue("inputTitle")
		pm.Status = r.FormValue("inputStatus")
		pm.Featured = r.FormValue("inputFeatured")
		pm.PostTeaser = r.FormValue("inputPostTeaser")
		pm.Keywords = r.FormValue("inputKeywords")
		pm.PostTeaser = r.FormValue("inputPostTeaser")
		pm.TeaserImage = r.FormValue("inputTeaserImage")

		var mm models.MediaModel
		if pm.TeaserImage != "" {
			mm.GetMedia(pm.TeaserImage)
			pm.TeaserImageSlug = mm.Slug
		}

		if err != nil {
			fmt.Printf("Error converting status to integer in post form: %s\n", err)
		}

		// Do validation here
		validate := true
		if pm.Title == "" {
			validate = false
			titleMessage = "Please provide a title"
			titleMessageError = true
		}

		if pm.Post == "" {
			validate = false
			postMessage = "Please provide post content"
			postMessageError = true
		}

		if pm.PostTeaser == "" {
			validate = false
			postTeaserMessage = "Please provide post teaser"
			postTeaserMessageError = true
		}

		if pm.Status == "enabled" || pm.Status == "disabled" {

		} else {
			statusMessage = "Invalid status code"
			statusMessageError = true
		}

		if pm.Featured == "yes" || pm.Featured == "no" {

		} else {
			featuredMessage = "Invalid status code"
			featuredMessageError = true
		}

		if pm.Keywords == "" {
			validate = false
			postKeywordsMessage = "Please provide post keywords"
			postKeywordsMessageError = true
		}

		if validate == true {

			// Create Record
			err = pm.InsertPost()
			if err != nil {
				fmt.Printf("error inserting post: %s\n", err)
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, "/admin/post", http.StatusSeeOther)
			return
		}
	}

	template, err := util.SiteTemplate("/admin/postadd.html")
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":                    "Add Post",
		"post":                     pm,
		"user":                     sess.User,
		"postMessage":              postMessage,
		"postMessageError":         postMessageError,
		"titleMessage":             titleMessage,
		"titleMessageError":        titleMessageError,
		"statusMessage":            statusMessage,
		"statusMessageError":       statusMessageError,
		"featuredMessage":          featuredMessage,
		"featuredMessageError":     featuredMessageError,
		"postTeaserMessage":        postTeaserMessage,
		"postTeaserMessageError":   postTeaserMessageError,
		"postKeywordsMessage":      postKeywordsMessage,
		"postKeywordsMessageError": postKeywordsMessageError,
		"pagekey":                  util.GetPageID(r),
		"token":                    sess.SessionToken,
		"teaserImageUrl":           teaserImageURL,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// PostDelete delete post
func PostDelete(w http.ResponseWriter, r *http.Request) {

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		//do something here
		fmt.Println(val)
	}

	// Model
	var pm models.PostModel

	// Load Model
	pm.GetPost(vars["id"])

	pm.DeletePost()

	http.Redirect(w, r, "/admin/post", http.StatusSeeOther)
}

// NormalizeNewlines normalizes \r\n (windows) and \r (mac)
// into \n (unix)
func NormalizeNewlines(d []byte) []byte {
	// replace CR LF \r\n (windows) with LF \n (unix)
	d = bytes.Replace(d, []byte{13, 10}, []byte{10}, -1)
	// replace CF \r (mac) with LF \n (unix)
	d = bytes.Replace(d, []byte{13}, []byte{10}, -1)
	return d
}