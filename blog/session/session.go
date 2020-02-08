package session

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"blog/blog/cache"
	"blog/blog/config"
	"blog/blog/requestfilter"
	"blog/blog/util"

	"github.com/gomodule/redigo/redis"
	uuid "github.com/satori/go.uuid"
)

const sessionTimeout = 84600

var users = map[string]string{
	"user1": "password1",
	"user2": "password2",
}

// Credentials Create a struct that models the structure of a user, both in the request body, and in the DB
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

// Item User items
type Item struct {
	Key string `json:"key"`
	Val string `json:"val"`
}

// User stores user information for a session
type User struct {
	Username  string `json:"username"`
	Items     []Item `json:"items"`
	IsAuth    bool   `json:"isauth"`
	IPAddress string `json:"ipaddress"`
	City      string `json:"city"`
	TimeZone  string `json:"timezone"`
	Country   string `json:"country"`
}

func (u *User) setItem(key string, value string) {
	var found bool

	for k := range u.Items {
		if u.Items[k].Key == key {
			u.Items[k].Val = value
			found = true
			break
		}
	}
	if found == false {
		u.Items = append(u.Items, Item{key, value})
	}
}

func (u *User) getItem(key string) string {
	for k := range u.Items {
		if u.Items[k].Key == key {
			return u.Items[k].Val
		}
	}
	return ""
}

// Session a user session object
type Session struct {
	SessionToken string
	IsAuth       bool
	User         User
}

// Get Get the session key value for provided key
func (s *Session) Get(key string) error {
	cache, err := cache.GetRedisConn()
	if err != nil {
		return fmt.Errorf("error connecting to redis during session creation: %s", err)
	}
	defer cache.Close()

	// We then get the name of the user from our cache, where we set the session token
	response, err := redis.String(cache.Do("GET", key))

	if err != nil {
		return fmt.Errorf("Error retrieving user object from redis: %s", err)
	}

	user := &User{}

	err = json.Unmarshal([]byte(response), user)

	s.User = *user
	s.IsAuth = user.IsAuth
	s.SessionToken = key

	return nil
}

// Authenticate Authenticate a user
func (s *Session) Authenticate(creds Credentials, r *http.Request, w http.ResponseWriter) (bool, error) {

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("error getting configuration: %s", err)
	}

	// Get the existing cookie
	c, err := r.Cookie("session_token")
	if err != nil {
		return false, fmt.Errorf("error getting cookie during authentication: %s", err)
	}

	s.SessionToken = c.Value

	// Get the expected password from our in memory map
	//expectedPassword, ok := users[creds.Username]

	// Authenticate the user here!!
	//if !ok || expectedPassword != creds.Password {
	//	return false, nil
	//}

	if creds.Username != cfg.GetAdminUser() || creds.Password != cfg.GetAdminPassword() {
		return false, nil
	}

	// Connect to Redis and get our user object
	cache, err := cache.GetRedisConn()
	defer cache.Close()
	if err != nil {
		// Expire the cookie
		return false, fmt.Errorf("error getting cache object for session %s with error  %s", c.Value, err)
	}

	// We should get a cache object, the above code should ensure this
	if cache == nil {
		return false, fmt.Errorf("Error getting cache object, empty object returned")
	}

	// get the name of the user from cache, where the session token exists
	response, err := redis.String(cache.Do("GET", s.SessionToken))
	if err != nil {
		c.Expires = time.Now().Add(-1)
		http.SetCookie(w, c)
		// If there is an error fetching from cache, return an internal server error status
		return false, fmt.Errorf("Error? No Response from Cache: %s", err)
	}

	// Unmarshal the user object from Redis
	user := &User{}
	err = json.Unmarshal([]byte(response), user)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		return false, fmt.Errorf("Error decoding json object: %s", err)
	}

	// Update the user object
	//var user User
	user.Username = creds.Username
	user.IsAuth = true
	s.IsAuth = true

	// Attempt to extract additional information from a context
	//var geoIP requestfilter.GeoIP
	var ctxKey util.CtxKey
	ctxKey = "geoip"

	if result := r.Context().Value(ctxKey); result != nil {

		//fmt.Println("Found context")
		//fmt.Println(result)
		// Type Assertion....
		geoIP, ok := result.(requestfilter.GeoIP)
		if !ok {
			fmt.Println("Could not perform type assertion on result to GeoIP type")
		}

		user.City = geoIP.City
		user.TimeZone = geoIP.TimeZone
		user.Country = geoIP.CountryName
		user.IPAddress = geoIP.IPAddress.String()
	} else {
		fmt.Println("Could not find ctxkey: geoip")
	}

	// Dont forget to update the session object with the new user information
	s.User = *user

	// Convert to JSON
	byteResult, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Error marshaling json object %s", err)
	}

	// Remove existing session
	_, err = cache.Do("DEL", s.SessionToken)
	if err != nil {
		return false, fmt.Errorf("Error removing session in redis: %s", err)
	}

	//fmt.Printf("Session Timeout in authentication %s", cfg.GetSessionTimeout())

	// Set/Replace the token in the cache, along with t he user whom it represents
	// The token has an expiry time of 120 seconds
	_, err = cache.Do("SETEX", s.SessionToken, cfg.GetIntegerSessionTimeout(), string(byteResult))
	if err != nil {
		return false, fmt.Errorf("Error saving session to redis: %s", err)
	}

	// Update the cookie with the same exipiration so they are synchronized
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   c.Value,
		Path:    "/",
		Expires: time.Now().Add(cfg.GetDurationTimeout()),
	})

	return true, nil
}

