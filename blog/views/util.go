package views

import (
	"blog/blog/requestfilter"
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

	return geoIP, fmt.Errorf("unable to find context for geoip")
}
