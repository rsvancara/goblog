//Package handlers Search Index handler
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	mediadao "goblog/internal/dao/media"
	mediatagsdao "goblog/internal/dao/mediatags"
	postdao "goblog/internal/dao/posts"
	sitetagsdao "goblog/internal/dao/sitesearchtags"
	"goblog/internal/models"
	"goblog/internal/util"

	"github.com/rs/zerolog/log"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

// SearchIndexListHandler Main Search management page
func (ctx *HTTPHandlerContext) SearchIndexListHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	template, err := util.SiteTemplate("/admin/searchindex.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	var sm sitetagsdao.SiteSearchTagsDAO
	var mm mediatagsdao.MediaTagsDAO

	err = mm.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err)
	}

	mediacount, err := mm.GetMediaTagsCount()
	if err != nil {
		log.Error().Err(err).Msg("error retrieving media tags count ")
	}

	err = sm.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err)
	}
	sitecount, err := sm.GetSiteSearchTagsCount()
	if err != nil {
		log.Error().Err(err)
	}

	out, err := tmpl.Execute(pongo2.Context{
		"title":     "Search Inex",
		"user":      sess.User,
		"bodyclass": "",
		"hidetitle": true,
		"mediatags": mediacount,
		"sitecount": sitecount,
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

// Site Search Handler - Display site search results
func (ctx *HTTPHandlerContext) SiteSearchHandler(w http.ResponseWriter, r *http.Request) {

	template, err := util.SiteTemplate("/sitesearchresult.html")
	if err != nil {
		log.Error().Err(err)
	}

	var siteSearchTagsDAO sitetagsdao.SiteSearchTagsDAO
	err = siteSearchTagsDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msg("error initialzing siteSearchTagsDAO data access object ")
	}

	var md mediadao.MediaDAO
	err = md.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "MediaDAO").Msg("error initialzing siteSearchTagsDAO data access object ")
	}

	var pd postdao.PostsDAO
	err = pd.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "MediaDAO").Msg("error initialzing siteSearchTagsDAO data access object ")
	}

	type SearchResult struct {
		Title       string
		DocType     string
		Slug        string
		KeyWords    string
		Description string
	}

	var searchResults []SearchResult

	ssmTags, err := siteSearchTagsDAO.SearchMediaTagsByName("landscape")
	if err != nil {
		log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msg("error initialzing siteSearchTagsDAO data access object ")
	}

	for _, v := range ssmTags {
		for _, k := range v.Documents {

			// Query the media database
			if k.DocType == "media" {
				var sr SearchResult

				media, err := md.GetMedia(k.DocumentID)
				if err != nil {
					log.Error().Err(err)
				}

				sr.Description = media.Description
				sr.DocType = "media"
				sr.Slug = media.Slug
				sr.KeyWords = media.Keywords
				sr.Title = media.Title

				searchResults = append(searchResults, sr)

			}

			// Query the post database
			if k.DocType == "post" {

				var sr SearchResult

				post, err := pd.GetPost(k.DocumentID)
				if err != nil {
					log.Error().Err(err)
				}

				sr.Description = post.PostTeaser
				sr.DocType = "post"
				sr.Slug = post.Slug
				sr.KeyWords = post.Keywords
				sr.Title = post.Title

				searchResults = append(searchResults, sr)

			}

		}
	}

	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":         "Search Inex",
		"bodyclass":     "",
		"hidetitle":     true,
		"pagekey":       util.GetPageID(r),
		"searchresults": searchResults,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)
}

