package views

import (
	"blog/blog/requestfilter"
	"blog/blog/session"
	"blog/blog/util"
	"fmt"
	"net/http"
)

//GeoIPContext get the geoip object
func GeoIPContext(r *http.Request) (requestfilter.GeoIP, error) {

	var geoIP requestfilter.GeoIP

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
			return geoIP, fmt.Errorf("could not perform type assertion on result to GeoIP type for ctxKey %s", ctxKey)
		}

		return geoIP, nil

	}

	return geoIP, fmt.Errorf("unable to find context for geoip %s", ctxKey)
}

//SessionContext get the session object
func SessionContext(r *http.Request) (session.Session, error) {

	var sess session.Session

	// Attempt to extract additional information from a context
	//var geoIP requestfilter.GeoIP
	var ctxKey util.CtxKey
	ctxKey = "session"

	if result := r.Context().Value(ctxKey); result != nil {

		//fmt.Println("Found context")
		//fmt.Println(result)
		// Type Assertion....
		sess, ok := result.(session.Session)
		if !ok {
			return sess, fmt.Errorf("could not perform type assertion on result to session.Session type for ctxKey %s", ctxKey)
		}

		return sess, nil

	}

	return sess, fmt.Errorf("unable to find context for session  %s on page %s", ctxKey, r.RequestURI)
}

// GetPageID get the page ID for a request
func GetPageID(r *http.Request) string {
	geoIP, err := GeoIPContext(r)
	if err != nil {
		fmt.Printf("error obtaining geoip context: %s\n", err)
	}

	return geoIP.PageID
}

// GetSession get session object for a request
func GetSession(r *http.Request) session.Session {
	sess, err := SessionContext(r)
	if err != nil {
		fmt.Printf("error obtaining session context: %s\n", err)
	}

	return sess
}
