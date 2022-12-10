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
type Session struct {
	SessionToken string `json:"sessiontoken"` // Session Token
	//IsAuth       bool   `json:"isauth"`       // Is Authenticated, true/false
	User User `json:"user"` // User Object
	TTL  int  `json:"ttl"`  // Session Time To Live
}

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

	// Authentication peformed here!!
	if creds.Username != cfg.GetAdminUser() || creds.Password != cfg.GetAdminPassword() {
		log.Info().Msgf("error authenticating user %s", creds.Username)
		return false, nil
	}

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

	s.User.Username = creds.Username
	s.User.IsAuth = true
	s.User.CreatedAt = time.Now()

	// Attempt to extract additional information from a context
	//var geoIP requestfilter.GeoIP
	// CtxKey Context Key
	type contextKey string
	var ctxKey contextKey = "geoip"
	//ctxKey = "geoip"

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
		log.Info().Msg("Could not find ctxkey: geoip")
	}

	err = s.Save(cache, cfg.RedisDB)
	if err != nil {
		log.Error().Err(err).Msgf("error saving saving session in authentication")
		return false, fmt.Errorf("error saving sessiont: %s", err)
	}

	return true, nil
}

// Set the new session
func (s *Session) Session(cache cache.Cache, redisdb string, r *http.Request, w http.ResponseWriter) error {

	cfg, err := config.GetConfig()
	if err != nil {
		log.Error().Err(err).Msg("error getting configuration")
		return err
	}

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
		user.CreatedAt = time.Now()

		var geoIP requestfilter.GeoIP

		// The context may exist, but maybe not....Try to get information from the context, if
		// not found then extract it manually by calling the API.  There are conditions
		// where this can be pulled from the context, but most likely you can not.  Maybe
		// we do not use the context route, since sessions come first...
		type contextKey string
		var ctxKey contextKey = "geoip"
		//ctxKey = "geoip"

		if result := r.Context().Value(ctxKey); result != nil {
			geoIP, ok := result.(requestfilter.GeoIP)
			if !ok {
				log.Info().Msg("could not perform type assertion on result to GeoIP type")
			}

			user.City = geoIP.City
			user.TimeZone = geoIP.TimeZone
			user.Country = geoIP.CountryName
			user.IPAddress = geoIP.IPAddress.String()
			user.Organization = geoIP.Organization
			user.ASN = geoIP.ASN
		} else {
			log.Info().Msgf("could not find ctxkey geoip during session %s, performing manual lookup", s.SessionToken)
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

		// Set the user
		s.User = user
		//s.IsAuth = false
		s.TTL = int(cfg.GetIntegerSessionTimeout())

		// Place the cookie into the response header
		log.Info().Msgf("Setting Cookie with id: %s\n", newCookie.Value)
		http.SetCookie(w, &newCookie)

		err := s.Save(cache, cfg.RedisDB)
		if err != nil {
			log.Error().Err(err).Msg("error saving new session to the redis cache")
		}

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

	return nil
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

		sess.User = *user
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