func (ctx *HTTPHandlerContext) SiteSearchIndexBuildTagsHandler(w http.ResponseWriter, r *http.Request) {

	log.Info().Msg("Initializing site search dao")
	var siteSearchTagsDAO sitetagsdao.SiteSearchTagsDAO
	err := siteSearchTagsDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msg("error initialzing siteSearchTagsDAO data access object ")
	}

	log.Info().Msg("Initializing mediadao")
	var mediaDAO mediadao.MediaDAO
	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("error initialzing media data access object ")
	}

	log.Info().Msg("Initializing postdao")
	var postdao postdao.PostsDAO
	err = postdao.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("error initialzing media data access object ")
	}

	log.Info().Msg("Clearing out all the tags")

	// Clear the Index
	err = siteSearchTagsDAO.DeleteAllTags()
	if err != nil {
		log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msg("Error deleting siteSearchTagsDAO while trying to build index")
	}

	log.Info().Msg("Gettng a list of all media")
	// Get a list of media
	media, err := mediaDAO.AllMediaSortedByDate()
	if err != nil {
		log.Error().Err(err).Str("service", "mediaDAO").Msg("Error getting all media in database")
	}

	// Iterate through the media and update the index
	for _, v := range media {

		for _, t := range v.Tags {
			// Check to see if the key word exists and if it does not, add it
			// If it does then update the keyword with the list of new documents
			fmt.Printf("Looking at tag %s\n", t)
			// Check if the document exists
			var stm models.SiteSearchTagsModel

			count, err := siteSearchTagsDAO.Exists(t.Keyword)
			if err != nil {
				log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msgf("Error attempting to get record count for keyword %s", t.Keyword)
			}
			if count == 0 {
				fmt.Printf("Adding new tag %s\n", t.Keyword)
				stm.Name = t.Keyword
				stm.TagsID = models.GenUUID()
				var docs []models.Documents
				var doc models.Documents
				doc.DocType = "media"
				doc.DocumentID = v.MediaID
				docs = append(docs, doc)
				stm.Documents = docs
				err = siteSearchTagsDAO.InsertSiteSearchTags(&stm)
				if err != nil {
					log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msgf("Error inserting record for keyword %s", t.Keyword)
				}
			} else {

				stm, err := siteSearchTagsDAO.GetSiteSearchTagByName(t.Keyword)
				if err != nil {

					log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msgf("Error getting record for keyword %s", t.Keyword)
				}

				fmt.Printf("found existing tag %s\n", stm.Name)

				// Get the list of documents
				docs := stm.Documents

				// For the list of documents, find the document ID we are looking for
				// If not found, then we update the document list with the document ID
				f := false
				for _, d := range docs {
					if d.DocumentID == v.MediaID {
						f = true
					}
				}
				// If not found then update
				if !f {
					fmt.Printf("Updating tag, %s with document id %s\n", t.Keyword, v.MediaID)
					var doc models.Documents
					doc.DocType = "media"
					doc.DocumentID = v.MediaID

					docs = append(docs, doc)
					stm.Documents = docs
					err = siteSearchTagsDAO.UpdateSiteSearchTags(&stm)
					if err != nil {
						log.Error().Err(err).Str("service", "mediatagsDAO").Msgf("Error updating record for keyword %s", t.Keyword)
					}
				}
			}
		}
	}

	log.Info().Msg("Gettng a list of all posts")

	// Get a list of media
	posts, err := postdao.AllPostsSortedByDate()
	if err != nil {
		log.Error().Err(err).Str("service", "postDAO").Msg("error getting all posts in database")
	}

	// Iterate through the media and update the index
	for _, p := range posts {

		for _, t := range p.Tags {
			// Check to see if the key word exists and if it does not, add it
			// If it does then update the keyword with the list of new documents
			fmt.Printf("Looking at tag %s\n", t)
			// Check if the document exists
			var stm models.SiteSearchTagsModel

			count, err := siteSearchTagsDAO.Exists(t.Keyword)
			if err != nil {
				log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msgf("Error attempting to get record count for keyword %s", t.Keyword)
			}
			if count == 0 {
				fmt.Printf("Adding new tag %s\n", t.Keyword)
				stm.Name = t.Keyword
				stm.TagsID = models.GenUUID()
				var docs []models.Documents
				var doc models.Documents
				doc.DocType = "post"
				doc.DocumentID = p.PostID
				docs = append(docs, doc)
				stm.Documents = docs
				err = siteSearchTagsDAO.InsertSiteSearchTags(&stm)
				if err != nil {
					log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msgf("Error inserting record for keyword %s", t.Keyword)
				}
			} else {

				stm, err := siteSearchTagsDAO.GetSiteSearchTagByName(t.Keyword)
				if err != nil {

					log.Error().Err(err).Str("service", "siteSearchTagsDAO").Msgf("Error getting record for keyword %s", t.Keyword)
				}

				fmt.Printf("found existing tag %s\n", stm.Name)

				// Get the list of documents
				docs := stm.Documents

				// For the list of documents, find the document ID we are looking for
				// If not found, then we update the document list with the document ID
				f := false
				for _, d := range docs {
					if d.DocumentID == p.PostID {
						f = true
					}
				}
				// If not found then update
				if !f {
					fmt.Printf("Updating tag, %s with document id %s\n", t.Keyword, p.PostID)
					var doc models.Documents
					doc.DocType = "post"
					doc.DocumentID = p.PostID

					docs = append(docs, doc)
					stm.Documents = docs
					err = siteSearchTagsDAO.UpdateSiteSearchTags(&stm)
					if err != nil {
						log.Error().Err(err).Str("service", "mediatagsDAO").Msgf("Error updating record for keyword %s", t.Keyword)
					}
				}
			}
		}
	}

	// Redirect back to the search page
	http.Redirect(w, r, "/admin/searchindex", http.StatusSeeOther)
}

// SearchIndexBuildTagsHandler Build tags search index.
func (ctx *HTTPHandlerContext) SearchIndexBuildTagsHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	//var sessions []sessionmanager.Session
	//sessions, err := sessionmanager.GetAllSessions(*ctx.cache, ctx.hConfig.RedisDB, "*")
	//if err != nil {
	//	log.Error().Err(err)
	//}

	template, err := util.SiteTemplate("/admin/buildsearchindex.html")
	if err != nil {
		log.Error().Err(err)
	}

	//template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title": "Search Inex",
		//"sessions":  sessions,
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
					//w.WriteHeader(http.StatusOK)
					//w.Header().Set("Content-Type", "application/json")
					//fmt.Fprintf(w, errorMessage, err, "could not get media object from database")
					log.Error().Err(err).Msgf("error getting media for id %s", d)
				} else {
					var doc document
					doc.DocumentID = d
					doc.SmallImageURL = fmt.Sprintf("image/%s/thumb", m.Slug)
					doc.Title = m.Title
					docs = append(docs, doc)
				}
			}
		}
	}

	b, err := json.Marshal(docs)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, errorMessage, err, "could not get mediatag object from database")
		log.Error().Err(err)
		return
	}

	fmt.Printf("{\"status\":\"success\", \"message\": \"mediatag found\",\"tags\":%s}\n", string(b))

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
