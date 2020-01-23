package requestfilter

import (
	"fmt"
	"net"
	"os"

	"github.com/oschwald/geoip2-golang"
)

// GeoIP Object
type GeoIP struct {
	IPAddress      string
	City           string
	CountryName    string
	CountryISOCode string
	TimeZone       string
	IsProxy        bool
	IsEU           bool
}

// Search get geoip information from ipaddress
func (g *GeoIP) Search(ipaddress string) error {

	ip := net.ParseIP(ipaddress)
	if ip == nil {
		return fmt.Errorf("Error converting string to IP Address")
	}

	if _, err := os.Stat("db/GeoIP2-City.mmdb"); os.IsNotExist(err) {
		return fmt.Errorf("Error opening city geodatabase")
	}

	db, err := geoip2.Open("db/GeoIP2-City.mmdb")
	if err != nil {
		return fmt.Errorf("Error opening country geodatabase")
	}
	defer db.Close()

	record, err := db.City(ip)

	// Each language is represented in a map
	g.City = record.City.Names["en"]

	// Each language is represented in a map
	g.CountryName = record.Country.Names["en"]

	g.CountryISOCode = record.Country.IsoCode

	g.IPAddress = ipaddress

	g.TimeZone = record.Location.TimeZone

	g.IsProxy = record.Traits.IsAnonymousProxy

	g.IsEU = record.Country.IsInEuropeanUnion

	return nil

}
