package main

import (
	"github.com/chengchung/vpsmon/collector"
	"github.com/chengchung/vpsmon/sdk"
)

type Config struct {
	//	name->config_map
	Clients    map[string]sdk.ClientConfig `yaml:"clients"`
	Collectors []collector.CollectorConfig `yaml:"collectors"`
}
