package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	mediadao "github.com/rsvancara/goblog/internal/dao/media"
	_ "github.com/rsvancara/goblog/internal/filters" //import pongo  plugins
	"github.com/rsvancara/goblog/internal/models"
	simplestorageservice "github.com/rsvancara/goblog/internal/s3"
	"github.com/rsvancara/goblog/internal/session"
	"github.com/rsvancara/goblog/internal/util"
)

// MediaHandler View full list of media sorted by date
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
	fmt.Fprintf(w, out)
}

// ViewMediaHandler View the media
func (ctx *HTTPHandlerContext) ViewMediaHandler(w http.ResponseWriter, r *http.Request) {

	var media models.MediaModel

	sess := util.GetSession(r)

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

	media, err = mediaDAO.GetMedia(vars["id"])
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error retrieving data from access object ")
	}

	template, err := util.SiteTemplate("/admin/mediaview.html")
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
	fmt.Fprintf(w, out)
}

// MediaAddHandler add media
func (ctx *HTTPHandlerContext) MediaAddHandler(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r, w)
	if err != nil {
		log.Printf("Session not available %s", err)

	}

	template, err := util.SiteTemplate("/admin/mediaadd.html")
	//template := "templates/admin/mediaadd.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error loading template %s", err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// PutMedia Upload file to server
func (ctx *HTTPHandlerContext) PutMedia(w http.ResponseWriter, r *http.Request) {

	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\",\"file\":\"error\"}\n"

	vars := mux.Vars(r)
	var media models.MediaModel

	//err = r.ParseForm()
	err := r.ParseMultipartForm(128 << 20)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "parsing multipart form")
		log.Printf("Error parsing multipart form %s", err)
		return
	}

	keywords := r.FormValue("keywords")
	description := r.FormValue("description")
	title := r.FormValue("title")
	category := r.FormValue("category")
	location := r.FormValue("location")

	file, handler, err := r.FormFile("file") // Retrieve the file from form data
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		log.Printf("Error opeinging file with error  %s", err)
		fmt.Fprintf(w, errorMessage, err, "opening file")
		return
	}

	defer file.Close() // Close the file when we finish

	// This is path which we want to store the file
	f, err := os.OpenFile("temp/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "error storing file")
		return
	}

	defer f.Close()

	// Copy the file to the destination path
	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "copying file to destination path")
		return
	}

	rf, err := os.OpenFile("temp/"+handler.Filename, os.O_RDONLY, 0666)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "opening file")
		return
	}
	defer rf.Close()

	// Get exif
	err = media.ExifExtractor(rf)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "extracting exif")
		return
	}

	h := sha256.New()
	if _, err := io.Copy(h, rf); err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "creating sha265")
		return
	}
	sha256 := hex.EncodeToString(h.Sum(nil))

	media.Keywords = keywords
	media.Checksum = string(sha256)
	media.Description = description
	media.Category = category
	media.FileName = handler.Filename
	media.Title = title
	media.Location = location
	media.S3Uploaded = "false"

	var mediaDAO mediadao.MediaDAO

	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	err = mediaDAO.InsertMedia(&media)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error retrieving data from access object ")
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "Inserting media into database")
		return
	}

	// Update Tags
	err = ctx.AddTagsSearchIndex(media)
	if err != nil {
		fmt.Println(err)
	}

	// Get s3 key
	simplestorageservice.S3KeyGenerator(&media)

	// Launch in a go routine so it is non blocking
	go simplestorageservice.AddFileToS3("temp/"+handler.Filename, &media, ctx.dbClient, ctx.hConfig)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"file %s uploaded\",\"file\":\"%s\"}\n", vars["id"], handler.Filename)
	return
}

//AddTagsSearchIndex When images are created, add tags to the tags index
//TODO: Convert to DAO
func (ctx *HTTPHandlerContext) AddTagsSearchIndex(media models.MediaModel) error {

	for _, v := range media.Tags {
		var mtm models.MediaTagsModel
		count, err := mtm.Exists(v.Keyword)
		if err != nil {
			return fmt.Errorf("Error attempting to get record count for keyword %s with error %s", v.Keyword, err)
		}

		fmt.Printf("Found %v media tag records\n", count)

		// Determine if the document exists already
		if count == 0 {
			var newMTM models.MediaTagsModel
			newMTM.Name = v.Keyword
			newMTM.TagsID = models.GenUUID()
			var docs []string
			docs = append(docs, media.MediaID)
			newMTM.Documents = docs
			fmt.Println(newMTM)
			err = newMTM.InsertMediaTags()
			if err != nil {
				return fmt.Errorf("Error inserting new media tag for keyword %s with error %s", v.Keyword, err)
			}
			// If not, then we add to existing documents
		} else {
			var mtm models.MediaTagsModel
			err := mtm.GetMediaTagByName(v.Keyword)
			if err != nil {
				return fmt.Errorf("Error getting current instance of mediatag for keyword %s with error %s", v.Keyword, err)
			}
			fmt.Printf("Found existing mediatag record for %s", mtm.Name)
			fmt.Println(mtm.Documents)

			// Get the list of documents
			docs := mtm.Documents

			// For the list of documents, find the document ID we are looking for
			// If not found, then we update the document list with the document ID
			f := 0
			for _, d := range docs {
				if d == media.MediaID {
					f = 1
				}
			}

			if f == 0 {
				fmt.Printf("Updating tag, %s with document id %s\n", v.Keyword, media.MediaID)
				docs = append(docs, media.MediaID)
				mtm.Documents = docs
				fmt.Println(mtm)
				err = mtm.UpdateMediaTags()
				if err != nil {
					return fmt.Errorf("Error updating mediatag for keyword %s with error %s", v.Keyword, err)
				}
			}
		}
	}
	return nil
}

//MediaSearchAPIHandler search by media tags
func (ctx *HTTPHandlerContext) MediaSearchAPIHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	bodyString := "{}"
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"Error: %s\", \"session\":\"%s\",\"results\":nil}\n", err, sess.SessionToken)
		return
	}
	bodyString = string(bodyBytes)

	var mediaDAO mediadao.MediaDAO

	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	mediaList, err := mediaDAO.MediaSearch(bodyString)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"status\":\"error\", \"message\": \"%s\", \"session\":\"%s\",\"results\":nil}\n", err, sess.SessionToken)
		return
	}

	mediaLength := len(mediaList)

	jsonBytes, err := json.Marshal(mediaList)

	//fmt.Println(string(jsonBytes))

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"query successful with %d results\", \"session\":\"%s\",\"results\":%s}\n", mediaLength, sess.SessionToken, string(jsonBytes))
	return
}

//MediaListViewHandler List Media objects
func (ctx *HTTPHandlerContext) MediaListViewHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	template, err := util.SiteTemplate("/admin/medialist.html")
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
	fmt.Fprintf(w, out)

}

//EditMediaAPIHandler edit media
func (ctx *HTTPHandlerContext) EditMediaAPIHandler(w http.ResponseWriter, r *http.Request) {
	//errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\"}\n"

	sess := util.GetSession(r)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"request recieved %s\"}\n", sess.SessionToken)
	return
}

// MediaDeleteHandler Delete media from the database and s3
func (ctx *HTTPHandlerContext) MediaDeleteHandler(w http.ResponseWriter, r *http.Request) {

	//sess := util.GetSession(r)

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
		fmt.Println(err)
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

	http.Redirect(w, r, "/admin/media", http.StatusSeeOther)
}
