package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	mediadao "goblog/internal/dao/media"
	postsdao "goblog/internal/dao/posts"
	"goblog/internal/models"
	"goblog/internal/sessionmanager"
	"goblog/internal/util"

	"github.com/rs/zerolog/log"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

// HomeHandler Displays the home page with list of posts
func (ctx *HTTPHandlerContext) HomeHandler(w http.ResponseWriter, r *http.Request) {

	var page int64 = 1
	var err error

	// HTTP URL Parameters
	page, err = strconv.ParseInt(r.URL.Query().Get("page"), 0, 64)
	if err != nil {
		log.Info().Msg("Page is not available")
	}

	if page == 0 {
		page = 1
	}

	sess := util.GetSession(r)

	var postDAO postsdao.PostsDAO
	err = postDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Get List
	posts, pageCount, hasNextPage, hasPrevPage, err := postDAO.AllPostsSortedByDatePaginated(page, 10)
	//posts, err := postDAO.AllPostsSortedByDate()
	if err != nil {
		fmt.Println(err)
	}

	template, err := util.SiteTemplate("/index.html")
	if err != nil {
		log.Error().Err(err)
	}

	//pageCount := len(posts)

	//template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":       "Index",
		"posts":       posts,
		"pagecount":   pageCount,
		"currentpage": page,
		"nextpage":    page + 1,
		"prevpage":    page - 1,
		"hasnextpage": hasNextPage,
		"hasprevpage": hasPrevPage,
		"user":        sess.User,
		"bodyclass":   "frontpage",
		"hidetitle":   true,
		"pagekey":     util.GetPageID(r),
		"token":       sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)
}

// PostViewHandler View individual post
func (ctx *HTTPHandlerContext) PostViewHandler(w http.ResponseWriter, r *http.Request) {

	var page int64 = 1
	var err error

	// HTTP URL Parameters
	page, err = strconv.ParseInt(r.URL.Query().Get("page"), 0, 64)
	if err != nil {
		log.Info().Msg("Page is not available")
	}

	if page == 0 {
		page = 1
	}

	sess := util.GetSession(r)

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {

	} else {
		log.Error().Msgf("Error getting url variable, id: %s", val)
	}

	var postDAO postsdao.PostsDAO
	err = postDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	pm, err := postDAO.GetPostBySlug(vars["id"])
	if err != nil {
		log.Error().Err(err).Msgf("Error getting post object from database: %s", err)

		template, err := util.SiteTemplate("/postnotfound.html")
		if err != nil {
			log.Error().Err(err)
		}
		//template := "templates/signin.html"
		tmpl := pongo2.Must(pongo2.FromFile(template))

		out, err := tmpl.Execute(pongo2.Context{
			"title":   fmt.Sprintf("Post Not Found for %s", vars["id"]),
			"user":    sess.User,
			"pagekey": util.GetPageID(r),
			"token":   sess.SessionToken,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal Error")
		}

		w.WriteHeader(http.StatusNotFound)

		fmt.Fprint(w, out)

		return

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
		log.Error().Err(err).Msgf("Error rendering markdown: %s", err)
	}

	template, err := util.SiteTemplate("/post.html")
	if err != nil {
		log.Error().Err(err)
	}

	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "Index",
		"post":    pm,
		"content": buf.String(),
		"user":    sess.User,
		"pagekey": util.GetPageID(r),
		"page":    page,
		"token":   sess.SessionToken,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, out)
}

// PostHandler View Post
func (ctx *HTTPHandlerContext) PostHandler(w http.ResponseWriter, r *http.Request) {
	sess := util.GetSession(r)

	var postDAO postsdao.PostsDAO
	err := postDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Create Record
	posts, err := postDAO.AllPostsSortedByDate()
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error retriving all posts sorted by date")
	}

	template, err := util.SiteTemplate("/admin/post.html")
	if err != nil {
		log.Error().Err(err)
	}

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
	fmt.Fprint(w, out)
}

