package auth

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	testClientID     = "test_client_id"
	testClientSecret = "test_client_secret"
	testOAuthCode    = "test_auth_code"

	testOAuthToken = "test_auth_token"
)

type testAuthServer struct {
	server          *httptest.Server
	authBuffer      bytes.Buffer
	idpTokenCounter int
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

func (tas *testAuthServer) Config(t *testing.T) (*oauth2.Config, *IdpConfig) {
	return &oauth2.Config{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   tas.server.URL + "/oauth/auth",
			TokenURL:  tas.server.URL + "/oauth/token",
		},
	}, testIdpConfig(t)
}

func (tas *testAuthServer) Close() {
	tas.server.Close()
}

func TestDefaultAuth(t *testing.T) {
	ctx := context.Background()
	tsp := NewFirebaseTokenSourceProvider()
	ts, _ := tsp.TokenSource(ctx, nil)
	if ts != nil {
		t.Errorf("Expected to fail to create a token source\n")
	}
}

func TestAuthCode(t *testing.T) {
	tas := newTestAuthServer()
	defer tas.Close()

	oauthConfig, idpConfig := tas.Config(t)
	tsp := NewFirebaseTokenSourceProviderWithParams(
		oauthConfig, idpConfig, &tas.authBuffer)
	ctx := context.Background()
	ts, err := tsp.TokenSource(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if now.After(tok.Expiry) {
		t.Errorf("Expected time left on token. Now: %v Token: %v\n",
			now, tok.Expiry)
	}
}

func TestRefreshAuth(t *testing.T) {
	tas := newTestAuthServer()
	defer tas.Close()

	oauthConfig, idpConfig := tas.Config(t)
	tsp := NewFirebaseTokenSourceProviderWithParams(
		oauthConfig, idpConfig, &tas.authBuffer)
	ctx := context.Background()
	ts, err := tsp.TokenSource(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	tok, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}

	tsp = NewFirebaseTokenSourceProviderWithParams(
		oauthConfig, idpConfig, &tas.authBuffer)
	ts, err = tsp.TokenSource(ctx, tok)
	if err != nil {
		t.Fatal(err)
	}
	tok, err = ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	if now.After(tok.Expiry) {
		t.Errorf("Expected time left on token. Now: %v Token: %v\n",
			now, tok.Expiry)
	}
}
