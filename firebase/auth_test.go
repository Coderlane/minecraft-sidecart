package firebase

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/Coderlane/minecraft-sidecart/internal"
)

func testAuthBuffer() *bytes.Buffer {
	var buf bytes.Buffer
	_, err := buf.WriteString(internal.TestOAuthCode + "\n")
	if err != nil {
		panic(err)
	}
	return &buf
}

func testAppWithAuth(t *testing.T, opts ...AuthOption) (*App, *Auth) {
	t.Helper()
	emulatorHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if emulatorHost == "" {
		t.Skipf("Environment variable FIREBASE_AUTH_EMULATOR_HOST is unset. " +
			"This test requires the firebase auth emulator.")
	}
	app := &App{
		APIKey: "test_api_key",
	}
	return app, app.NewAuth(append(opts, WithEmulatorHost(emulatorHost))...)
}

func TestSignInWithConsoleAuthenticates(t *testing.T) {
	tas := internal.NewFakeAuthServer()
	defer tas.Close()

	ctx := context.Background()
	muc := MemoryUserCache{}
	_, auth := testAppWithAuth(t, WithUserCache(&muc))

	user, tok, err := auth.SignInWithConsoleWithInput(ctx,
		tas.Config(t), testAuthBuffer())
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	if now.After(tok.Expiry) {
		t.Errorf("Expected time left on token. Now: %v Token: %v\n",
			now, tok.Expiry)
	}
	t.Logf("%+v\n", user)
}

func TestSignInWithConsoleWithInvalidTokenFails(t *testing.T) {
	tas := internal.NewFakeAuthServer()
	defer tas.Close()

	ctx := context.Background()
	_, auth := testAppWithAuth(t)

	// Use an invalid auth code
	var buf bytes.Buffer
	buf.WriteString("incorrect\n")
	_, _, err := auth.SignInWithConsoleWithInput(ctx,
		tas.Config(t), &buf)
	if err == nil {
		t.Errorf("Expected to fail to authenticate")
	}
	t.Log(err)
}

func TestSignInWithConsoleNeedsInput(t *testing.T) {
	tas := internal.NewFakeAuthServer()
	defer tas.Close()

	ctx := context.Background()
	_, auth := testAppWithAuth(t)

	// No mock input will cause us to fail with EOF
	_, _, err := auth.SignInWithConsole(ctx, tas.Config(t))
	if err == nil {
		t.Errorf("Expected to fail to get input")
	}
	t.Log(err)
}

func TestSignInWithTokenWithInvalidTokenFails(t *testing.T) {
	tas := internal.NewFakeAuthServer()
	defer tas.Close()

	ctx := context.Background()
	_, auth := testAppWithAuth(t)

	_, _, err := auth.SignInWithToken(ctx, &oauth2.Token{})
	if err == nil {
		t.Errorf("Expected to fail to authenticate")
	}
	t.Log(err)
}

func TestSignInWithUserAuthenticates(t *testing.T) {
	tas := internal.NewFakeAuthServer()
	defer tas.Close()

	ctx := context.Background()
	_, auth := testAppWithAuth(t)

	user, _, err := auth.SignInWithConsoleWithInput(ctx,
		tas.Config(t), testAuthBuffer())
	if err != nil {
		t.Fatal(err)
	}

	_, auth = testAppWithAuth(t)
	tok, err := auth.SignInWithUser(ctx, user)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	if now.After(tok.Expiry) {
		t.Errorf("Expected time left on token. Now: %v Token: %v\n",
			now, tok.Expiry)
	}
}

func TestSignOutSignsOut(t *testing.T) {
	tas := internal.NewFakeAuthServer()
	defer tas.Close()

	ctx := context.Background()
	muc := MemoryUserCache{}
	_, auth := testAppWithAuth(t, WithUserCache(&muc))

	user, _, err := auth.SignInWithConsoleWithInput(ctx,
		tas.Config(t), testAuthBuffer())
	if err != nil {
		t.Fatal(err)
	}

	auth.SignOut()
	if found, _ := muc.Get(user.UserID); found != nil {
		t.Errorf("Expected to not find user in cache.")
	}
	if auth.CurrentUser() != nil {
		t.Errorf("Expected to not have a current user.")
	}
}

func TestTokenCacheRefreshes(t *testing.T) {
	tas := internal.NewFakeAuthServer()
	defer tas.Close()

	ctx := context.Background()
	_, auth := testAppWithAuth(t)

	user, _, err := auth.SignInWithConsoleWithInput(ctx,
		tas.Config(t), testAuthBuffer())
	if err != nil {
		t.Fatal(err)
	}

	muc := MemoryUserCache{"default": user}
	_, auth = testAppWithAuth(t, WithUserCache(&muc))

	if !reflect.DeepEqual(user, auth.CurrentUser()) {
		t.Errorf("Expected to reuse user: %v != %v\n", user, auth.CurrentUser())
	}

	tok, err := auth.Token()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	if now.After(tok.Expiry) {
		t.Errorf("Expected time left on token. Now: %v Token: %v\n",
			now, tok.Expiry)
	}

	newtok, err := auth.Token()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(*tok, *newtok) {
		t.Errorf("Expected to reuse token: %v != %v\n", *tok, *newtok)
	}
}

func TestTokenCacheRefreshWithInvalidTokenFails(t *testing.T) {
	tas := internal.NewFakeAuthServer()
	defer tas.Close()

	ctx := context.Background()
	_, auth := testAppWithAuth(t)

	user, _, err := auth.SignInWithConsoleWithInput(ctx,
		tas.Config(t), testAuthBuffer())
	if err != nil {
		t.Fatal(err)
	}

	user.RefreshToken = "invalid"
	muc := MemoryUserCache{"default": user}
	_, auth = testAppWithAuth(t, WithUserCache(&muc))

	if !reflect.DeepEqual(user, auth.CurrentUser()) {
		t.Errorf("Expected to reuse user: %v != %v\n", user, auth.CurrentUser())
	}

	_, err = auth.Token()
	if err == nil {
		t.Errorf("Expected to fail to fetch a new token")
	}
	t.Log(err)
}

func TestTokenCacheUnauthFails(t *testing.T) {
	_, auth := testAppWithAuth(t)
	_, err := auth.Token()
	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestGoogleAuthProviderParsesValid(t *testing.T) {
	authProviderConfigs[GoogleAuthProvider] = []byte(`{
  "installed": {
    "client_id": "test.apps.googleusercontent.com",
    "project_id": "test",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "test",
    "redirect_uris": [
      "urn:ietf:wg:oauth:2.0:oob",
      "http://localhost"
    ]
  }
}`)

	cfg := NewOAuthConfig(GoogleAuthProvider)
	if cfg.ClientSecret != "test" {
		t.Errorf("Incorrect client secret: %s\n", cfg.ClientSecret)
	}
}

func TestGoogleAuthProviderHandlesInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	authProviderConfigs[GoogleAuthProvider] = []byte(`{
  "installed": {
    "client_id": "test.a`)
	NewOAuthConfig(GoogleAuthProvider)
}
