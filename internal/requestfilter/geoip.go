package requestfilter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"goblog/internal/config"

	"github.com/rs/zerolog/log"
)

// GeoIP Object
type GeoIP struct {
	IsFound        bool
	IsPrivate      bool
	IPAddress      net.IP
	City           string
	CountryName    string
	CountryISOCode string
	TimeZone       string
	IsProxy        bool
	IsEU           bool
	ASN            string
	Organization   string
	Network        string
	PageID         string
}

type GeoIPMessage struct {
	Message     string      `json:"message"`
	IsError     bool        `json:"is_error"`
	GeoLocation GeoLocation `json:"geo_location"`
}

type GeoLocation struct {
	IsFound        bool   `json:"is_found"`
	IsPrivate      bool   `json:"is_private"`
	IpAddr         string `json:"ip_addr"`
	City           string `json:"city"`
	CountryName    string `json:"country_name"`
	CountryISOCode string `json:"country_iso_code"`
	TimeZone       string `json:"time_zone"`
	IsProxy        bool   `json:"is_proxy"`
	IsEU           bool   `json:"is_eu"`
	ASN            int    `json:"ans"`
	Organization   string `json:"organization"`
	Network        string `json:"network"`
}

// Search get geoip information from ipaddress
func (g *GeoIP) GeoIPSearch(ipaddress string, config config.AppConfig) error {

	start := time.Now()

	if ipaddress == "::1" {
		g.CountryISOCode = "US"
		g.CountryName = "Merica"
		g.IsEU = false
		g.IsPrivate = true
		g.IPAddress = net.IPv6loopback
		g.City = "Boise"
		return nil
	}

	if ipaddress == "127.0.0.1" {
		g.CountryISOCode = "US"
		g.CountryName = "Merica"
		g.IsEU = false
		g.IsPrivate = true
		g.IPAddress = net.ParseIP("127.0.0.1")
		g.City = "Boise"
		return nil
	}

	ip := net.ParseIP(ipaddress)
	if ip == nil {
		g.IsFound = false
		return fmt.Errorf("error converting string [ %s ] to IP Address", ipaddress)
	}

	log.Info().Msgf("Searching for IP %s", ip.String())

	geoServiceURI := fmt.Sprintf("%s%s", strings.TrimSuffix(config.GeoService, "\n"), ip.String())

	response, err := http.Get(geoServiceURI)

	if err != nil {
		g.IsFound = false
		log.Error().Err(err).Str("service", "GeoService").Msgf("Error querying [%s]", geoServiceURI)
		return fmt.Errorf("Error querying [%s] with error %s", geoServiceURI, err)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		g.IsFound = false
		log.Error().Err(err).Str("service", "GeoService").Msgf("Error getting response data from geoservice  for URI [%s]", geoServiceURI)
		return fmt.Errorf("Error getting response data from geoservice for uri [%s] with error %s", geoServiceURI, err)
	}

	var geoipMessage GeoIPMessage
	err = json.Unmarshal(responseData, &geoipMessage)
	if err != nil {
		g.IsFound = false
		log.Error().Err(err).Str("service", "GeoService").Msgf("Error unmarshalling responsed data for URI [%s]", geoServiceURI)
		return fmt.Errorf("Error unmarshalling responsed data for uri [%s] with error %s", geoServiceURI, err)
	}

	fmt.Println(geoipMessage)

	if geoipMessage.IsError == true {
		g.IsFound = false
		log.Error().Err(fmt.Errorf(geoipMessage.Message)).Str("service", "GeoService").Msg("The geocode service experienced an error")
		return fmt.Errorf("The geocode service experienced an error %s", geoipMessage.Message)
	}

	geoLoc := geoipMessage.GeoLocation

	//if _, err := os.Stat("db/GeoIP2-City.mmdb"); os.IsNotExist(err) {
	//	g.IsFound = false
	//	return fmt.Errorf("error opening city geodatabase")
	//}

	//db, err := geoip2.Open("db/GeoIP2-City.mmdb")
	//if err != nil {
	//	g.IsFound = false
	//	return fmt.Errorf("error opening city geodatabase")
	//}
	//defer db.Close()

	//record, err := db.City(ip)
	//if err != nil {
	//	g.IsFound = false
	//	return fmt.Errorf("error getting database record: %s", err)
	//}

	// Each language is represented in a map
	g.City = geoLoc.City

	// Each language is represented in a map
	g.CountryName = geoLoc.CountryName

	g.CountryISOCode = geoLoc.CountryISOCode

	g.IPAddress = ip

	g.TimeZone = geoLoc.TimeZone

	g.IsProxy = geoLoc.IsProxy

	g.IsEU = geoLoc.IsEU

	g.Organization = geoLoc.Organization

	g.ASN = strconv.Itoa(geoLoc.ASN)

	elapsed := time.Since(start)
	log.Printf("geoipa took %s \n", elapsed)

	return nil

}
