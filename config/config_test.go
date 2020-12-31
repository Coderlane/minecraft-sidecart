package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config, err := ParseConfigFile("../sidecart_config.yml")
	if err != nil {
		t.Fatal(err)
	}
	if config.RCON.Address != "localhost:25575" {
		t.Errorf("Unexpected address. Got: %s Expected: %s\n",
			config.RCON.Address, "localhost:25575")
	}
}

func TestMissingConfig(t *testing.T) {
	_, err := ParseConfigFile("not_a_file.yml")
	if err == nil {
		t.Fatalf("Expected an error.")
	}
}

func TestCorruptConfig(t *testing.T) {
	_, err := ParseConfig([]byte("bad_data"))
	if err == nil {
		t.Fatalf("Expected an error.")
	}
}
