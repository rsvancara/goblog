package requestfilter

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/oschwald/geoip2-golang"
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

// Search get geoip information from ipaddress
func (g *GeoIP) SearchCity(ipaddress string) error {

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

	if _, err := os.Stat("db/GeoIP2-City.mmdb"); os.IsNotExist(err) {
		g.IsFound = false
		return fmt.Errorf("error opening city geodatabase")
	}

	db, err := geoip2.Open("db/GeoIP2-City.mmdb")
	if err != nil {
		g.IsFound = false
		return fmt.Errorf("error opening city geodatabase")
	}
	defer db.Close()

	record, err := db.City(ip)
	if err != nil {
		g.IsFound = false
		return fmt.Errorf("error getting database record: %s", err)
	}

	// Each language is represented in a map
	g.City = record.City.Names["en"]

	// Each language is represented in a map
	g.CountryName = record.Country.Names["en"]

	g.CountryISOCode = record.Country.IsoCode

	g.IPAddress = ip

	g.TimeZone = record.Location.TimeZone

	g.IsProxy = record.Traits.IsAnonymousProxy

	g.IsEU = record.Country.IsInEuropeanUnion

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
