package cmd

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/Coderlane/minecraft-sidecart/daemon"
	"github.com/Coderlane/minecraft-sidecart/firebase"
	"github.com/Coderlane/minecraft-sidecart/internal"
)

var restoreConfigPaths = daemon.ConfigPaths

type testContext struct {
	testDir string

	app  *firebase.App
	auth *firebase.Auth
	srv  *internal.FakeAuthServer

	ctx    context.Context
	cancel func()
}

func testAppWithAuth(t *testing.T,
	opts ...firebase.AuthOption) (*firebase.App, *firebase.Auth) {
	t.Helper()
	emulatorHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if emulatorHost == "" {
		t.Skipf("Environment variable FIREBASE_AUTH_EMULATOR_HOST is unset. " +
			"This test requires the firebase auth emulator.")
	}
	app := &firebase.App{
		ProjectID: "test",
		APIKey:    "test_api_key",
	}
	return app, app.NewAuth(append(opts, firebase.WithEmulatorHost(emulatorHost))...)
}

func newTestContext(t *testing.T, opts ...firebase.AuthOption) *testContext {
	ctx, cancel := context.WithCancel(context.Background())
	daemon.DefaultRootDir = t.TempDir()
	app, auth := testAppWithAuth(t, opts...)

	tc := &testContext{
		testDir: t.TempDir(),

		app:  app,
		auth: auth,
		srv:  internal.NewFakeAuthServer(),

		ctx:    ctx,
		cancel: cancel,
	}
	daemon.ConfigPaths = append(daemon.ConfigPaths,
		path.Join(tc.testDir, "daemon.json"))
	return tc
}

func (tc *testContext) Stop() {
	tc.cancel()
	tc.srv.Close()
	daemon.ConfigPaths = restoreConfigPaths
}

const testServerConfig = `
motd=A Minecraft Server
rcon.port=25575
max-players=25
enable-rcon=true
rcon.password=hunter2
server-ip=
server-port=25565`

func (tc *testContext) createTestServer(t *testing.T) string {
	dir := t.TempDir()
	err := ioutil.WriteFile(path.Join(dir, "server.properties"),
		[]byte(testServerConfig), 0600)
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func (tc *testContext) newApp() *cli.App {
	app := cli.NewApp()
	app.Setup()
	app.Metadata = map[string]interface{}{
		"auth": tc.auth,
		"app":  tc.app,
	}
	app.Commands = Commands
	return app
}

func (tc *testContext) StartDaemon(t *testing.T) {
	go func() {
		cliContext := &cli.Context{
			App:     tc.newApp(),
			Context: tc.ctx,
		}
		if err := daemonCommand.Action(cliContext); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Millisecond * 250)
}

func TestDaemonStart(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Stop()

	tc.cancel()
	app := tc.newApp()
	err := app.RunContext(tc.ctx, []string{"test", "daemon"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAuthSignIn(t *testing.T) {
	t.Skipf("github.com/urfave/cli/pull/1225")
	tc := newTestContext(t)
	defer tc.Stop()
	tc.StartDaemon(t)

	app := tc.newApp()
	app.Metadata["oauth"] = tc.srv.Config(t)
	app.Reader = strings.NewReader(internal.TestOAuthCode + "\n")
	err := app.Run([]string{"test", "auth", "signin"})
	if err != nil {
		t.Fatal(err)
	}

	err = app.Run([]string{"test", "auth", "signout"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAuthSignInWithoutDaemonSucceeds(t *testing.T) {
	t.Skipf("github.com/urfave/cli/pull/1225")
	tc := newTestContext(t)
	defer tc.Stop()

	app := tc.newApp()
	app.Metadata["oauth"] = tc.srv.Config(t)
	app.Reader = strings.NewReader(internal.TestOAuthCode + "\n")
	err := app.Run([]string{"test", "auth", "signin"})
	if err != nil {
		t.Fatal(err)
	}

	err = app.Run([]string{"test", "auth", "signout"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAuthSignInWithBadCodeFails(t *testing.T) {
	tc := newTestContext(t)
	defer tc.Stop()
	tc.StartDaemon(t)

	app := tc.newApp()
	app.Metadata["oauth"] = tc.srv.Config(t)
	app.Reader = strings.NewReader("invalid\n")

	err := app.Run([]string{"test", "auth", "signin"})
	if err == nil {
		t.Errorf("Expected to fail")
	}
	t.Log(err)
}

func TestServerAdd(t *testing.T) {
	cache := &firebase.MemoryUserCache{
		"default": &firebase.User{},
	}
	tc := newTestContext(t, firebase.WithUserCache(cache))
	defer tc.Stop()
	tc.StartDaemon(t)

	app := tc.newApp()
	err := app.Run([]string{"test", "server", "add",
		"--name", "test", "--path", tc.createTestServer(t)})
	if err != nil {
		t.Fatal(err)
	}
}

func TestServerAddDuplicateFails(t *testing.T) {
	cache := &firebase.MemoryUserCache{
		"default": &firebase.User{},
	}
	tc := newTestContext(t, firebase.WithUserCache(cache))
	defer tc.Stop()
	tc.StartDaemon(t)

	app := tc.newApp()
	serverPath := tc.createTestServer(t)
	err := app.Run([]string{"test", "server", "add",
		"--name", "test", "--path", serverPath})
	if err != nil {
		t.Fatal(err)
	}

	err = app.Run([]string{"test", "server", "add",
		"--name", "test", "--path", serverPath})
	if err == nil {
		t.Errorf("Expected to fail to add a server with a duplicate path.")
	}
	t.Log(err)
}

func TestServerAddWithInvalidPathFails(t *testing.T) {
	cache := &firebase.MemoryUserCache{
		"default": &firebase.User{},
	}
	tc := newTestContext(t, firebase.WithUserCache(cache))
	defer tc.Stop()
	tc.StartDaemon(t)

	app := tc.newApp()
	err := app.Run([]string{"test", "server", "add",
		"--name", "test", "--path", "invalid"})
	if err == nil {
		t.Errorf("Expected to fail to add a server with an invalid path.")
	}
	t.Log(err)
}

func TestServerAddWithoutDaemonFails(t *testing.T) {
	cache := &firebase.MemoryUserCache{
		"default": &firebase.User{},
	}
	tc := newTestContext(t, firebase.WithUserCache(cache))
	defer tc.Stop()

	app := tc.newApp()
	err := app.Run([]string{"test", "server", "add",
		"--name", "test", "--path", tc.createTestServer(t)})
	if err == nil {
		t.Errorf("Expected to fail to add a server without a damon running.")
	}
	t.Log(err)
}
