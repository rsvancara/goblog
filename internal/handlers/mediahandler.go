package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"

	mediadao "goblog/internal/dao/media"
	mediatags "goblog/internal/dao/mediatags"
	_ "goblog/internal/filters" //import pongo  plugins
	"goblog/internal/models"
	"goblog/internal/requestfilter"
	simplestorageservice "goblog/internal/s3"
	"goblog/internal/session"
	"goblog/internal/util"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

// MediaHandler HTTP Handler for View full list of media sorted by date in admin view
func (ctx *HTTPHandlerContext) MediaHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	medialist, err := mediaDAO.AllMediaSortedByDate()
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error retrieving objects from media database ")
	}

	template, err := util.SiteTemplate("/admin/media.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Index",
		"media":     medialist,
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
	fmt.Fprint(w, out)
}

// ViewMediaHandler HTTP Handler to View the media in admin view
func (ctx *HTTPHandlerContext) ViewMediaHandler(w http.ResponseWriter, r *http.Request) {

	var media models.MediaModel

	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {

		log.Error().Msgf("Error getting url variable, id: %s", val)
	}

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	media, err = mediaDAO.GetMedia(vars["id"])
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error retrieving data from access object ")
	}

	template, err := util.SiteTemplate("/admin/mediaview.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/mediaview.html"
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
	fmt.Fprint(w, out)
}

// MediaAddHandler HTTP Handler to view admin add media page
func (ctx *HTTPHandlerContext) MediaAddHandler(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		log.Printf("Session not available %s", err)
	}

	template, err := util.SiteTemplate("/admin/mediaadd.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/mediaadd.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error loading template %s", err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)
}

//PutMediaAPI Supports multi file upload in an API used in admin interface
func (ctx *HTTPHandlerContext) PutMediaAPIV2(w http.ResponseWriter, r *http.Request) {
	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\",\"file\":\"error\"}\n"

	vars := mux.Vars(r)

	data := make(map[string]string)

	data["keywords"] = r.FormValue("keywords")
	data["description"] = r.FormValue("description")
	data["title"] = r.FormValue("title")
	data["category"] = r.FormValue("category")
	data["location"] = r.FormValue("location")
	data["sitekey"] = "tryingadventure"

	//err = r.ParseForm()
	err := r.ParseMultipartForm(90 << 20)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "parsing multipart form")
		log.Error().Err(err).Msgf("Error parsing multipart form")
		//log.Printf("Error parsing multipart form %s", err)
		return
	}

	// Copy file information from the form

	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		log.Error().Err(err).Msg("Error opening file from form post")

		fmt.Fprintf(w, errorMessage, err, "Could not find uploaded form file in the post reqest")
		return
	}

	defer file.Close() // Close the file when we finish

	// Create the path to store the temporary file
	letterBytes := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	n := 12
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	tpath := fmt.Sprintf("temp/%s", string(b))

	_, err = os.Stat(tpath)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(tpath, 0755)
		if errDir != nil {
			log.Error().Err(err).Msg("Error creating temporary directory")
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, errorMessage, "Error creating directory", "")
		}
	}

	// Create the file handle for storing the file
	f, err := os.OpenFile(tpath+"/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		log.Error().Err(err).Msg("Error copying file from post to temporary file")
		fmt.Fprintf(w, errorMessage, err, "Error copying file from post to temporary file")
		return
	}

	defer f.Close()

	// Finally copy form file to the filehandle
	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		log.Error().Err(err).Msg("Error copying temporary file to destination path")

		fmt.Fprintf(w, errorMessage, err, "Error copying temporary file to destination path")
		return
	}

	cfg := ctx.hConfig

	log.Info().Msgf("Uploading image %s to image processing service at %s/api/meida/add/v1", handler.Filename, cfg.ImageService)

	req, err := util.ImageUploadRequest(cfg.GetImageServiceURI()+"/api/media/add/v1", data, "file", tpath+"/"+handler.Filename)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		log.Error().Err(err).Msg("Error creating request")
		fmt.Fprintf(w, errorMessage, err, "Error creating request")
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		log.Error().Err(err).Msg("Error uploading file and parameters")
		fmt.Fprintf(w, errorMessage, err, "Error uploading file and parameters")
		return
	}

	log.Info().Msgf("Uploading image %s to image processing service at %s/api/media/add/v1 with status code %d", handler.Filename, cfg.ImageService, resp.StatusCode)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		log.Error().Err(err).Msgf("Bad status code returned: %d", resp.StatusCode)
		fmt.Fprintf(w, errorMessage, fmt.Sprintf("Bad status code returned: %d", resp.StatusCode), "")
	}

	// Remove the file when done
	err = os.Remove(tpath + "/" + handler.Filename)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		log.Error().Err(err).Msg("Error removing temporary file")
		fmt.Fprintf(w, errorMessage, err, "Error removing temporary directory")
		return
	}

	// Remove the directory too
	err = os.Remove(tpath)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		log.Error().Err(err).Msg("Error removing temporary image directory")
		fmt.Fprintf(w, errorMessage, err, "Error removing temporary image directory")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"file %s uploaded\",\"file\":\"%s\"}\n", vars["id"], handler.Filename)

}

