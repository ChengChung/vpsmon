package collector

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chengchung/vpsmon/sdk"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const namespace = "vpsmon"

var (
	ErrUnexpectedClientType = errors.New("unexpected client type")
	ErrClientNotFound       = errors.New("client not found")
	ErrUnknownCollector     = errors.New("unknown collector")
	// ErrNoData indicates the collector found no data to collect, but had no other error.
	ErrNoData = errors.New("collector returned no data")

	scrapeDurationDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_duration_seconds"),
		"vpsmon: Duration of a collector scrape.",
		[]string{"collector"},
		nil,
	)
	scrapeSuccessDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "scrape", "collector_success"),
		"vpsmon: Whether a collector succeeded.",
		[]string{"collector"},
		nil,
	)
)

func IsNoDataError(err error) bool {
	return err == ErrNoData
}

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

type factoryFunc func(client sdk.SDKClient, collectors []string) (Collector, error)

var factories = make(map[string]factoryFunc)

func registerCollectorFactory(name string, factory factoryFunc) {
	factories[name] = factory
}

func NewCollector(name string, client sdk.SDKClient, collectors []string) (Collector, error) {
	factory, ok := factories[name]
	if !ok {
		logrus.Errorf("unknown collector: %s", name)
		return nil, ErrUnknownCollector
	}
	return factory(client, collectors)
}

var (
	runningCollectors        = map[string]Collector{}
	runningCollectorsCounter = map[string]int{}
)

func InitCollectors(cfgs []CollectorConfig, clientGetter func(string) sdk.SDKClient) error {
	for _, cfg := range cfgs {
		client := clientGetter(cfg.ClientRef)
		if client == nil {
			logrus.Errorf("client not found: %s", cfg.ClientRef)
			return ErrClientNotFound
		}
		typeName := client.Type()
		collector, err := NewCollector(typeName, client, cfg.Collectors)
		if err != nil {
			logrus.Errorf("failed to create collector: %s", err)
			return err
		}
		runningCollectorsCounter[typeName] = runningCollectorsCounter[typeName] + 1
		collectorName := fmt.Sprintf("%s-%d", typeName, runningCollectorsCounter[typeName])
		runningCollectors[collectorName] = collector
		logrus.Infof("collector initialized: %s", collectorName)
	}

	return nil
}

type CollectorConfig struct {
	ClientRef  string   `yaml:"clientRef"`
	Collectors []string `yaml:"collectors"`
}

func NewVPSMonCollector() (prometheus.Collector, error) {
	return &VPSMonCollector{Collectors: runningCollectors}, nil
}

type VPSMonCollector struct {
	Collectors map[string]Collector
}

// Describe implements the prometheus.Collector interface.
func (n VPSMonCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc
	ch <- scrapeSuccessDesc
}

// Collect implements the prometheus.Collector interface.
func (n VPSMonCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(n.Collectors))
	for name, c := range n.Collectors {
		go func(name string, c Collector) {
			defer wg.Done()
			execute(name, c, ch)
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success float64

	if err != nil {
		if IsNoDataError(err) {
			logrus.Debug("collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			logrus.Error("collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}
		success = 0
	} else {
		logrus.Debug("collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = 1
	}
	ch <- prometheus.MustNewConstMetric(scrapeDurationDesc, prometheus.GaugeValue, duration.Seconds(), name)
	ch <- prometheus.MustNewConstMetric(scrapeSuccessDesc, prometheus.GaugeValue, success, name)
}
