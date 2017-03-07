package jsonexporter

import (
	"fmt"
	"github.com/urfave/cli"
	"github.com/kawamuray/prometheus-exporter-harness/harness"
	"github.com/prometheus/client_golang/prometheus"
)

type ScrapeType struct {
	Configure  func(*Config, Endpoint, *harness.MetricRegistry)
	NewScraper func(*Config, Endpoint) (JsonScraper, error)
}

var ScrapeTypes = map[string]*ScrapeType{
	"object": {
		Configure: func(config *Config, endpoint Endpoint,reg *harness.MetricRegistry) {
			for subName := range config.Values {
				name := harness.MakeMetricName(endpoint.Prefix + config.Name, subName)
				reg.Register(
					name,
					prometheus.NewGaugeVec(prometheus.GaugeOpts{
						Name: name,
						Help: config.Help + " - " + subName,
					}, config.labelNames()),
				)
			}
		},
		NewScraper: NewObjectScraper,
	},
	"value": {
		Configure: func(config *Config, endpoint Endpoint, reg *harness.MetricRegistry) {
			reg.Register(
				endpoint.Prefix + config.Name,
				prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Name: config.Name,
					Help: config.Help,
				}, config.labelNames()),
			)
		},
		NewScraper: NewValueScraper,
	},
}

var DefaultScrapeType = "value"

func Init(c *cli.Context, reg *harness.MetricRegistry) (harness.Collector, error) {
	args := c.Args()

	if len(args) < 2 {
		cli.ShowAppHelp(c)
		return nil, fmt.Errorf("not XXenough arguments")
	}

	var (
		endpointsPath = args[0]
		configPath = args[1]
	)

	configs, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	endpoints, err := loadEndpoints(endpointsPath)
	fmt.Errorf("Endpoints : <%s>", endpoints)
	if err != nil {
		return nil, err
	}
  scrapesCount := len(configs)
	scrapers := make([]JsonScraper, scrapesCount* len(endpoints))
	for j, endpoint := range endpoints {
		for i, config := range configs { // apply the same configuration to all endpoints
			tpe := ScrapeTypes[config.Type]
			if tpe == nil {
				return nil, fmt.Errorf("unknown scrape type;type:<%s>", config.Type)
			}
			tpe.Configure(config, endpoint, reg)
			scraper, err := tpe.NewScraper(config, endpoint)
			if err != nil {
				return nil, fmt.Errorf("failed to create scraper;name:<%s>,err:<%s>", config.Name, err)
			}
			scrapers[j*scrapesCount + i] = scraper
		}
	}

	return NewCollector(endpoints, scrapers), nil
}
