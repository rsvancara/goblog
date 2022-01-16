package sessionmanager

import (
	"encoding/json"
	"fmt"
	"goblog/internal/cache"
	"goblog/internal/config"
	"goblog/internal/requestfilter"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

// Session a user session object
//type Session struct {
//	SessionToken string
//	IsAuth       bool
//	User         User
//	TTL          int
//}

// Get Get the session key value for provided key
// func (s *Session) Get(key string) error {
// 	cache, err := cache.InitPool()
// 	if err != nil {
// 		return fmt.Errorf("error connecting to redis during session creation: %s", err)
// 	}
// 	defer cache.Close()

// 	// We then get the name of the user from our cache, where we set the session token
// 	response, err := redis.String(cache.Do("GET", key))
// 	if err != nil {
// 		return fmt.Errorf("error retrieving user object from redis: %s", err)
// 	}

// 	user := &User{}

// 	err = json.Unmarshal([]byte(response), user)
// 	if err != nil {
// 		return fmt.Errorf("error unmarshaling session %s with error %s", key, err)
// 	}

// 	ttl, err := redis.Int(cache.Do("TTL", key))
// 	if err != nil {
// 		return fmt.Errorf("error unmarshaling session %s with error %s", key, err)
// 	}

// 	s.User = *user
// 	s.IsAuth = user.IsAuth
// 	s.SessionToken = key
// 	s.TTL = ttl

// 	return nil
// }

// Authenticate Authenticate a user
func (s *Session) Authenticate(cache cache.Cache, redisdb string, creds Credentials, r *http.Request, w http.ResponseWriter) (bool, error) {

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
		log.Info().Msgf("error authenticating user %s", creds.Username)
		return false, nil
	}

	// Connect to Redis and get our user object
	// cache, err := cache.GetRedisConn()
	// defer cache.Close()
	// if err != nil {
	// 	// Expire the cookie
	// 	return false, fmt.Errorf("error getting cache object for session %s with error  %s", c.Value, err)
	// }

	// // We should get a cache object, the above code should ensure this
	// if cache == nil {
	// 	return false, fmt.Errorf("error getting cache object, empty object returned")
	// }

	// get the name of the user from cache, where the session token exists
	//response, err := redis.String(cache.Do("GET", s.SessionToken))
	err = s.Get(cache, cfg.RedisDB, s.SessionToken)
	if err != nil {
		c.Expires = time.Now().Add(-1)
		http.SetCookie(w, c)
		log.Error().Err(err).Msgf("error getting session from cache")
		// If there is an error fetching from cache, return an internal server error status
		return false, fmt.Errorf("error? No Response from Cache: %s", err)
	}

	// Unmarshal the user object from Redis
	//user := &User{}
	//err = json.Unmarshal([]byte(response), user)
	//if err != nil {
	//	// If there is an error fetching from cache, return an internal server error status
	//	return false, fmt.Errorf("error decoding json object: %s", err)
	//}

	// Update the user object
	//var user User
	s.User.Username = creds.Username
	s.User.IsAuth = true
	s.IsAuth = true

	// Attempt to extract additional information from a context
	//var geoIP requestfilter.GeoIP
	// CtxKey Context Key
	type contextKey string
	var ctxKey contextKey
	ctxKey = "geoip"

	if result := r.Context().Value(ctxKey); result != nil {

		// Type Assertion....
		geoIP, ok := result.(requestfilter.GeoIP)
		if !ok {
			fmt.Println("Could not perform type assertion on result to GeoIP type")
		}

		s.User.City = geoIP.City
		s.User.TimeZone = geoIP.TimeZone
		s.User.Country = geoIP.CountryName
		s.User.IPAddress = geoIP.IPAddress.String()
	} else {
		fmt.Println("Could not find ctxkey: geoip")
	}

	// Dont forget to update the session object with the new user information
	//s.User = *user

	// Convert to JSON
	//byteResult, err := json.Marshal(user)
	//if err != nil {
	//	fmt.Printf("Error marshaling json object %s", err)
	//}

	// Remove existing session
	//_, err = cache.Do("DEL", s.SessionToken)
	//if err != nil {
	//	return false, fmt.Errorf("error removing session in redis: %s", err)
	//}

	// Set/Replace the token in the cache, along with t he user whom it represents
	// The token has an expiry time of 120 seconds
	//_, err = cache.Do("SETEX", s.SessionToken, cfg.GetIntegerSessionTimeout(), string(byteResult))
	//if err != nil {
	//	return false, fmt.Errorf("error saving session to redis: %s", err)
	//}

	// Update the cookie with the same exipiration so they are synchronized
	// http.SetCookie(w, &http.Cookie{
	// 	Name:    "session_token",
	// 	Value:   c.Value,
	// 	Path:    "/",
	// 	Expires: time.Now().Add(cfg.GetDurationTimeout()),
	// })
	err = s.Save(cache, cfg.RedisDB)
	if err != nil {
		log.Error().Err(err).Msgf("error saving saving session in authentication")
		return false, fmt.Errorf("error saving sessiont: %s", err)
	}

	return true, nil
}

