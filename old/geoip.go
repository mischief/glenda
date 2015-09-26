package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kballard/goirc/irc"
)

func init() {
	RegisterModule("geoip", func() Module {
		return &GeoipMod{}
	})
}

//{"ip":"76.126.11.172","country_code":"US","country_name":"United States",
// "region_code":"CA","region_name":"California","city":"San Francisco",
// "zipcode":"94110","latitude":37.7484,"longitude":-122.4156,"metro_code":"807","areacode":"415"}

/*
{
    "city": "San Francisco",
    "country_code": "US",
    "country_name": "United States",
    "ip": "192.30.252.131",
    "latitude": 37.77,
    "longitude": -122.394,
    "metro_code": 807,
    "region_code": "CA",
    "region_name": "California",
    "time_zone": "America/Los_Angeles",
    "zip_code": "94107"
}

*/
type FreeGeoip struct {
	Ip           string
	Country_code string
	Country_name string
	Region_code  string
	Region_name  string
	City         string
	Zipcode      string
	Latitude     float64
	Longitude    float64
	Metro_code   float64
	Areacode     string
}

type GeoipMod struct {
	urlfmt string
}

func (g *GeoipMod) Init(b *Bot, conn irc.SafeConn) error {
	g.urlfmt = "http://freegeoip.net/%s/%s"

	b.Hook("geo", func(b *Bot, sender, cmd string, args ...string) error {
		if len(args) != 1 {
			return nil
		}

		b.Conn.Privmsg(sender, g.geo(args[0]))
		return nil
	})

	log.Printf("geoip module initialized with urlfmt %s", g.urlfmt)
	return nil
}

func (g *GeoipMod) Reload() error {
	return nil
}

func (g *GeoipMod) Call(args ...string) error {
	return nil
}

// return human readable form of geoip data
func (g *GeoipMod) geo(ip string) string {
	var geo FreeGeoip
	var body []byte

	url := fmt.Sprintf(g.urlfmt, "json", ip)

	resp, err := http.Get(url)
	if err != nil {
		goto bad
	}

	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		goto bad
	}

	if err = json.Unmarshal(body, &geo); err != nil {
		goto bad
	}

	return fmt.Sprintf("%s: %s - %s - %s : φ%f° λ%f°",
		geo.Ip, geo.Country_name, geo.Region_name, geo.City, geo.Latitude, geo.Longitude)

bad:
	return fmt.Sprintf("geoip error: %s", err)
}
