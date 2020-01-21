package session

import (
	"encoding/json"
	"fmt"
	"net/http"

	"blog/blog/cache"

	"github.com/gomodule/redigo/redis"
	uuid "github.com/satori/go.uuid"
)

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

// Create a session object
func (s *Session) Create(creds Credentials) error {
	cache, err := cache.GetRedisConn()

	defer cache.Close()

	// Create a new random session token
	s.SessionToken = uuid.NewV4().String()

	var user User

	user.Username = creds.Username

	byteResult, err := json.Marshal(user)
	if err != nil {
		fmt.Printf("Error marshaling json object %s", err)
	}

	// Set the token in the cache, along with the user whom it represents
	// The token has an expiry time of 120 seconds
	_, err = cache.Do("SETEX", s.SessionToken, "1800", string(byteResult))
	if err != nil {
		return fmt.Errorf("Error saving session to redis: %s", err)
	}

	return nil
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

// Session Session Test if user is authenticated
func (s *Session) Session(r *http.Request) error {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return err
		}
		return nil
	}

	s.SessionToken = c.Value

	cache, err := cache.GetRedisConn()

	defer cache.Close()

	if err != nil {
		return fmt.Errorf("Error getting cache object, empty object returned: %s", err)
	}

	if cache == nil {
		return fmt.Errorf("Error getting cache object, empty object returned")
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := redis.String(cache.Do("GET", s.SessionToken))
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		return fmt.Errorf("Error? No Response from Cache: %s", err)
	}

	user := &User{}

	err = json.Unmarshal([]byte(response), user)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		return fmt.Errorf("Error decoding json object: %s", err)
	}
	s.User = *user
	s.IsAuth = true

	return nil
}