// // SetValue get the session value for provided key
// func (s *Session) SetValue(key string, value string) error {
// 	if s.SessionToken == "" {
// 		return fmt.Errorf("please use Session.Session() method to instantiate the session object")
// 	}

// 	cache, err := cache.GetRedisConn()
// 	if err != nil {
// 		return fmt.Errorf("error connecting to redis while trying to set value for key %s with error %s", key, err)
// 	}

// 	defer cache.Close()

// 	if err != nil {
// 		return fmt.Errorf("error connecting to redis: %s", err)
// 	}

// 	if cache == nil {

// 		return fmt.Errorf("error connecting to redis, empty connection object returned")
// 	}

// 	// We then get the name of the user from our cache, where we set the session token
// 	response, err := redis.String(cache.Do("GET", s.SessionToken))

// 	if err != nil {
// 		return fmt.Errorf("error retrieving user object from redis: %s", err)
// 	}

// 	user := &User{}

// 	err = json.Unmarshal([]byte(response), user)

// 	user.SetItem(key, value)

// 	byteUser, err := json.Marshal(user)
// 	if err != nil {
// 		return fmt.Errorf("error marshing user object to byte array json: %s", err)
// 	}

// 	_, err = redis.String(cache.Do("SET", s.SessionToken, string(byteUser)))
// 	if err != nil {
// 		return fmt.Errorf("error storing object in redis: %s", err)
// 	}

// 	return nil
// }

// GetValue get the session value for provided key
// func (s *Session) GetValue(key string) (string, error) {
// 	if s.SessionToken == "" {
// 		return "", fmt.Errorf("please use Session.Session() method to instantiate the session object")
// 	}

// 	cache, err := cache.GetRedisConn()
// 	defer cache.Close()
// 	if err != nil {
// 		return "", fmt.Errorf("error connecting to redis while trying to get value for key %s with error %s", key, err)
// 	}

// 	if cache == nil {
// 		return "", fmt.Errorf("error connecting to redis, empty connection object returned")
// 	}

// 	// We then get the name of the user from our cache, where we set the session token
// 	response, err := redis.String(cache.Do("GET", s.SessionToken))
// 	if err != nil {
// 		return "", fmt.Errorf("error retrieving user object from redis: %s", err)
// 	}

// 	user := &User{}

// 	err = json.Unmarshal([]byte(response), user)
// 	if err != nil {
// 		return "", fmt.Errorf("error reading json object from redis: %s", err)
// 	}

// 	val := user.GetItem(key)

// 	return val, nil
// }