//MediaUpdateTitleHandler update the media title API
func (ctx *HTTPHandlerContext) MediaUpdateTitleHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	type Title struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	}

	var title Title

	err := json.NewDecoder(r.Body).Decode(&title)
	if err != nil {
		log.Error().Err(err).Str("service", "media").Msg("Error decoding json string ")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"Error: %s\", \"session\":\"%s\"}\n", err, sess.SessionToken)
		return
	}

	var mediaDAO mediadao.MediaDAO

	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"Error: %s\", \"session\":\"%s\"}\n", err, sess.SessionToken)
		return
	}

	model, err := mediaDAO.GetMedia(title.ID)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msgf("Error getting media for %s ", title.ID)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"Error finding media object: %s\", \"session\":\"%s\"}\n", err, sess.SessionToken)
		return
	}

	model.Title = title.Title

	err = mediaDAO.UpdateMedia(model)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msgf("Error updating media object for %s with title %s ", title.ID, title.Title)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"Error udating media object:  %s\", \"session\":\"%s\"}\n", err, sess.SessionToken)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"Title updated successfully\", \"session\":\"%s\"}\n", sess.SessionToken)

}

//MediaSearchAPIHandler search by media tags
func (ctx *HTTPHandlerContext) MediaSearchAPIHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	type Search struct {
		Search string `json:"search"`
	}

	var search Search

	// Decode the search string
	err := json.NewDecoder(r.Body).Decode(&search)
	if err != nil {
		log.Error().Err(err).Str("service", "media").Msg("Error decoding json string ")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"%s\", \"session\":\"%s\",\"results\":nil}\n", err, sess.SessionToken)
		return
	}

	log.Info().Msgf("Searching for %s", search.Search)

	//bodyString := "{}"
	//bodyBytes, err := ioutil.ReadAll(r.Body)
	//if err != nil {
	//	w.Header().Set("Content-Type", "application/json")
	//	fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"Error: %s\", \"session\":\"%s\",\"results\":nil}\n", err, sess.SessionToken)
	//	return
	//}
	//bodyString = string(bodyBytes)

	var mediaDAO mediadao.MediaDAO

	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"%s\", \"session\":\"%s\",\"results\":nil}\n", err, sess.SessionToken)
		return
	}

	var mediasearch mediadao.MediaSearch

	mediasearch.SearchString = search.Search

	mediaList, err := mediaDAO.MediaSearch(mediasearch)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"%s\", \"session\":\"%s\",\"results\":nil}\n", err, sess.SessionToken)
		return
	}

	mediaLength := len(mediaList)

	jsonBytes, err := json.Marshal(mediaList)
	if err != nil {
		log.Error().Err(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"query successful with %d results\", \"session\":\"%s\",\"results\":%s}\n", mediaLength, sess.SessionToken, string(jsonBytes))

}

//MediaListViewHandler List Media objects
func (ctx *HTTPHandlerContext) MediaListViewHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	template, err := util.SiteTemplate("/admin/medialist.html")
	if err != nil {
		log.Error().Err(err)
	}

	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Index",
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
	fmt.Fprint(w, out)

}

