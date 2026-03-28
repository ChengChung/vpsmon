package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/chengchung/vpsmon/collector"
	"github.com/chengchung/vpsmon/sdk"
	_ "github.com/chengchung/vpsmon/sdk/imports"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	confPath    = flag.String("conf", "/etc/vpsmon/config.yml", "path to config file")
	metricsPath = flag.String("metrics-path", "/metrics", "path to expose metrics")
	listenAddr  = flag.String("listen-addr", ":23874", "address to listen on for HTTP requests")
)

func main() {
	flag.Parse()

	Init()

	handler, err := insertHandler()
	if err != nil {
		logrus.Panicf("cannot create metrics handler %s", err)
	}
	mux := http.NewServeMux()
	mux.Handle(*metricsPath, handler)

	logrus.Infof("start server at %s", *listenAddr)
	if err := http.ListenAndServe(*listenAddr, mux); err != nil {
		logrus.Error(err.Error())
		os.Exit(1)
	}
}

func Init() {
	data, err := os.ReadFile(*confPath)
	if err != nil {
		logrus.Fatalf("fail to parse config %s", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		logrus.Fatalf("fail to parse config %s", err)
	}

	if err := sdk.InitClients(cfg.Clients); err != nil {
		logrus.Fatalf("fail to init clients: %s", err)
	}

	if err := collector.InitCollectors(cfg.Collectors, sdk.GetClient); err != nil {
		logrus.Fatalf("fail to init collectors: %s", err)
	}

	logrus.Infof("init config success")
}

func insertHandler() (http.Handler, error) {
	r := prometheus.NewRegistry()
	r.MustRegister(version.NewCollector("vpsmon"))
	vpscolls, err := collector.NewVPSMonCollector()
	if err != nil {
		return nil, fmt.Errorf("cannot create collectors: %s", err)
	}
	if err := r.Register(vpscolls); err != nil {
		return nil, fmt.Errorf("cannot register vpsmon collectors: %s", err)
	}

	return promhttp.HandlerFor(
		r,
		promhttp.HandlerOpts{
			ErrorLog:            logrus.StandardLogger(),
			ErrorHandling:       promhttp.ContinueOnError,
			MaxRequestsInFlight: 3,
		},
	), nil
}
