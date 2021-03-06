package covid

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var BaseUrl = "https://services1.arcgis.com"

var Path = "/0MSEUqKaxRlEPj5g/arcgis/rest/services/ncov_cases/FeatureServer/2/query"

type Response struct {
	Values []Case `json:"features"`
}
type Case struct {
	Attrs struct {
		Id         int     `json:"OBJECTID"`
		Country    string  `json:"Country_Region"`
		LastUpdate int     `json:"Last_Update"`
		Latitude   float64 `json:"Lat"`
		Longitude  float64 `json:"Long_"`
		Confirmed  int     `json:"Confirmed"`
		Active     int     `json:"Active"`
		Recovered  int     `json:"Recovered"`
		Deaths     int     `json:"Deaths"`
	} `json:"attributes"`
}

type CountryResponse struct {
	Values []Country `json:"features"`
}

type Country struct {
	Attrs struct {
		Id   int    `json:"OBJECTID"`
		Name string `json:"Country_Region"`
	} `json:"attributes"`
}
type StatsResponse struct {
	Values []Stats `json:"features"`
}

type Stats struct {
	Attrs struct {
		Value int `json:"value"`
	} `json:"attributes"`
}

func getAllUrl() string {
	_url, _ := url.Parse(BaseUrl)
	_url.Path += Path
	parameters := url.Values{}
	parameters.Add("f", "json")
	parameters.Add("where", "Confirmed > 0")
	parameters.Add("returnGeometry", "false")
	parameters.Add("spatialRel", "esriSpatialRelIntersects")
	parameters.Add("outFields", "*")
	parameters.Add("orderByFields", "Confirmed desc")
	parameters.Add("resultOffset", "0")
	parameters.Add("resultRecordCount", "200")
	parameters.Add("cacheHint", "true")
	_url.RawQuery = parameters.Encode()
	return _url.String()
}

func getCountryUrl(id int) string {
	_url, _ := url.Parse(BaseUrl)
	_url.Path += Path
	parameters := url.Values{}
	parameters.Add("f", "json")
	parameters.Add("where", "OBJECTID ="+strconv.Itoa(id))
	parameters.Add("returnGeometry", "false")
	parameters.Add("spatialRel", "esriSpatialRelIntersects")
	parameters.Add("outFields", "*")
	parameters.Add("orderByFields", "Confirmed desc")
	parameters.Add("resultOffset", "0")
	parameters.Add("resultRecordCount", "200")
	parameters.Add("cacheHint", "true")
	_url.RawQuery = parameters.Encode()
	return _url.String()
}

func getTotalUrl(field string) string {
	_url, _ := url.Parse(BaseUrl)
	_url.Path += Path
	parameters := url.Values{}
	parameters.Add("f", "json")
	parameters.Add("where", "Confirmed > 0")
	parameters.Add("returnGeometry", "false")
	parameters.Add("spatialRel", "esriSpatialRelIntersects")
	parameters.Add("outFields", "*")
	parameters.Add("outStatistics", fmt.Sprintf("[{\"statisticType\":\"sum\",\"onStatisticField\":\"%s\",\"outStatisticFieldName\":\"value\"}]", field))
	parameters.Add("cacheHint", "true")
	_url.RawQuery = parameters.Encode()
	return _url.String()
}

// GetData returns the entire data of every country
func GetData() ([]Case, error) {
	resp, err := http.Get(getAllUrl())
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return res.Values, nil
}

// GetCountryById returns the statistics/data of
// a particular country specified by id
func GetCountryById(id int) (Case, error) {
	resp, err := http.Get(getCountryUrl(id))
	if err != nil {
		return Case{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Case{}, err
	}
	defer resp.Body.Close()

	var res Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return Case{}, err
	}

	return res.Values[0], nil
}

// ListCountries returns a "list" or slice of countries, that shows the id and name
// of a country as understood by this api. This method is handy when you want
// the ID or name of a country to use with methods GetCountryById and GetCountryByName
// respectively
func ListCountries() ([]Country, error) {
	resp, err := http.Get(getAllUrl())
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var res CountryResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	return res.Values, nil
}

func getCountryId(name string) (int, error) {
	countries, err := ListCountries()
	if err != nil {
		return -1, err
	}

	for _, country := range countries {
		if strings.ToLower(country.Attrs.Name) == strings.ToLower(name) {
			return country.Attrs.Id, nil
		}
	}
	return -1, errors.New("Wrong country name")
}

// GetCountryByName returns the statistics/data of
// a particular country as supplied by name
func GetCountryByName(name string) (Case, error) {
	id, err := getCountryId(name)
	if err != nil {
		return Case{}, err
	}

	resp, err := http.Get(getCountryUrl(id))
	if err != nil {
		return Case{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Case{}, err
	}
	defer resp.Body.Close()

	var res Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return Case{}, err
	}

	return res.Values[0], nil
}

func getTotal(field string) (Stats, error) {
	resp, err := http.Get(getTotalUrl(field))
	if err != nil {
		return Stats{}, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Stats{}, err
	}
	defer resp.Body.Close()

	var res StatsResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		return Stats{}, err
	}

	return res.Values[0], nil
}

// GetTotalActive returns the total number of active cases globally
func GetTotalActive() (int, error) {
	data, err := getTotal("Active")
	if err != nil {
		return -1, err
	}
	return data.Attrs.Value, nil
}


// GetTotalConfirmed returns the total number of confirmed cases globally
func GetTotalConfirmed() (int, error) {
	data, err := getTotal("Confirmed")
	if err != nil {
		return -1, err
	}
	return data.Attrs.Value, err
}

// GetTotalRecovered returns the total number of recovered cases globally
func GetTotalRecovered() (int, error) {
	data, err := getTotal("Recovered")
	if err != nil {
		return -1, err
	}
	return data.Attrs.Value, nil
}

// GetTotalDeaths returns the total number of deaths recorded globally
func GetTotalDeaths() (int, error) {
	data, err := getTotal("Deaths")
	if err != nil {
		return -1, err
	}
	return data.Attrs.Value, nil
}
