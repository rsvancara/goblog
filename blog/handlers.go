package blog

import (
	"context"
	"fmt"
	"net/http"
	"time"

	//"bf.go/blog/mongo"
	//"bf.go/blog/models"
	"bf.go/blog/db"
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
	template := "templates/admin.html"
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

	template := "templates/post.html"
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

	template := "templates/media.html"
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

	// Database
	var ms db.Session
	var dbconfig db.Config
	dbconfig.DBUri = "mongodb://host.docker.internal:27017"
	err = ms.NewSession(&dbconfig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	defer ms.Close()

	err = ms.Client.Ping(context.TODO(), nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}

	// Model
	var pm models.PostModel

	// Load Model
	pm.GetPost(vars["id"])

	// HTTP Template
	template := "templates/postadd.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":             "Add Post",
		"post":              pm,
		"user":              sess.User.Username,
		"postMessage":       postMessage,
		"postMessageError":  postMessageError,
		"titleMessage":      titleMessage,
		"titleMessageError": titleMessageError,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
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

	template := "templates/postadd.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{
		"title":             "Add Post",
		"post":              pm,
		"user":              sess.User.Username,
		"postMessage":       postMessage,
		"postMessageError":  postMessageError,
		"titleMessage":      titleMessage,
		"titleMessageError": titleMessageError,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
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
