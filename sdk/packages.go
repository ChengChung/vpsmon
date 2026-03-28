package sdk

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type clientFactory interface {
	New(string, map[string]string) (SDKClient, error)
}

var factoryMap = make(map[string]clientFactory)

func RegisterClientFactory(name string, factory clientFactory) {
	factoryMap[name] = factory
}

type SDKClient interface {
	Name() string
	Type() string
}

func NewClient(name string, typeName string, cfg map[string]string) (SDKClient, error) {
	factory, ok := factoryMap[typeName]
	if !ok {
		return nil, fmt.Errorf("client factory not found for name: %s", typeName)
	}
	return factory.New(name, cfg)
}

func InitClients(cfgs map[string]ClientConfig) error {
	for name, cfg := range cfgs {
		client, err := NewClient(name, cfg.Type, cfg.Config)
		if err != nil {
			logrus.Errorf("failed to create client %s: %s", name, err)
			return err
		}
		clients[name] = client
	}

	return nil
}

func GetClient(name string) SDKClient {
	return clients[name]
}

var clients = map[string]SDKClient{}

type ClientConfig struct {
	Type   string            `yaml:"type"`
	Config map[string]string `yaml:"config"`
}
