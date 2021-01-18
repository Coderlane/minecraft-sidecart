package firebase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

const (
	testClientID     = "test_client_id"
	testClientSecret = "test_client_secret"
	testOAuthCode    = "test_auth_code"

	testOAuthToken = "test_auth_token"
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

type testAuthServer struct {
	server     *httptest.Server
	authBuffer bytes.Buffer
}

func validateValues(key, value string,
	values url.Values, w http.ResponseWriter) bool {
	if len(values[key]) == 0 {
		errorStr := fmt.Sprintf("Invalid %s: Got: %v Expected: %s",
			key, values[key], value)
		http.Error(w, errorStr, http.StatusBadRequest)
		return false
	}
	if values[key][0] == value {
		return true
	}
	errorStr := fmt.Sprintf("Invalid %s: Got: %v Expected: %s",
		key, values[key], value)
	http.Error(w, errorStr, http.StatusBadRequest)
	return false
}

func newTestAuthServer() *testAuthServer {
	tas := &testAuthServer{}
	tas.authBuffer.WriteString(testOAuthCode + "\n")
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", tas.handleOAuthToken)
	tas.server = httptest.NewServer(mux)
	return tas
}

func (tas *testAuthServer) handleOAuthToken(
	w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	values, _ := url.ParseQuery(string(data))
	if !validateValues("client_id", testClientID, values, w) {
		return
	}
	if !validateValues("client_secret", testClientSecret, values, w) {
		return
	}
	if !validateValues("code", testOAuthCode, values, w) {
		return
	}
	tokenValues := url.Values{}
	tokenValues.Add("access_token", testOAuthToken)
	tokenValues.Add("id_token", testIdToken(nil))
	tokenValues.Add("token_type", "Bearer")
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Write([]byte(tokenValues.Encode()))
}

func (tas *testAuthServer) Config(t *testing.T) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   tas.server.URL + "/oauth/auth",
			TokenURL:  tas.server.URL + "/oauth/token",
		},
	}
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

func (tas *testAuthServer) Close() {
	tas.server.Close()
}

type testUserCache map[string]*User

func (tuc testUserCache) Get(userID string) (*User, error) {
	user, ok := tuc[userID]
	if !ok {
		return nil, fmt.Errorf("Not found.")
	}
	return user, nil
}

func (tuc *testUserCache) Set(userID string, user *User) error {
	(*tuc)[userID] = user
	return nil
}

func TestSignInWithConsoleAuthenticates(t *testing.T) {
	tas := newTestAuthServer()
	defer tas.Close()

	ctx := context.Background()
	_, auth := testAppWithAuth(t)

	user, tok, err := auth.SignInWithConsoleWithInput(ctx,
		tas.Config(t), &tas.authBuffer)
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

func TestTokenCacheRefreshes(t *testing.T) {
	tas := newTestAuthServer()
	defer tas.Close()

	ctx := context.Background()
	_, auth := testAppWithAuth(t)

	user, _, err := auth.SignInWithConsoleWithInput(ctx,
		tas.Config(t), &tas.authBuffer)
	if err != nil {
		t.Fatal(err)
	}

	tuc := testUserCache{"default": user}
	_, auth = testAppWithAuth(t, WithUserCache(&tuc))

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

	cfg, err := NewOAuthConfig(GoogleAuthProvider)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ClientSecret != "test" {
		t.Errorf("Incorrect client secret: %s\n", cfg.ClientSecret)
	}
}

func TestGoogleAuthProviderHandlesInvalid(t *testing.T) {
	authProviderConfigs[GoogleAuthProvider] = []byte(`{
  "installed": {
    "client_id": "test.a`)
	_, err := NewOAuthConfig(GoogleAuthProvider)
	if err == nil {
		t.Fatal("expected an error")
	}
}