//MediaEditHandler http handler for editing media
func (ctx *HTTPHandlerContext) MediaEditHandler(w http.ResponseWriter, r *http.Request) {

	// Media Object populated from form object
	var media models.MediaModel

	// Form Management Variables
	formTitle := ""
	formTitleError := false
	formDescription := ""
	formDescriptionError := false
	formKeywords := ""
	formKeywordsError := false
	formCategory := ""
	formCategoryError := false
	formLocation := ""
	formLocationError := false

	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	// Load Media
	err := media.GetMedia(vars["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	// Test if we are a POST to capture form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		// Loading form
		media.Title = r.FormValue("title")
		media.Keywords = r.FormValue("keywords")
		media.Description = r.FormValue("description")
		media.Category = r.FormValue("category")
		media.Location = r.FormValue("location")

		// Do validation here
		validate := true
		if media.Title == "" {
			validate = false
			formTitle = "Please provide a title"
			formTitleError = true
		}

		if media.Keywords == "" {
			validate = false
			formKeywords = "Please provide keywords"
			formKeywordsError = true
		}

		if media.Description == "" {
			validate = false
			formDescription = "Please provide a description"
			formDescriptionError = true
		}

		if media.Category == "" {
			validate = false
			formCategory = "Please provide a category"
			formCategoryError = true
		}

		if media.Location == "" {
			validate = false
			formLocation = "Please provide a location"
			formLocationError = true
		}

		fmt.Println(validate)
		if validate {

			var mediaDAO mediadao.MediaDAO

			err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
			if err != nil {
				log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
			}

			// Update Record
			err = mediaDAO.UpdateMedia(media)
			if err != nil {
				fmt.Println(err)
			}

			var mediatagsDAO mediatags.MediaTagsDAO

			err = mediatagsDAO.Initialize(ctx.dbClient, ctx.hConfig)
			if err != nil {
				log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
			}

			// Update Tags
			err = mediatagsDAO.AddTagsSearchIndex(media)
			if err != nil {
				fmt.Println(err)
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, fmt.Sprintf("/admin/media/view/%s", vars["id"]), http.StatusSeeOther)
			return
		}
	}

	// HTTP Template
	template, err := util.SiteTemplate("/admin/mediaedit.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/mediaedit.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":                "Edit Media",
		"media":                media,
		"user":                 sess.User,
		"formTitle":            formTitle,
		"formTitleError":       formTitleError,
		"formKeywords":         formKeywords,
		"formKeywordsError":    formKeywordsError,
		"formDescription":      formDescription,
		"formDescriptionError": formDescriptionError,
		"formCategory":         formCategory,
		"formCategoryError":    formCategoryError,
		"formLocation":         formLocation,
		"formLocationError":    formLocationError,
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

//EditMediaAPIHandler edit media
func (ctx *HTTPHandlerContext) EditMediaAPIHandler(w http.ResponseWriter, r *http.Request) {
	//errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\"}\n"

	sess := util.GetSession(r)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"request recieved %s\"}\n", sess.SessionToken)

}

// MediaDeleteHandler Delete media from the database and s3
func (ctx *HTTPHandlerContext) MediaDeleteHandler(w http.ResponseWriter, r *http.Request) {

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	//var media models.MediaModel

	media, err := mediaDAO.GetMedia(vars["id"])
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msgf("Error getting media from database for id %s", vars["id"])
		return
	}

	simplestorageservice.DeleteS3Object(media.S3Location)

	simplestorageservice.DeleteS3Object(media.S3Thumbnail)

	simplestorageservice.DeleteS3Object(media.S3LargeView)

	simplestorageservice.DeleteS3Object(media.S3VeryLarge)

	err = mediaDAO.DeleteMedia(media)
	if err != nil {
		fmt.Println(err)
	}

	http.Redirect(w, r, "/admin/medialist", http.StatusSeeOther)
}

// PhotoViewHandler View File
func (ctx *HTTPHandlerContext) PhotoViewHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		fmt.Printf("Error getting url variable, id: %s", val)
	}

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	// Load Media
	media, err := mediaDAO.GetMediaBySlug(vars["id"])
	if err != nil {
		fmt.Println(err)
		return
	}

	template, err := util.SiteTemplate("/mediaview.html")
	if err != nil {
		log.Error().Err(err)
	}

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
	fmt.Fprint(w, out)
}