// Set the new session
func (s *Session) Session(cache cache.Cache, redisdb string, r *http.Request, w http.ResponseWriter) error {

	cfg, err := config.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("error getting configuration")
		return err
	}

	// We can obtain the session token from the requests cookies, which come with every request
	// This code ensures a session token is always created

	//cache, err := cache.GetRedisConn()
	//if err != nil {
	//	return fmt.Errorf("error creating a redis session object for %s with error: %s", s.SessionToken, err)
	//}
	//defer cache.Close()

	//if cache == nil {
	//	return fmt.Errorf("redis connection is nil for session token %s", s.SessionToken)
	//}

	// Default there is no cookie error

	// Attempt to get the current cookie and if it does not exist, build a new cookie
	c, err := r.Cookie("session_token")
	if err != nil {
		log.Info().Err(err).Msg("error finding coodie or cookie does not exist")
	}

	// No cookie exists so we can create a new cookie
	if err == http.ErrNoCookie {
		// No cookie exists, create the new cookie.
		fmt.Printf("no cookie exists, need to create a new cookie: %s\n", err)

		// Generate session token
		s.SessionToken = uuid.NewV4().String()

		// Create a new cookie
		var newCookie http.Cookie
		newCookie.Name = "session_token"
		newCookie.Value = s.SessionToken
		newCookie.Path = "/"
		newCookie.Expires = time.Now().Add(cfg.GetDurationTimeout())

		// Populate an anonymous user object

		var user User
		user.Username = "anonymous"
		user.IsAuth = false

		var geoIP requestfilter.GeoIP

		// The context may exist, but maybe not....Try to get information from the context, if
		// not found then extract it manually by calling the API.  There are conditions
		// where this can be pulled from the context, but most likely you can not.  Maybe
		// we do not use the context route, since sessions come first...
		type contextKey string
		var ctxKey contextKey
		ctxKey = "geoip"

		if result := r.Context().Value(ctxKey); result != nil {
			geoIP, ok := result.(requestfilter.GeoIP)
			if !ok {
				fmt.Println("Could not perform type assertion on result to GeoIP type")
			}

			user.City = geoIP.City
			user.TimeZone = geoIP.TimeZone
			user.Country = geoIP.CountryName
			user.IPAddress = geoIP.IPAddress.String()
			user.Organization = geoIP.Organization
			user.ASN = geoIP.ASN
		} else {
			fmt.Printf("Could not find ctxkey geoip during session %s, performing manual lookup\n", s.SessionToken)
			ipaddress, _ := requestfilter.GetIPAddress(r)
			err := geoIP.GeoIPSearch(ipaddress, cfg)
			if err != nil {
				log.Error().Err(err).Msgf("GeoIP Address search error: %s", ipaddress)

			}
			user.City = geoIP.City
			user.TimeZone = geoIP.TimeZone
			user.Country = geoIP.CountryName
			user.IPAddress = geoIP.IPAddress.String()
			user.Organization = geoIP.Organization
			user.ASN = geoIP.ASN

		}

		//TODO: Set the user
		s.User = user
		s.IsAuth = false
		s.TTL = int(cfg.GetIntegerSessionTimeout())

		//fmt.Println(s)

		// Place the cookie into the response header
		log.Info().Msgf("Setting Cookie with id: %s\n", newCookie.Value)
		http.SetCookie(w, &newCookie)

		// Create the Redis Object which holds the session variables
		// in a backend Redis Cache
		//fmt.Printf("Creating redis session for cookie %s\n", newCookie.Value)

		// Convert object to JSON
		//byteResult, err := json.Marshal(s.User)
		//if err != nil {
		//	fmt.Printf("Error marshaling json object %s\n", err)
		//	return err
		//}

		//fmt.Printf("creating redis connection for session token %s\n", s.SessionToken)
		//fmt.Printf("Session Timeout %s is now %d\n", cfg.SessionTimeout, cfg.GetIntegerSessionTimeout())
		// Set the token in the cache, along with the user whom it represents
		// The token has an expiry time of 120 seconds
		//_, err = cache.Do("SETEX", s.SessionToken, cfg.GetIntegerSessionTimeout(), string(byteResult))
		//if err != nil {
		//	return fmt.Errorf("Error saving session to redis: %s", err)
		//}

		s.Save(cache, cfg.RedisDB)

		// Since we have an error We need to flag the creation of a new cookie
		//isCookieError = true
		//fmt.Printf("Creating new session %s\n", s.SessionToken)

		return nil

	}

	//
	// Assume cookie exists
	//

	// Get the session token
	//s.SessionToken = c.Value
	log.Info().Msgf("looking for redis cache object for token %s", c.Value)
	err = s.Get(cache, cfg.RedisDB, c.Value)
	if err != nil {
		return fmt.Errorf("error getting cache object for token %s with error %s", c.Value, err)
	}

	// get the name of the user from cache, where the session token exists
	//response, err := redis.String(cache.Do("GET", s.SessionToken))
	//if err != nil {
	//	// If there is an error fetching from cache, return an internal server error status
	//	return fmt.Errorf("error? No Response from Cache: %s", err)
	//}

	// Unmarshal the user object from Redis
	//user := &User{}
	//err = json.Unmarshal([]byte(response), user)
	//if err != nil {
	// If there is an error fetching from cache, return an internal server error status
	//	return fmt.Errorf("error decoding json object: %s", err)
	//}

	//fmt.Printf("got user object from session %s\n", user.Username)

	// Set the session object
	//s.User = *user
	//s.IsAuth = s.User.IsAuth

	return nil
}

// GetAllSessions get all the sessions stored in Redis
// func GetAllSessions() ([]Session, error) {

// 	var sessions []Session

// 	keys, err := getKeys("*")
// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	for _, v := range keys {
// 		var sess Session
// 		sess.Get(v)

// 		sessions = append(sessions, sess)
// 	}

// 	// Sort by the TTL
// 	sort.Slice(sessions, func(i, j int) bool {
// 		return sessions[i].TTL < sessions[j].TTL
// 	})

// 	return sessions, nil
// }

