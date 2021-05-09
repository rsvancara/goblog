package requestfilter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/oschwald/geoip2-golang"

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
	message     string      `json:"message"`
	isError     bool        `json:"is_error"`
	geoLocation GeoLocation `json:"geo_location"`
}

type GeoLocation struct {
	isFound        bool   `json:"is_found"`
	isPrivate      bool   `json:"is_private"`
	ipAddr         string `json:"ip_addr"`
	city           string `json:"city"`
	countryName    string `json:"country_name"`
	countryISOCode string `json:"country_iso_code"`
	timeZone       string `json:"time_zone"`
	isProxy        bool   `json:"is_proxy"`
	isEU           bool   `json:"is_eu"`
	asn            int    `json:"ans"`
	organization   string `json:"organization"`
	network        string `json:"network"`
}

// Search get geoip information from ipaddress
func (g *GeoIP) SearchCity(ipaddress string, config config.AppConfig) error {

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

	geoServiceURI := config.GeoService

	response, err := http.Get(geoServiceURI + ip.String())

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

	if geoipMessage.isError == false {
		g.IsFound = false
		log.Error().Err(fmt.Errorf(geoipMessage.message)).Str("service", "GeoService").Msg("The geocode service experienced an error")
		return fmt.Errorf("The geocode service experienced an error %s", geoipMessage.message)
	}

	geoLoc := geoipMessage.geoLocation

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
	g.City = geoLoc.city

	// Each language is represented in a map
	g.CountryName = geoLoc.countryName

	g.CountryISOCode = geoLoc.countryISOCode

	g.IPAddress = ip

	g.TimeZone = geoLoc.timeZone

	g.IsProxy = geoLoc.isProxy

	g.IsEU = geoLoc.isEU

	elapsed := time.Since(start)
	log.Printf("geoipa took %s \n", elapsed)

	return nil

}

// Search get geoip information from ipaddress
func (g *GeoIP) SearchASN(ipaddress string) error {
	start := time.Now()

	if ipaddress == "::1" {
		g.Organization = "None"
		g.ASN = "None"
		return nil
	}

	if ipaddress == "127.0.0.1" {
		g.Organization = "None"
		g.ASN = "None"
		return nil
	}

	ip := net.ParseIP(ipaddress)
	if ip == nil {
		g.IsFound = false
		return fmt.Errorf("error converting string [ %s ] to IP Address", ipaddress)
	}

	if _, err := os.Stat("db/GeoLite2-ASN.mmdb"); os.IsNotExist(err) {
		g.IsFound = false
		return fmt.Errorf("error opening ASN geodatabase")
	}

	dbasn, err := geoip2.Open("db/GeoLite2-ASN.mmdb")
	if err != nil {
		g.IsFound = false
		return fmt.Errorf("error opening ASN geodatabase")
	}
	defer dbasn.Close()

	record, err := dbasn.ASN(ip)
	if err != nil {
		g.IsFound = false
		return fmt.Errorf("error getting database record: %s", err)
	}

	g.Organization = record.AutonomousSystemOrganization
	g.ASN = fmt.Sprint(record.AutonomousSystemNumber)

	//fmt.Println(g)

	elapsed := time.Since(start)
	log.Printf("geoipa took %s \n", elapsed)

	return nil
}
