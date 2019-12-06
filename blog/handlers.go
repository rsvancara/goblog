package blog

import (
	"fmt"
	"net/http"
	"time"

	"github.com/flosch/pongo2"
	"github.com/gomodule/redigo/redis"
	uuid "github.com/satori/go.uuid"
)

// ContextKey key used by contexts to uniquely identify attributes
type ContextKey string

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

// Credentials Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

func getRedisConn() redis.Conn {
	// Establish a connection to the Redis server listening on port
	// 6379 of the local machine. 6379 is the default port, so unless
	// you've already changed the Redis configuration file this should
	// work.
	conn, err := redis.Dial("tcp", "10.152.64.116:32777")
	if err != nil {
		fmt.Println(err)
	}
	// Importantly, use defer to ensure the connection is always
	// properly closed before exiting the main() function.

	return conn
}

// AuthHandler authorize user
func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// We can obtain the session token from the requests cookies, which come with every request
		c, err := r.Cookie("session_token")

		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				//w.WriteHeader(http.StatusUnauthorized)
				//fmt.Fprintf(w, "Error %s", err)
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error  %s", err)
			return
		}
		sessionToken := c.Value

		cache := GetRedisConn()

		defer cache.Close()

		// We then get the name of the user from our cache, where we set the session token
		response, err := cache.Do("GET", sessionToken)
		if err != nil {
			// If there is an error fetching from cache, return an internal server error status
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if response == nil {
			// If the session token is not present in cache, return an unauthorized error
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)

	})
}

// HomeHandler Home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)

	template := "templates/index.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}

// Signin Sign into the application
func Signin(w http.ResponseWriter, r *http.Request) {

	fmt.Println(r.Method)
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		var creds Credentials
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

		cache := GetRedisConn()
		defer cache.Close()

		// Create a new random session token
		sessionToken := uuid.NewV4().String()
		// Set the token in the cache, along with the user whom it represents
		// The token has an expiry time of 120 seconds
		_, err := cache.Do("SETEX", sessionToken, "120", creds.Username)
		if err != nil {
			// If there is an error in setting the cache, return an internal server error
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Finally, we set the client cookie for "session_token" as the session token we just generated
		// we also set an expiry time of 120 seconds, the same as the cache
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(120 * time.Second),
		})

		template := "templates/index.html"
		tmpl := pongo2.Must(pongo2.FromFile(template))

		out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println(err)
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, out)

	} else {
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

}

func Welcome(w http.ResponseWriter, r *http.Request) {

	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(fmt.Sprintf("There was a problem: %s!", err)))
			fmt.Println("Bad request no cookie for you!")

			return
		}
		fmt.Println("Bad request no cookie for you!")
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	//var redisCtxID ContextKey = "redis"
	//ctx := r.Context()

	cache := GetRedisConn()

	//var cache = (redis.Conn)(nil)

	//cache = ctx.Value(redisCtxID).(redis.Conn)

	if cache == nil {
		// If there is an error fetching from cache, return an internal server error status
		fmt.Println("Need a cache")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := cache.Do("GET", sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		fmt.Println("Error? No Response from Cache")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if response == nil {
		// If the session token is not present in cache, return an unauthorized error
		fmt.Println("Error? No Response")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	fmt.Println("Write Someting?")
	// Finally, return the welcome message to the user
	w.Write([]byte(fmt.Sprintf("Welcome %s!", response)))
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value

	var redisCtxID ContextKey = "redis"
	ctx := r.Context()
	cache := ctx.Value(redisCtxID).(redis.Conn)

	response, err := cache.Do("GET", sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if response == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// The code uptil this point is the same as the first part of the `Welcome` route

	// Now, create a new session token for the current user
	newSessionToken := uuid.NewV4().String()
	_, err = cache.Do("SETEX", newSessionToken, "120", fmt.Sprintf("%s", response))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Delete the older session token
	_, err = cache.Do("DEL", sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users `session_token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   newSessionToken,
		Expires: time.Now().Add(120 * time.Second),
	})
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	template := "templates/about.html"
	tmpl := pongo2.Must(pongo2.FromFile(template))

	out, err := tmpl.Execute(pongo2.Context{"title": "Index", "greating": "Hello"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, out)
}
