package session

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"blog/blog/cache"

	"github.com/gomodule/redigo/redis"
	uuid "github.com/satori/go.uuid"
)

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
	Username string `json:"username"`
	Items    []Item `json:"items"`
	IsAuth   bool   `json:"isauth"`
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

// Create a session object in Redis
func (s *Session) Create() error {

	// Create the user object
	var user User
	user.Username = "anonymous"
	user.IsAuth = false

	// Convert object to JSON
	byteResult, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Error marshaling json object %s", err)
	}

	cache, err := cache.GetRedisConn()
	if err != nil {
		return fmt.Errorf("error connecting to redis during session creation: %s", err)
	}
	defer cache.Close()

	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds
	_, err = cache.Do("SETEX", s.SessionToken, "86400", string(byteResult))
	if err != nil {
		return fmt.Errorf("Error saving session to redis: %s", err)
	}

	return nil
}

// Authenticate Authenticate a user
func (s *Session) Authenticate(creds Credentials, r *http.Request, w http.ResponseWriter) (bool, error) {

	// Get the existing cookie
	c, err := r.Cookie("session_token")

	s.SessionToken = c.Value

	// Get the expected password from our in memory map
	expectedPassword, ok := users[creds.Username]

	// Authenticate the user here!!
	if !ok || expectedPassword != creds.Password {
		return false, nil
	}

	// Update the user object
	var user User
	user.Username = creds.Username
	user.IsAuth = true
	s.IsAuth = true
	// Dont forget to update the session object with the new user information
	s.User = user

	// Convert to JSON
	byteResult, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Error marshaling json object %s", err)
	}

	// Connect to Redis
	cache, err := cache.GetRedisConn()
	if err != nil {
		return false, fmt.Errorf("error connecting to redis during authentication: %s", err)
	}
	defer cache.Close()

	// Remove existing session
	_, err = cache.Do("DEL", s.SessionToken)
	if err != nil {
		return false, fmt.Errorf("Error removing session in redis: %s", err)
	}

	// Set/Replace the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds
	_, err = cache.Do("SETEX", s.SessionToken, "86400", string(byteResult))
	if err != nil {
		return false, fmt.Errorf("Error saving session to redis: %s", err)
	}

	// Update the cookie with the same exipiration so they are synchronized
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   c.Value,
		Path:    "/",
		Expires: time.Now().Add(86400 * time.Second),
	})

	return true, nil
}

// SetValue get the session value for provided key
func (s *Session) SetValue(key string, value string) error {
	if s.SessionToken == "" {
		return fmt.Errorf("Please use Session.Session() method to instantiate the session object")
	}

	cache, err := cache.GetRedisConn()

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
		return "", fmt.Errorf("Error connecting to redis: %s", err)
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

	// We can obtain the session token from the requests cookies, which come with every request
	// This code ensures a session token is always created
	s.SessionToken = uuid.NewV4().String()
	var newCookie http.Cookie
	newCookie.Name = "session_token"
	newCookie.Value = s.SessionToken
	newCookie.Path = "/"
	newCookie.Expires = time.Now().Add(86400 * time.Second)

	isCookieError := false

	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			//Create new cookie
			http.SetCookie(w, &newCookie)
			// Create redis object
			s.Create()
			isCookieError = true
			fmt.Printf("Creating new session %s\n", s.SessionToken)
		}
	}

	if isCookieError == true {
		c = &newCookie
	}

	// Set the session token value to the cookie value so they match
	s.SessionToken = c.Value

	// Connect to Redis and get our user object
	cache, err := cache.GetRedisConn()
	defer cache.Close()
	if err != nil {
		return fmt.Errorf("Error getting cache object, empty object returned: %s", err)
	}

	// We should get a cache object, the above code should ensure this
	if cache == nil {
		return fmt.Errorf("Error getting cache object, empty object returned")
	}

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

	// Set the session object
	s.User = *user
	s.IsAuth = s.User.IsAuth

	return nil
}

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

	fmt.Println(sessions)

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
