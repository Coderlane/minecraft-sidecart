package config

import (
	"io/ioutil"

	"github.com/magiconair/properties"
)

type Config struct {
	ServerIP   string `properties:"server-ip"`
	ServerPort int    `properties:"server-port"`

	RCONEnabled  bool   `properties:"enable-rcon"`
	RCONPassword string `properties:"rcon.password"`
	RCONPort     int    `properties:"rcon.port"`
}

func ParseConfigFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseConfig(data)
}

func ParseConfig(data []byte) (*Config, error) {
	prop, err := properties.Load(data, properties.UTF8)
	if err != nil {
		return nil, err
	}
	var config Config
	err = prop.Decode(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
