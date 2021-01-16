package middleware

import (
	"context"
	"net/http"
	"net/http/httputil"

	"github.com/rs/zerolog/log"
	"github.com/rsvancara/goblog/internal/models"
	"github.com/rsvancara/goblog/internal/requestfilter"
	"github.com/rsvancara/goblog/internal/util"

	requestviewdao "github.com/rsvancara/goblog/internal/dao/requestview"
)

// GeoFilterMiddleware Middleware that matches paths to filter rules.
func (mw *MiddleWareContext) GeoFilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var geoIP requestfilter.GeoIP

		ipaddress, _ := requestfilter.GetIPAddress(r)
		//fmt.Printf("IP Address: %s | request: %s\n", ipaddress, r.RequestURI)

		// for testing...we inject an IP Address
		//if ipaddress == "" {
		//	ipaddress = "73.83.74.114"
		//}

		geoIP.PageID = models.GenUUID()

		if ipaddress != "" && ipaddress != "[" {

			err := geoIP.Search(ipaddress)
			if err != nil {
				log.Error().Err(err).Str("service", "geotagger").Msgf("Error IP Address not found in the database for IP Address: %s in geotagger middleware", ipaddress)
			}

			if requestfilter.IsPrivateSubnet(geoIP.IPAddress) {
				// Handle situations where we have a private ip address
				// 	1. In development this is ok
				//  2. In production something should be considered wrong
				//  3. Send to capta page?
			}
			if geoIP.IsFound == true {
				// Apply filter rules
				// Filter on IP
				// Filter on City
				// Filter on Country
				// Filter on timezone
				// Filter on EU

				// Filters are based on request path,
				// path is matched to a list of rules in a database
				// and returned to be evaluated.
				// Based on the match condition, action is taken, allow, deny, redirect

			}
		}

		var ctxKey util.CtxKey
		ctxKey = "geoip"
		ctx := context.WithValue(r.Context(), ctxKey, geoIP)

		sess, err := util.SessionContext(r)
		if err != nil {
			log.Error().Err(err).Str("service", "geotagger").Msg("Error getting session for context in geotagger middleware")
		}

		var requestviewDAO requestviewdao.RequestViewDAO

		err = requestviewDAO.Initialize(mw.dbClient, mw.hConfig)
		if err != nil {
			log.Error().Err(err).Str("service", "requestviewdao").Msg("Error initialzing media data access object in geotagger middleware")
		}

		// Save a copy of this request for debugging.
		//TODO: find some library that can produce sain output of the the request that can be stored in a database
		//TODO: Move this into another filter that can be chained with this one....enrich enrich...filter
		requestDump, err := httputil.DumpRequest(r, false)
		if err != nil {
			log.Error().Err(err).Str("service", "geotagger").Msg("Error getting request object in geotagger middleware")
		}
		rawRequest := string(requestDump)

		var rv models.RequestView
		rv.IPAddress = geoIP.IPAddress.String()
		rv.HeaderUserAgent = r.Header.Get("User-Agent")
		rv.PTag = geoIP.PageID
		rv.RequestURL = r.RequestURI
		rv.SessionID = sess.SessionToken
		rv.City = geoIP.City
		rv.Country = geoIP.CountryName
		rv.RawRequest = rawRequest
		err = requestviewDAO.CreateRequestView(&rv)
		if err != nil {
			log.Error().Err(err).Str("service", "requestviewdao").Msg("Error creating requestview data access object in geotagger middleware")
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
