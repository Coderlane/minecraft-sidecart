package daemon

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

func testIdToken(t *testing.T) string {
	idTokenClaim := jws.ClaimSet{
		Aud: "test",
		Exp: time.Now().Add(time.Hour).Unix(),
		Iat: time.Now().Unix(),
		Sub: "test",
	}
	idTokenData, err := json.Marshal(idTokenClaim)
	if err != nil {
		t.Fatal(err)
	}
	return string(idTokenData)
}

func testAppWithAuth(t *testing.T) (*firebase.App, *firebase.Auth) {
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
	return app, app.NewAuth(firebase.WithEmulatorHost(emulatorHost))
}

func testNewUnauthDaemon(t *testing.T, ctx context.Context) (*Daemon, *firebase.User) {
	app, auth := testAppWithAuth(t)
	testToken := oauth2.Token{}
	user, _, err := auth.SignInWithToken(ctx,
		testToken.WithExtra(map[string]interface{}{"id_token": testIdToken(t)}))
	if err != nil {
		t.Fatal(err)
	}
	dae, err := NewDaemon(ctx, app, auth)
	if err != nil {
		t.Fatal(err)
	}
	return dae, user
}

func testNewDaemon(t *testing.T, ctx context.Context) *Daemon {
	dae, user := testNewUnauthDaemon(t, ctx)
	var token oauth2.Token
	err := dae.SignIn(user, &token)
	if err != nil {
		t.Fatal(err)
	}
	return dae
}

func TestDaemonSignIn(t *testing.T) {
	testDir := t.TempDir()
	restore := testAddConfigPath(testDir)
	defer restore()

	ctx := context.Background()
	dae, user := testNewUnauthDaemon(t, ctx)

	var token oauth2.Token
	err := dae.SignIn(user, &token)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDaemonAddServerMonitorsServer(t *testing.T) {
	testDir := t.TempDir()
	restore := testAddConfigPath(testDir)
	defer restore()

	ctx := context.Background()
	dae := testNewDaemon(t, ctx)

	testCreateTestServer(t, testDir)
	spec := ServerSpec{
		Path: testDir,
	}
	var id string
	err := dae.AddServer(spec, &id)
	if err != nil {
		t.Fatal(err)
	}
}
