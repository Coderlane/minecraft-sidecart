package server

import (
	"io/ioutil"
	"path"
	"testing"

	config "github.com/Coderlane/go-minecraft-config"
	//config "github.com/Coderlane/go-minecraft-config"
)

const testServerProperties = `
rcon.port=25575
gamemode=survival
enable-query=true
motd=A Minecraft Server
query.port=25565
`
const testDenyIPs = `[
  {
    "ip": "127.0.0.134",
    "created": "2020-09-14 23:05:05 -0400",
    "source": "Server",
    "expires": "forever",
    "reason": "Deny by an operator."
  }
]`

const testDenyUsers = `[
  {
    "uuid": "9b15dea6-606e-47a4-a241-000000000000",
    "name": "Test",
    "created": "2020-09-14 23:01:51 -0400",
    "source": "Server",
    "expires": "forever",
    "reason": "Banned by an operator."
  }
]`

const testAllowUsers = `[
  {
    "uuid": "9b15dea6-606e-47a4-a241-000000000000",
    "name": "Test"
  }
]`

const testOPs = `[
  {
    "uuid": "9b15dea6-606e-47a4-a241-000000000000",
    "name": "Test",
    "level": 4,
    "bypassesPlayerLimit": false
  }
]`

func writeConfigFile(t *testing.T, dir, file, data string) {
	err := ioutil.WriteFile(path.Join(dir, file), []byte(data), 0600)
	if err != nil {
		t.Fatal(err)
	}
}

func setupConfigFiles(t *testing.T, dir string) {
	writeConfigFile(t, dir, config.MinecraftConfigFile, testServerProperties)
	writeConfigFile(t, dir, config.MinecraftAllowUserFile, testAllowUsers)
	writeConfigFile(t, dir, config.MinecraftDenyIPFile, testDenyIPs)
	writeConfigFile(t, dir, config.MinecraftDenyUserFile, testDenyUsers)
	writeConfigFile(t, dir, config.MinecraftOperatorUserFile, testOPs)
}

func TestMinecraftConfigWatcher(t *testing.T) {
	testDir := t.TempDir()
	setupConfigFiles(t, testDir)

	type testCase struct {
		File     string
		Contents string

		Type minecraftConfigType
	}
	testCases := []testCase{
		{config.MinecraftConfigFile, testServerProperties, minecraftConfigTypeServer},
		{config.MinecraftDenyIPFile, testDenyIPs, minecraftConfigTypeDenyIP},
		{config.MinecraftDenyUserFile, testDenyUsers, minecraftConfigTypeDenyUser},
		{config.MinecraftOperatorUserFile, testOPs, minecraftConfigTypeOperatorUser},
		{config.MinecraftAllowUserFile, testAllowUsers, minecraftConfigTypeAllowUser},
	}

	for _, tc := range testCases {
		t.Run(tc.File, func(t *testing.T) {

			watcher, err := newMinecraftConfigWatcher(testDir)
			if err != nil {
				t.Fatal(err)
			}
			defer watcher.Stop()

			writeConfigFile(t, testDir, tc.File, tc.Contents)
			var cfgEvent configEvent
			select {
			case cfgEvent = <-watcher.ConfigEvents:
				t.Log(cfgEvent)
			case err = <-watcher.Errors:
				t.Error(err)
			}
			if tc.Type != cfgEvent.Type {
				t.Errorf("Expected: %v Got %v\n", tc.Type, cfgEvent.Type)
			}
		})
	}
}

func TestMinecraftConfigWatcherInvalidConfig(t *testing.T) {
	testDir := t.TempDir()
	setupConfigFiles(t, testDir)

	watcher, err := newMinecraftConfigWatcher(testDir)
	if err != nil {
		t.Fatal(err)
	}
	defer watcher.Stop()

	writeConfigFile(t, testDir, config.MinecraftDenyIPFile, "1239d--s[]")
	select {
	case <-watcher.ConfigEvents:
		t.Errorf("expected an error")
	case err = <-watcher.Errors:
	}
	if err == nil {
		t.Errorf("expected an error")
	}
}

func TestMinecraftConfigWatcherMissingConfig(t *testing.T) {
	testDir := t.TempDir()
	_, err := newMinecraftConfigWatcher(testDir)
	if err == nil {
		t.Errorf("expected an error")
	}
}
