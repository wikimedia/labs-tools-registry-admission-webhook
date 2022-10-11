package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ListenOn   string   `default:"0.0.0.0:8080"`
	TLSCert    string   `default:"/etc/webhook/certs/cert.pem"`
	TLSKey     string   `default:"/etc/webhook/certs/key.pem"`
	Registries []string `default:"docker-registry.tools.wmflabs.org"`
	Debug      bool     `default:"true"`
	BuildID    string   `default:"nobuildid"`
}

func GetConfigFromEnv() (*Config, error) {
	config := &Config{}
	envconfig.Process("", config)
	if len(config.Registries) < 1 {
		return nil, fmt.Errorf(
			"got no registries, at least one is required, make sure to set the REGISTRIES env var to a comma separated list of registries (ex. 'docker-registry.tools.wmflabs.org', or 'registry1,registry2')",
		)
	}
	return config, nil
}
