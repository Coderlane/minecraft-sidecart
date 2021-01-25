package daemon

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const testServerConfig = `
motd=A Minecraft Server
rcon.port=25575
max-players=25
enable-rcon=true
rcon.password=hunter2
server-ip=
server-port=25565`

func testAddConfigPath(dir string) func() {
	restoreConfigPaths := ConfigPaths
	restore := func() {
		ConfigPaths = restoreConfigPaths
	}
	ConfigPaths = append(ConfigPaths, path.Join(dir, "daemon.json"))
	return restore
}

func testCreateTestServer(t *testing.T, dir string) {
	os.MkdirAll(dir, 0600)
	err := ioutil.WriteFile(path.Join(dir, "server.properties"),
		[]byte(testServerConfig), 0600)
	if err != nil {
		t.Fatal(err)
	}
}

func TestServerManagerAddStoresServer(t *testing.T) {
	testDir := t.TempDir()
	restore := testAddConfigPath(testDir)
	defer restore()

	mgr, err := newServerManager()
	if err != nil {
		t.Fatal(err)
	}

	testCreateTestServer(t, testDir)
	err = mgr.addServer("test", testDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	mgr, err = newServerManager()
	if err != nil {
		t.Fatal(err)
	}

	if !mgr.hasPath(testDir) {
		t.Errorf("expected to find server after reload")
	}
	if mgr.hasPath(testDir + "test") {
		t.Errorf("expected to not find invalid server")
	}
}

func TestServerManagerHandlesInvalidConfig(t *testing.T) {
	testDir := t.TempDir()
	restore := testAddConfigPath(testDir)
	defer restore()

	err := ioutil.WriteFile(path.Join(testDir, "daemon.json"), []byte("{["), 0600)
	if err != nil {
		t.Fatal(err)
	}

	_, err = newServerManager()
	if err == nil {
		t.Errorf("Expected to fail")
	}
}
