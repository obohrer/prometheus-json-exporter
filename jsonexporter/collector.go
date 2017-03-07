package jsonexporter

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/kawamuray/jsonpath" // Originally: "github.com/NickSardo/jsonpath"
	"github.com/kawamuray/prometheus-exporter-harness/harness"
	"io/ioutil"
	"net/http"
)

type Collector struct {
	Endpoints []Endpoint
	scrapers []JsonScraper
}

func compilePath(path string) (*jsonpath.Path, error) {
	// All paths in this package is for extracting a value.
	// Complete trailing '+' sign if necessary.
	if path[len(path)-1] != '+' {
		path += "+"
	}

	paths, err := jsonpath.ParsePaths(path)
	if err != nil {
		return nil, err
	}
	return paths[0], nil
}

func compilePaths(paths map[string]string) (map[string]*jsonpath.Path, error) {
	compiledPaths := make(map[string]*jsonpath.Path)
	for name, value := range paths {
		if len(value) < 1 || value[0] != '$' {
			// Static value
			continue
		}
		compiledPath, err := compilePath(value)
		if err != nil {
			return nil, fmt.Errorf("failed to parse path;path:<%s>,err:<%s>", value, err)
		}
		compiledPaths[name] = compiledPath
	}
	return compiledPaths, nil
}

func NewCollector(endpoints []Endpoint, scrapers []JsonScraper) *Collector {
	return &Collector{
		Endpoints: endpoints,
		scrapers: scrapers,
	}
}

func (col *Collector) fetchJson(endpoint Endpoint) ([]byte, error) {
	url := endpoint.URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch json from endpoint;endpoint:<%s>,err:<%s>", url, err)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body;err:<%s>", err)
	}

	return data, nil
}

func (col *Collector) Collect(reg *harness.MetricRegistry) {
	for _, endpoint := range col.Endpoints {
		json, err := col.fetchJson(endpoint)
		if err != nil {
			log.Error(err)
			return
		}

		for i := 0; i < len(col.scrapers); i++ {
			if err := col.scrapers[i].Scrape(json, endpoint, reg); err != nil {
				log.Errorf("error while scraping json from;err:<%s>;endpoint:<%s>", err, endpoint.URL)
				continue
			}
	  }
	}
}