// PostEditHandler admin edit the post
func (ctx *HTTPHandlerContext) PostEditHandler(w http.ResponseWriter, r *http.Request) {

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

	var postDAO postsdao.PostsDAO
	err := postDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	var mediaDAO mediadao.MediaDAO
	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
	}

	// Load Model
	pm, err := postDAO.GetPost(vars["id"])
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msgf("error getting post from database for id %s", vars["id"])
		// TODO: Need to send user to page that says model can not be found...bla bla
	}

	var teaserImageURL string
	// Get image URL if teaser image is present
	if pm.TeaserImage != "" {

		mm, err := mediaDAO.GetMedia(pm.TeaserImage)
		if err != nil {
			log.Error().Err(err).Str("service", "mediadao").Msgf("error getting image from database for id %s", pm.TeaserImage)
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

		if validate {

			// Create Record
			err = postDAO.UpdatePost(&pm)
			if err != nil {
				log.Error().Err(err).Str("service", "postdao").Msg("Error updating post")
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, "/admin/post", http.StatusSeeOther)
			return
		}
	}

	// HTTP Template
	template, err := util.SiteTemplate("/admin/postedit.html")
	if err != nil {
		log.Error().Err(err)
	}

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
	fmt.Fprint(w, out)
}

// PostAdminViewHandler view the post in the admin view
func (ctx *HTTPHandlerContext) PostAdminViewHandler(w http.ResponseWriter, r *http.Request) {

	//http Session
	var sess sessionmanager.Session
	err := sess.Session(*ctx.cache, ctx.hConfig.RedisDB, r, w)
	if err != nil {
		log.Error().Err(err).Str("service", "session").Msg("Session not available")
	}

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		log.Error().Err(err).Str("service", "postdao").Msgf("Error no id was provided for post: %s", val)
		//TODO: Redirect somewhere like an error page
	}

	var postDAO postsdao.PostsDAO
	err = postDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Load Model
	pm, err := postDAO.GetPost(vars["id"])
	if err != nil {
		log.Error().Err(err).Str("service", "session").Msgf("Error getting post for postid %s", vars["id"])
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
		log.Error().Err(err).Str("service", "markdown").Msg("Error rendering markdown ")
	}

	// HTTP Template
	template, err := util.SiteTemplate("/admin/postview.html")
	if err != nil {
		log.Error().Err(err)
	}

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
	fmt.Fprint(w, out)
}

// PostAddHandler add post
func (ctx *HTTPHandlerContext) PostAddHandler(w http.ResponseWriter, r *http.Request) {

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
	var sess sessionmanager.Session
	err := sess.Session(*ctx.cache, ctx.hConfig.RedisDB, r, w)
	if err != nil {
		log.Error().Err(err).Str("service", "session").Msg("Session not available")
	}

	var postDAO postsdao.PostsDAO
	err = postDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	var mediaDAO mediadao.MediaDAO
	err = mediaDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "mediadao").Msg("Error initialzing media data access object ")
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

		if pm.TeaserImage != "" {
			mm, err := mediaDAO.GetMedia(pm.TeaserImage)
			if err != nil {
				log.Error().Err(err).Str("service", "mediadao").Msg("Error retrieving media by tag ")
			}
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

		if validate {

			// Create Record
			err = postDAO.InsertPost(&pm)
			if err != nil {
				log.Error().Err(err).Str("service", "postdao").Msg("error inserting post")
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, "/admin/post", http.StatusSeeOther)
			return
		}
	}

	template, err := util.SiteTemplate("/admin/postadd.html")
	if err != nil {
		log.Error().Err(err)
	}

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
	fmt.Fprint(w, out)
}

// PostDeleteHandler delete post
func (ctx *HTTPHandlerContext) PostDeleteHandler(w http.ResponseWriter, r *http.Request) {

	// HTTP URL Parameters
	vars := mux.Vars(r)
	if val, ok := vars["id"]; ok {
		//do something here
		fmt.Println(val)
	}

	var postDAO postsdao.PostsDAO
	err := postDAO.Initialize(ctx.dbClient, ctx.hConfig)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error initialzing post data access object ")
	}

	// Load Model
	pm, err := postDAO.GetPost(vars["id"])
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msgf("Error retrieving post for id %s ", vars["id"])
	}

	err = postDAO.DeletePost(&pm)
	if err != nil {
		log.Error().Err(err).Str("service", "postdao").Msg("Error deleting post")
	}

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