// SetValue get the session value for provided key
func (s *Session) SetValue(key string, value string) error {
	if s.SessionToken == "" {
		return fmt.Errorf("Please use Session.Session() method to instantiate the session object")
	}

	cache, err := cache.GetRedisConn()
	if err != nil {
		return fmt.Errorf("Error connecting to redis while trying to set value for key %s with error %s", key, err)
	}

	defer cache.Close()

	if err != nil {
		return fmt.Errorf("Error connecting to redis: %s", err)
	}

	if cache == nil {

		return fmt.Errorf("Error connecting to redis, empty connection object returned")
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := redis.String(cache.Do("GET", s.SessionToken))

	if err != nil {
		return fmt.Errorf("Error retrieving user object from redis: %s", err)
	}

	user := &User{}

	err = json.Unmarshal([]byte(response), user)

	user.setItem(key, value)

	byteUser, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("Error marshing user object to byte array json: %s", err)
	}

	response, err = redis.String(cache.Do("SET", s.SessionToken, string(byteUser)))
	if err != nil {
		return fmt.Errorf("Error storing object in redis: %s", err)
	}

	return nil
}

// GetValue get the session value for provided key
func (s *Session) GetValue(key string) (string, error) {
	if s.SessionToken == "" {
		return "", fmt.Errorf("Please use Session.Session() method to instantiate the session object")
	}

	cache, err := cache.GetRedisConn()
	defer cache.Close()
	if err != nil {
		return "", fmt.Errorf("Error connecting to redis while trying to get value for key %s with error %s", key, err)
	}

	if cache == nil {
		return "", fmt.Errorf("Error connecting to redis, empty connection object returned")
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := redis.String(cache.Do("GET", s.SessionToken))
	if err != nil {
		return "", fmt.Errorf("Error retrieving user object from redis: %s", err)
	}

	user := &User{}

	err = json.Unmarshal([]byte(response), user)
	if err != nil {
		return "", fmt.Errorf("Error reading json object from redis: %s", err)
	}

	val := user.getItem(key)

	return val, nil
}

// Session Session Test if user isauthenticated
func (s *Session) Session(r *http.Request, w http.ResponseWriter) error {

	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("error getting configuration: %s", err)
		return err
	}

	// We can obtain the session token from the requests cookies, which come with every request
	// This code ensures a session token is always created

	// Generate session token
	s.SessionToken = uuid.NewV4().String()

	//fmt.Printf("%s\n", cfg.GetDurationTimeout())

	// Create a new cookie
	var newCookie http.Cookie
	newCookie.Name = "session_token"
	newCookie.Value = s.SessionToken
	newCookie.Path = "/"
	newCookie.Expires = time.Now().Add(cfg.GetDurationTimeout())

	cache, err := cache.GetRedisConn()
	if err != nil {
		return fmt.Errorf("error creating a redis session object for %s with error: %s", s.SessionToken, err)
	}
	defer cache.Close()

	if cache == nil {
		return fmt.Errorf("redis connection is nil for session token %s", s.SessionToken)
	}

	// Default there is no cookie error

	isCookieError := false

	// Attempt to get the current cookie and if it does not exist, build a new cookie
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// No cookie exists, create the new cookie.
			fmt.Printf("no cookie exists, need to create a new cookie: %s\n", err)

			// Populate an anonymous user object

			var user User
			user.Username = "anonymous"
			user.IsAuth = false

			// Attempt to extract additional information from a context
			var ctxKey util.CtxKey
			ctxKey = "geoip"

			if result := r.Context().Value(ctxKey); result != nil {

				//fmt.Println("Found context")
				//fmt.Println(result)
				// Type Assertion....
				geoIP, ok := result.(requestfilter.GeoIP)
				if !ok {
					fmt.Println("Could not perform type assertion on result to GeoIP type")
				}

				user.City = geoIP.City
				user.TimeZone = geoIP.TimeZone
				user.Country = geoIP.CountryName
				user.IPAddress = geoIP.IPAddress.String()
			} else {
				fmt.Printf("Could not find ctxkey: geoip during session %s\n", s.SessionToken)
			}

			//TODO: Set the user
			s.User = user

			//fmt.Println(s)

			// Set the new cookie
			fmt.Printf("Setting Cookie with id: %s\n", newCookie.Value)
			http.SetCookie(w, &newCookie)

			// Create the Redis Object which holds the session variables
			// in a backend Redis Cache
			//fmt.Printf("Creating redis session for cookie %s\n", newCookie.Value)

			// Convert object to JSON
			byteResult, err := json.Marshal(s.User)
			if err != nil {
				fmt.Printf("Error marshaling json object %s\n", err)
				return err
			}

			//fmt.Printf("creating redis connection for session token %s\n", s.SessionToken)
			//fmt.Printf("Session Timeout %s is now %d\n", cfg.SessionTimeout, cfg.GetIntegerSessionTimeout())
			// Set the token in the cache, along with the user whom it represents
			// The token has an expiry time of 120 seconds
			_, err = cache.Do("SETEX", s.SessionToken, cfg.GetIntegerSessionTimeout(), string(byteResult))
			if err != nil {
				return fmt.Errorf("Error saving session to redis: %s", err)
			}

			// Since we have an error We need to flag the creation of a new cookie
			isCookieError = true
			//fmt.Printf("Creating new session %s\n", s.SessionToken)
		}
	}

	// Create the new cookie because of the error above
	if isCookieError == true {
		c = &newCookie
	}

	// Set the session token value to the cookie value so they match
	s.SessionToken = c.Value

	// get the name of the user from cache, where the session token exists
	response, err := redis.String(cache.Do("GET", s.SessionToken))
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		return fmt.Errorf("Error? No Response from Cache: %s", err)
	}

	// Unmarshal the user object from Redis
	user := &User{}
	err = json.Unmarshal([]byte(response), user)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		return fmt.Errorf("Error decoding json object: %s", err)
	}

	//fmt.Printf("got user object from session %s\n", user.Username)

	// Set the session object
	s.User = *user
	s.IsAuth = s.User.IsAuth

	return nil
}

// GetAllSessions get all the sessions stored in Redis
func GetAllSessions() ([]Session, error) {

	var sessions []Session

	keys, err := getKeys("*")
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range keys {
		var sess Session
		sess.Get(v)

		sessions = append(sessions, sess)
	}

	return sessions, nil
}

func getKeys(pattern string) ([]string, error) {

	// Connect to Redis and get our user object
	cache, err := cache.GetRedisConn()
	defer cache.Close()
	if err != nil {
		return nil, fmt.Errorf("Error getting cache object, empty object returned: %s", err)
	}

	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(cache.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil

}

// DeleteSession deletes a redis session
func DeleteSession(id string) error {

	cache, err := cache.GetRedisConn()

	_, err = cache.Do("DELETE", id)
	if err != nil {
		return fmt.Errorf("Error deleting session key %s in redis: %s", id, err)
	}

	return nil
}
