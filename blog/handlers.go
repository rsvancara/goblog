package blog

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	//"bf.go/blog/mongo"
	//"bf.go/blog/models"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"

	"bf.go/blog/models"
	"bf.go/blog/session"
	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
)

// ContextKey key used by contexts to uniquely identify attributes
type ContextKey string

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

// AuthHandler authorize user
func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var sess session.Session

		err := sess.Session(r)

		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// HomeHandler Home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("No session cookie found: %s \n", err)
		//http.Redirect(w, r, "/signin", http.StatusSeeOther)
		//return
	}

	template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// Signin Sign into the application
func Signin(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		var creds session.Credentials
		creds.Username = r.FormValue("username")
		creds.Password = r.FormValue("password")

		// Get the expected password from our in memory map
		expectedPassword, ok := users[creds.Username]

		// If a password exists for the given user
		// AND, if it is the same as the password we received, the we can move ahead
		// if NOT, then we return an "Unauthorized" status
		if !ok || expectedPassword != creds.Password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var sess session.Session

		err := sess.Create(creds)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println(err)
		}

		// Finally, we set the client cookie for "session_token" as the session token we just generated
		// we also set an expiry time of 120 seconds, the same as the cache
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sess.SessionToken,
			Expires: time.Now().Add(1800 * time.Second),
		})

		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	template := "templates/signin.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)

}

// AdminHome admin home page
func AdminHome(w http.ResponseWriter, r *http.Request) {

	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}
	template := "templates/admin/admin.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
	// w.Write([]byte(fmt.Sprintf("Welcome %s!", sess.User.Username)))
}

// AboutHandler about page
func AboutHandler(w http.ResponseWriter, r *http.Request) {

	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template := "templates/about.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// Post post
func Post(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not availab le %s\n", err)
	}

	// Create Record
	posts, err := models.AllPostsSortedByDate()
	if err != nil {
		fmt.Println(err)
	}

	template := "templates/admin/post.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "posts": posts, "user": sess.User.Username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// Media media
func Media(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s\n", err)
	}

	template := "templates/admin/media.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
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

	//http Session
	var sess session.Session
	err := sess.Session(r)
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
	var pm models.PostModel

	// Load Model
	pm.GetPost(vars["id"])

	// Test if we are a POST to capture form submission
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		pm.Post = r.FormValue("inputPost")
		pm.Title = r.FormValue("inputTitle")
		pm.Status = r.FormValue("inputStatus")
		if err != nil {
			fmt.Printf("Error converting status to integer in post form: %s\n", err)
		}
		//pm.Keywords = r.FormValue("")

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

		if pm.Status == "enabled" || pm.Status == "disabled" {

		} else {
			statusMessage = "Invalid status code"
			statusMessageError = true
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
	template := "templates/admin/postedit.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":              "Edit Post",
		"post":               pm,
		"user":               sess.User.Username,
		"postMessage":        postMessage,
		"postMessageError":   postMessageError,
		"titleMessage":       titleMessage,
		"titleMessageError":  titleMessageError,
		"statusMessage":      statusMessage,
		"statusMessageError": statusMessageError,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// PostView edit the post
func PostView(w http.ResponseWriter, r *http.Request) {

	//http Session
	var sess session.Session
	err := sess.Session(r)
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

	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.TOC
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	extensions := parser.CommonExtensions | parser.Tables | parser.SpaceHeadings
	parser := parser.NewWithExtensions(extensions)
	// This library does not deal well with \r\n that most browswers use, r
	// replace the \r\n with \n
	md := []byte(pm.Post)
	md = NormalizeNewlines(md)
	content := string(markdown.ToHTML(md, parser, renderer))

	// HTTP Template
	template := "templates/admin/postview.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":   "Edit Post",
		"post":    pm,
		"content": content,
		"user":    sess.User.Username,
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

	var pm models.PostModel
	// Form Variables
	titleMessage := ""
	titleMessageError := false
	postMessage := ""
	postMessageError := false
	statusMessage := ""
	statusMessageError := false

	// HTTP Session
	var sess session.Session
	err := sess.Session(r)
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
		if err != nil {
			fmt.Printf("Error converting status to integer in post form: %s\n", err)
		}
		//pm.Keywords = r.FormValue("")

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

		if pm.Status == "enabled" || pm.Status == "disabled" {

		} else {
			statusMessage = "Invalid status code"
			statusMessageError = true
		}

		if validate == true {

			// Create Record
			err = pm.InsertPost()
			if err != nil {
				fmt.Println(err)
			}

			// Redirect on success otherwise fall through the form
			// and display any errors
			http.Redirect(w, r, "/admin/post", http.StatusSeeOther)
			return
		}
	}

	template := "templates/admin/postadd.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":              "Add Post",
		"post":               pm,
		"user":               sess.User.Username,
		"postMessage":        postMessage,
		"postMessageError":   postMessageError,
		"titleMessage":       titleMessage,
		"titleMessageError":  titleMessageError,
		"statusMessage":      statusMessage,
		"statusMessageError": statusMessageError,
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

	//http Session
	var sess session.Session
	err := sess.Session(r)
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
	var pm models.PostModel

	// Load Model
	pm.GetPost(vars["id"])

	pm.DeletePost()

	http.Redirect(w, r, "/admin/post", http.StatusSeeOther)
}

// MediaAdd add media
func MediaAdd(w http.ResponseWriter, r *http.Request) {
	var sess session.Session
	err := sess.Session(r)
	if err != nil {
		fmt.Printf("Session not available %s", err)
	}

	template := "templates/about.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello", "user": sess.User.Username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
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
