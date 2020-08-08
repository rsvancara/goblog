package views

import (
	"blog/blog/models"
	"blog/blog/session"
	"blog/blog/util"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

// SearchIndexListHandler Main Search management page
func SearchIndexListHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	var sessions []session.Session
	sessions, err := session.GetAllSessions()

	template, err := util.SiteTemplate("/admin/searchindex.html")
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

	fmt.Println(count)

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
	fmt.Fprintf(w, out)

	return
}

// SearchIndexBuildTagsHandler Build tags search index.
func SearchIndexBuildTagsHandler(w http.ResponseWriter, r *http.Request) {

	sess := util.GetSession(r)

	var sessions []session.Session
	sessions, err := session.GetAllSessions()

	template, err := util.SiteTemplate("/admin/buildsearchindex.html")
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
	fmt.Fprintf(w, out)

	return
}

// SearchIndexMediaTagsAPI Search for media tags via an API
func SearchIndexMediaTagsAPI(w http.ResponseWriter, r *http.Request) {
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
			if f == false {
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"status\":\"success\", \"message\": \"mediatag found\",\"tags\":%s}\n", string(b))

	return
}

// SearchIndexBuilderMediaHandler builds the tag index for media
func SearchIndexBuilderMediaHandler(w http.ResponseWriter, r *http.Request) {

	BuildTagsSearchIndex()

	http.Redirect(w, r, "/admin/searchindex", http.StatusSeeOther)
}

// BuildTagsSearchIndex Build search index
func BuildTagsSearchIndex() error {

	// Get a list of media
	media, err := models.AllMediaSortedByDate()
	if err != nil {
		fmt.Printf("error getting list of media with error %s\n", err)
	}

	// Iterate through the media and update the index
	for _, v := range media {

		for _, t := range v.Tags {
			// Check to see if the key word exists and if it does not, add it
			// If it does then update the keyword with the list of new documents
			fmt.Printf("Looking at tag %s\n", t)
			// Check if the document exists
			var mtm models.MediaTagsModel

			count, err := mtm.Exists(t.Keyword)
			if err != nil {
				fmt.Printf("Error attempting to get record count for keyword %s with error %s", t.Keyword, err)
			}
			if count == 0 {
				fmt.Printf("Adding new tag %s\n", t.Keyword)
				mtm.Name = t.Keyword
				mtm.TagsID = models.GenUUID()
				var docs []string
				docs = append(docs, v.MediaID)
				mtm.Documents = docs
				err = mtm.InsertMediaTags()
				if err != nil {
					fmt.Printf("error inserting media for name %s with error %s\n", t.Keyword, err)
				}
			} else {

				err := mtm.GetMediaTagByName(t.Keyword)
				if err != nil {
					fmt.Printf("Error attempting to find keyword %s with error %s\n", t.Keyword, err)
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
					err = mtm.UpdateMediaTags()
					if err != nil {
						fmt.Printf("error updating media tags for %s with error %s\n", t.Keyword, err)
					}
				}
			}
		}
	}
	return nil
}
