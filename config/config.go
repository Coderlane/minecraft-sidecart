package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type RCONConfig struct {
	Address  string
	Password string
}

type Config struct {
	RCON RCONConfig
}

func ParseConfigFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseConfig(data)
}

func ParseConfig(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
