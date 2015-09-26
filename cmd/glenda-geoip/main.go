package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/mischief/glenda/util"

	"github.com/koding/kite"
)

func main() {
	k := util.NewKite("geoip", "1.0.0")
	k.HandleFunc("geoip.geoip", geoip).DisableAuthentication()
	<-k.ServerCloseNotify()
}

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

// return human readable form of geoip data
func geoip(r *kite.Request) (result interface{}, err error) {
	var geo FreeGeoip
	var body []byte

	args := util.ArgSlice(r)
	if len(args) < 1 {
		return nil, fmt.Errorf("no query")
	}

	ip := args[0]

	url := fmt.Sprintf("http://freegeoip.net/%s/%s", "json", ip)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("geoip error: %s", err)
	}

	defer resp.Body.Close()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("geoip error: %s", err)
	}

	if err = json.Unmarshal(body, &geo); err != nil {
		return nil, fmt.Errorf("geoip error: %s", err)
	}

	out := fmt.Sprintf("%s: %s - %s - %s : φ%f° λ%f°", geo.Ip, geo.Country_name, geo.Region_name, geo.City, geo.Latitude, geo.Longitude)

	return out, nil
}