// GetMediaAPI View File
func (ctx *HTTPHandlerContext) GetMediaAPI(w http.ResponseWriter, r *http.Request) {

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

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	media, err := mediaDAO.GetMedia(vars["id"])
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
}

// ServerImageHandler proxy image requests through a handler to obfuscate
// the s3 bucket location
func (ctx *HTTPHandlerContext) ServerImageHandler(wr http.ResponseWriter, req *http.Request) {
	//log.Println(req.RemoteAddr, " ", req.Method, " ", req.URL)

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	cfg := ctx.hConfig

	slug := ""
	mediaType := ""

	// HTTP URL Parameters
	vars := mux.Vars(req)
	if val, ok := vars["slug"]; ok {
		slug = vars["slug"]
	} else {
		fmt.Printf("Error getting url variable, slug: %s\n", val)
	}

	// HTTP URL Parameters
	if val, ok := vars["type"]; ok {
		mediaType = vars["type"]
	} else {
		fmt.Printf("Error getting url variable, type: %s\n", val)
	}

	media, err := mediaDAO.GetMediaBySlug(slug)
	if err != nil {
		fmt.Printf("error getting media by slug: %s", err)
	}

	s3Path := ""

	if mediaType == "thumb" {
		s3Path = media.S3Thumbnail
	}

	if mediaType == "large" {
		s3Path = media.S3LargeView
	}

	if mediaType == "original" {
		s3Path = media.S3Location
	}

	// Generate S3 URL
	var mediaRequest http.Request
	mediaURL, err := url.Parse("https://" + cfg.GetS3Bucket() + ".s3-us-west-2.amazonaws.com" + s3Path)
	if err != nil {
		log.Printf("ServeHTTP: %s", err)
	}

	mediaRequest.URL = mediaURL

	fmt.Printf("proxy for media slug id %s for image type %s using url %s\n", slug, mediaType, mediaURL)

	// Create client
	client := &http.Client{}

	//delHopHeaders(req.Header)

	clientIP, err := requestfilter.GetIPAddress(req)
	if err != nil {
		fmt.Printf("error getting ip address in proxy to send to s3 bucke with error %s", err)
	}

	appendHostToXForwardHeader(req.Header, clientIP)

	resp, err := client.Do(&mediaRequest)
	if err != nil {

		http.Error(wr, fmt.Sprintf("error proxying url %s with error %s", mediaURL, err), http.StatusInternalServerError)
		log.Printf("error proxying url %s with error %s\n", mediaURL, err)
		return
	}

	defer resp.Body.Close()

	log.Info().Msgf("%s %s", req.RemoteAddr, resp.Status)

	delHopHeaders(resp.Header)

	copyHeader(wr.Header(), resp.Header)
	wr.WriteHeader(resp.StatusCode)
	wr.Header().Set("Content-Type", "image/jpeg") // <-- set the content-type header
	io.Copy(wr, resp.Body)
}

// ViewCategoryHandler View the media
func (ctx *HTTPHandlerContext) ViewCategoryHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	var medialist []models.MediaModel

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["category"]; ok {
		// Load Media
		medialist, err = mediaDAO.GetMediaListByCategory(vars["category"])
		if err != nil {
			log.Error().Err(err).Str("service", "mediadao").Msgf("Error getting media with variable, category %s", val)
		}
	} else {
		log.Error().Err(err).Str("service", "mediadao").Msgf("Error getting url variable, category: %s", val)
	}

	template, err := util.SiteTemplate("/category.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/mediaview.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     fmt.Sprintf("Category - %s", vars["category"]),
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
	fmt.Fprint(w, out)
}

//ViewCategoriesHandler view all categories
func (ctx *HTTPHandlerContext) ViewCategoriesHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	var mediaDAO mediadao.MediaDAO

	err := mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	medialist, err := mediaDAO.AllCategories()
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error getting media list")
	}

	template, err := util.SiteTemplate("/categories.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/mediaview.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Media Categories",
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
	fmt.Fprint(w, out)

}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func delHopHeaders(header http.Header) {
	for _, h := range hopHeaders {
		header.Del(h)
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	// If we aren't the first proxy retain prior
	// X-Forwarded-For information as a comma+space
	// separated list and fold multiple headers into one.
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ", ") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}
