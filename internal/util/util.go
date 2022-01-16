package util

import (
	"bytes"
	"fmt"
	"goblog/internal/config"
	"goblog/internal/requestfilter"
	"goblog/internal/sessionmanager"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

// CtxKey Context Key
type CtxKey string

// SiteTemplate loads the correct template directory for the site
func SiteTemplate(path string) (string, error) {

	cfg, err := config.GetConfig()
	if err != nil {
		return "", fmt.Errorf("error loading template directory %s", err)
	}
	return "sites/" + cfg.Site + path, nil

}

// Creates a new file upload http request with optional extra params
func ImageUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

//GeoIPContext get the geoip object
func GeoIPContext(r *http.Request) (requestfilter.GeoIP, error) {

	var geoIP requestfilter.GeoIP

	// Attempt to extract additional information from a context
	//var geoIP requestfilter.GeoIP
	var ctxKey CtxKey
	ctxKey = "geoip"

	if result := r.Context().Value(ctxKey); result != nil {

		//fmt.Println("Found context")
		//fmt.Println(result)
		// Type Assertion....
		geoIP, ok := result.(requestfilter.GeoIP)
		if !ok {
			return geoIP, fmt.Errorf("could not perform type assertion on result to GeoIP type for ctxKey %s on page %s", ctxKey, r.RequestURI)
		}

		return geoIP, nil

	}

	return geoIP, fmt.Errorf("unable to find context for geoip %s on page %s", ctxKey, r.RequestURI)
}

//SessionContext get the session object
func SessionContext(r *http.Request) (sessionmanager.Session, error) {

	var sess sessionmanager.Session

	// Attempt to extract additional information from a context
	var ctxKey CtxKey
	ctxKey = "session"

	if result := r.Context().Value(ctxKey); result != nil {

		//fmt.Println("Found context")
		//fmt.Println(result)
		// Type Assertion....
		sess, ok := result.(sessionmanager.Session)
		if !ok {
			return sess, fmt.Errorf("could not perform type assertion on result to session.Session type for ctxKey %s", ctxKey)
		}

		return sess, nil

	}

	return sess, fmt.Errorf("unable to find context for session %s on page %s", ctxKey, r.RequestURI)
}

// util.GetPageID get the page ID for a request
func GetPageID(r *http.Request) string {
	geoIP, err := GeoIPContext(r)
	if err != nil {
		log.Error().Err(err).Msgf("error obtaining geoip context")
		return ""
	}

	return geoIP.PageID
}

// GetSession get session object for a request
func GetSession(r *http.Request) sessionmanager.Session {
	sess, err := SessionContext(r)
	if err != nil {
		fmt.Printf("error obtaining session context: %s\n", err)
	}

	return sess
}
