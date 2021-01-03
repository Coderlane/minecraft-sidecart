package config

import (
	"io/ioutil"
	"reflect"
	"testing"
)

const testServerConfig = `
motd=A Minecraft Server
max-players=20
rcon.port=25575
enable-rcon=true
rcon.password=hunter2
server-ip=
server-port=25565`

func TestDefaultConfig(t *testing.T) {
	testPath := t.TempDir() + "/server.properties"
	ioutil.WriteFile(testPath, []byte(testServerConfig), 0600)
	config, err := ParseConfigFile(testPath)
	if err != nil {
		t.Fatal(err)
	}
	expectedConfig := Config{
		MotD:       "A Minecraft Server",
		MaxPlayers: 20,

		ServerPort: 25565,

		RCONEnabled:  true,
		RCONPassword: "hunter2",
		RCONPort:     25575,
	}
	if !reflect.DeepEqual(expectedConfig, *config) {
		t.Fatalf("Expected: %v Got: %v\n", expectedConfig, *config)
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