// getKey internal method for getting keys for a supplied pattern, like "*"
// func getKeys(pattern string) ([]string, error) {

// 	// Connect to Redis and get our user object
// 	cache, err := cache.GetRedisConn()
// 	defer cache.Close()
// 	if err != nil {
// 		return nil, fmt.Errorf("error getting cache object, empty object returned: %s", err)
// 	}

// 	iter := 0
// 	keys := []string{}
// 	for {
// 		arr, err := redis.Values(cache.Do("SCAN", iter, "MATCH", pattern))
// 		if err != nil {
// 			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
// 		}

// 		iter, _ = redis.Int(arr[0], nil)
// 		k, _ := redis.Strings(arr[1], nil)
// 		keys = append(keys, k...)

// 		if iter == 0 {
// 			break
// 		}
// 	}

// 	return keys, nil

// }

// DeleteSession deletes a redis session
// func DeleteSession(id string) error {

// 	cache, err := cache.GetRedisConn()
// 	if err != nil {
// 		return fmt.Errorf("error deleting cache object %s, empty object returned: %s", id, err)
// 	}
// 	//fmt.Printf("deleting redis key %s\n", id)

// 	_, err = cache.Do("DEL", id)
// 	if err != nil {
// 		return fmt.Errorf("Error deleting session key %s in redis: %s", id, err)
// 	}

// 	return nil
// }

//
// NEW
//
//

// Session a user session object
type Session struct {
	SessionToken string `json:"sessiontoken"` // Session Token
	IsAuth       bool   `json:"isauth"`       // Is Authenticated, true/false
	User         User   `json:"user"`         // User Object
	TTL          int    `json:"ttl"`          // Session Time To Live
}

// Does the key exist
func (s *Session) Exists(cache cache.Cache, redisdb string, key string) (bool, error) {

	// Test if key exists
	exists, err := cache.Exists(key)
	if err != nil {
		return false, fmt.Errorf("error finding key %s in redis: %s", key, err)
	}

	return exists, nil

}

