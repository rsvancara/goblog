//Package handlers Search Index handler
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	mediadao "goblog/internal/dao/media"
	mediatagsdao "goblog/internal/dao/mediatags"
	"goblog/internal/models"
	"goblog/internal/session"
	"goblog/internal/util"

	"github.com/rs/zerolog/log"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

// SearchIndexListHandler Main Search management page
func (ctx *HTTPHandlerContext) SearchIndexListHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	var sessions []session.Session
	sessions, err := session.GetAllSessions(*ctx.cachePool, *ctx.hConfig)
	if err != nil {
		log.Error().Err(err)
	}

	template, err := util.SiteTemplate("/admin/searchindex.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	// Check if the document exists
	var mtm models.MediaTagsModel

	var count int64

	count, err = mtm.GetMediaTagsCount()
	if err != nil {
		count = 0
		fmt.Printf("error retrieving media tags count with error %s\n", err)
	}

	//fmt.Println(count)

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Search Inex",
		"sessions":  sessions,
		"user":      sess.User,
		"bodyclass": "",
		"hidetitle": true,
		"mediatags": count,
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

// SearchIndexBuildTagsHandler Build tags search index.
func (ctx *HTTPHandlerContext) SearchIndexBuildTagsHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	var sessions []session.Session
	sessions, err := session.GetAllSessions(*ctx.cachePool, *ctx.hConfig)
	if err != nil {
		log.Error().Err(err)
	}

	template, err := util.SiteTemplate("/admin/buildsearchindex.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Search Inex",
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
	fmt.Fprint(w, out)
}

// SearchIndexMediaTagsAPI Search for media tags via an API
func (ctx *HTTPHandlerContext) SearchIndexMediaTagsAPI(w http.ResponseWriter, r *http.Request) {
	errorMessage := "{\"status\":\"error\", \"message\": \"error: %s in %s\"}\n"

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if _, ok := vars["name"]; ok {

	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, "Error find NAME", "NAME was not available in the URL or could not be parsed")
		return
	}

	mtms, err := models.SearchMediaTagsByName(vars["name"])
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "could not get mediatag object from database")
		return
	}

	type document struct {
		DocumentID    string `json:"document_id"`
		SmallImageURL string `json:"small_image_url"`
		Title         string `json:"title"`
	}

	var docs []document

	for _, v := range mtms {
		for _, d := range v.Documents {
			f := false
			for _, x := range docs {
				if x.DocumentID == d {
					f = true
				}
			}
			if !f {
				var m models.MediaModel

				err := m.GetMedia(d)
				if err != nil {
					w.WriteHeader(http.StatusOK)
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintf(w, errorMessage, err, "could not get media object from database")
				}
				var doc document
				doc.DocumentID = d
				doc.SmallImageURL = fmt.Sprintf("image/%s/thumb", m.Slug)
				doc.Title = m.Title
				docs = append(docs, doc)
			}
		}
	}

	b, err := json.Marshal(docs)
	if err != nil {
		log.Error().Err(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"mediatag found\",\"tags\":%s}\n", string(b))
}

// SearchIndexBuilderMediaHandler builds the tag index for media
func (ctx *HTTPHandlerContext) SearchIndexBuilderMediaHandler(w http.ResponseWriter, r *http.Request) {

	var mediatagsDAO mediatagsdao.MediaTagsDAO
	err := mediatagsDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediatagsDAO").Msg("Error initialzing mediatags data access object ")
	}

	var mediaDAO mediadao.MediaDAO
	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	// Clear the Index
	err = mediatagsDAO.DeleteAllTags()
	if err != nil {
		log.Error().Err(err).Str("service", "mediatagsDAO").Msg("Error deleting mediatags while trying to build index")
	}

	// Get a list of media
	media, err := mediaDAO.AllMediaSortedByDate()
	if err != nil {
		log.Error().Err(err).Str("service", "mediatagsDAO").Msg("Error getting all media in database")
	}

	// Iterate through the media and update the index
	for _, v := range media {

		for _, t := range v.Tags {
			// Check to see if the key word exists and if it does not, add it
			// If it does then update the keyword with the list of new documents
			fmt.Printf("Looking at tag %s\n", t)
			// Check if the document exists
			var mtm models.MediaTagsModel

			count, err := mediatagsDAO.Exists(t.Keyword)
			if err != nil {
				log.Error().Err(err).Str("service", "mediatagsDAO").Msgf("Error attempting to get record count for keyword %s", t.Keyword)
			}
			if count == 0 {
				fmt.Printf("Adding new tag %s\n", t.Keyword)
				mtm.Name = t.Keyword
				mtm.TagsID = models.GenUUID()
				var docs []string
				docs = append(docs, v.MediaID)
				mtm.Documents = docs
				err = mediatagsDAO.InsertMediaTags(&mtm)
				if err != nil {
					log.Error().Err(err).Str("service", "mediatagsDAO").Msgf("Error inserting record for keyword %s", t.Keyword)
				}
			} else {

				mtm, err := mediatagsDAO.GetMediaTagByName(t.Keyword)
				if err != nil {

					log.Error().Err(err).Str("service", "mediatagsDAO").Msgf("Error getting record for keyword %s", t.Keyword)
				}

				fmt.Printf("found existing tag %s\n", mtm.Name)

				// Get the list of documents
				docs := mtm.Documents

				// For the list of documents, find the document ID we are looking for
				// If not found, then we update the document list with the document ID
				f := 0
				for _, d := range docs {
					if d == v.MediaID {
						f = 1
					}
				}
				// If not found then update
				if f == 0 {
					fmt.Printf("Updating tag, %s with document id %s\n", t.Keyword, v.MediaID)

					docs = append(docs, v.MediaID)
					mtm.Documents = docs
					err = mediatagsDAO.UpdateMediaTags(&mtm)
					if err != nil {
						log.Error().Err(err).Str("service", "mediatagsDAO").Msgf("Error updating record for keyword %s", t.Keyword)
					}
				}
			}
		}
	}

	http.Redirect(w, r, "/admin/searchindex", http.StatusSeeOther)
}