// Get the session key value for provided key
func (s *Session) Get(cache cache.Cache, redisdb string, key string) error {

	// Test if key exists
	exists, err := cache.Exists(key)
	if err != nil {
		return fmt.Errorf("error finding key %s in redis: %s", key, err)
	}

	// Return an empty session
	if !exists {
		var u User
		s.User = u
		return fmt.Errorf("error finding key %s in redis: %s", key, err)
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := cache.GetKey(key)
	if err != nil {
		return fmt.Errorf("error retrieving user object from redis: %s", err)
	}

	user := &User{}

	err = json.Unmarshal([]byte(response), user)
	if err != nil {
		return fmt.Errorf("error unmarshaling session %s with error %s", key, err)
	}

	ttl, err := cache.GetTTL(key)
	if err != nil {
		return fmt.Errorf("error unmarshaling session %s with error %s", key, err)
	}

	s.User = *user
	s.IsAuth = user.IsAuth
	s.SessionToken = key
	s.TTL = ttl

	return nil
}

// Save the session
func (s *Session) Save(cache cache.Cache, redisdb string) error {

	// Make sure the session has a token
	if s.SessionToken == "" {
		return fmt.Errorf("empty token provided [%s]", s.SessionToken)
	}

	if len(s.SessionToken) < 6 {
		return fmt.Errorf("please provide a token greater than six charagers in length to ensure unique values")
	}

	// Test if the token exists
	exists, err := cache.Exists(s.SessionToken)
	if err != nil {
		return fmt.Errorf("error checking is session exists ith error %s", err)
	}

	if exists {
		err := cache.Delete(s.SessionToken)
		if err != nil {
			return fmt.Errorf("error removing session %s", err.Error())
		}
	} else {
		log.Debug().Msg("Session does not exist and will be created")
	}

	// Convert to JSON
	byteResult, err := json.Marshal(s.User)
	if err != nil {
		return fmt.Errorf("error marshaling json object %s", err)
	}

	// Save or replace the cache item with a defined TTL
	cache.SetEx(s.SessionToken, string(byteResult), int64(s.TTL))

	return nil
}

// Set Session Item for session key.  Mostly a helper function so that items can be updated without
// having to deal with the entire user object and saving it back
func (s *Session) SetSessionItem(cache cache.Cache, redisdb string, key string, itemkey string, itemvalue string) error {
	// Make sure the session has a token
	if s.SessionToken == "" {
		return fmt.Errorf("empty token provided [%s]", s.SessionToken)
	}

	// Test if the token exists
	exists, err := cache.Exists(s.SessionToken)
	if err != nil {
		return fmt.Errorf("error finding session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	if !exists {
		return fmt.Errorf("could not find sessiont %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := cache.GetKey(key)
	if err != nil {
		return fmt.Errorf("could not find session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	user := &User{}

	err = json.Unmarshal([]byte(response), user)
	if err != nil {
		return fmt.Errorf("could extract JSON for session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	user.SetItem(itemkey, itemvalue)

	s.User = *user

	//fmt.Println(s)
	err = s.Save(cache, redisdb)
	if err != nil {
		return fmt.Errorf("could not save session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	return nil
}

// Set Session Item for session key.  Mostly a helper function so that items can be updated without
// having to deal with the entire user object and saving it back
func (s *Session) GetSessionItem(cache cache.Cache, redisdb string, key string, itemkey string) (string, error) {
	// Make sure the session has a token
	if s.SessionToken == "" {
		return "", fmt.Errorf("empty token provided [%s]", s.SessionToken)
	}

	// Test if the token exists
	exists, err := cache.Exists(s.SessionToken)
	if err != nil {
		return "", fmt.Errorf("error finding session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	if !exists {
		return "", fmt.Errorf("error finding session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := cache.GetKey(key)
	if err != nil {
		return "", fmt.Errorf("could not find session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	user := &User{}

	err = json.Unmarshal([]byte(response), user)
	if err != nil {
		return "", fmt.Errorf("could extract JSON for session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	val := user.GetItem(itemkey)

	return val, nil
}

// Delete Session Item for session key.  Mostly a helper function so that items can be updated without
// having to deal with the entire user object and saving it back
func (s *Session) DeleteSessionItem(cache cache.Cache, redisdb string, key string, itemkey string) error {
	// Make sure the session has a token
	if s.SessionToken == "" {
		return fmt.Errorf("empty token provided [%s]", s.SessionToken)
	}

	// Test if the token exists
	exists, err := cache.Exists(s.SessionToken)
	if err != nil {
		return fmt.Errorf("error finding session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	if !exists {
		return fmt.Errorf("could not find sessiont %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	// We then get the name of the user from our cache, where we set the session token
	response, err := cache.GetKey(key)
	if err != nil {
		return fmt.Errorf("could not find session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	user := &User{}

	err = json.Unmarshal([]byte(response), user)
	if err != nil {
		return fmt.Errorf("could extract JSON for session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	user.DeleteItem(itemkey)

	s.User = *user

	//fmt.Println(s)
	err = s.Save(cache, redisdb)
	if err != nil {
		return fmt.Errorf("could not save session %s while attempting to set item %s with error %s", key, itemkey, err)
	}

	return nil
}

// Delete the session key value for provided key
func (s *Session) Delete(cache cache.Cache, redisdb string, key string) error {

	err := cache.Delete(key)
	if err != nil {
		return fmt.Errorf("error removing session %s with error %s", key, err.Error())
	}

	return nil
}

// Get all the sessions in the cahce.  Does not extract the TTL unless it is stored in the user JSON
func GetAllSessions(cache cache.Cache, redisdb string, pattern string) ([]Session, error) {

	keylist, err := cache.GetAllVals(pattern)
	if err != nil {
		return nil, fmt.Errorf("error getting keys for filter %s with error %s", pattern, err.Error())
	}

	var sessions []Session

	for key, val := range keylist {

		var sess Session

		user := &User{}

		err = json.Unmarshal([]byte(val), user)
		if err != nil {
			return nil, fmt.Errorf("error unmarshaling session %s with error %s", key, err)
		}

		// This is too slow, need to find another way, maybe this all needs to be prepopulated

		//ttl, err := cache.GetTTL(key)
		//if err != nil {
		//	return nil, fmt.Errorf("error unmarshaling session %s with error %s", key, err)
		//}
		//user.TTL = ttl

		sess.User = *user
		sess.IsAuth = user.IsAuth
		sess.SessionToken = key
		sess.TTL = user.TTL

		sessions = append(sessions, sess)

	}

	return sessions, nil

}

func DeleteSession(cache cache.Cache, redisdb string, key string) error {
	// Make sure the session has a token
	if key == "" {
		return fmt.Errorf("empty token provided [%s]", key)
	}

	// Test if the token exists
	exists, err := cache.Exists(key)
	if err != nil {
		return fmt.Errorf("error finding session %s  with error %s", key, err)
	}

	if !exists {
		return fmt.Errorf("could not find sessiont %s  with error %s", key, err)
	}

	cache.Delete(key)
	if err != nil {
		return fmt.Errorf("error deleting session %s with error %s", key, err)
	}

	return nil

}
