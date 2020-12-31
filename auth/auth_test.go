package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
)

const (
	testClientID     = "test_client_id"
	testClientSecret = "test_client_secret"
	testAuthCode     = "test_auth_code"
	testAuthToken    = "test_auth_token"
	testAuthToken0   = "test_auth_token0"
	testAuthToken1   = "test_auth_token1"
)

type testAuthServer struct {
	server       *httptest.Server
	authBuffer   bytes.Buffer
	tokenCounter int
}

func validateHeader(key, value string,
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

func NewTestAuthServer() *testAuthServer {
	tas := &testAuthServer{}
	tas.authBuffer.WriteString(testAuthCode + "\n")
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		values, _ := url.ParseQuery(string(buf))
		if !validateHeader("client_id", testClientID, values, w) {
			return
		}
		if !validateHeader("client_secret", testClientSecret, values, w) {
			return
		}
		if !validateHeader("code", testAuthCode, values, w) {
			return
		}
		tokenValues := url.Values{}
		tokenValues.Add("access_token", fmt.Sprintf("%s%d", testAuthToken, tas.tokenCounter))
		tokenValues.Add("token_type", "Bearer")
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write([]byte(tokenValues.Encode()))
		tas.tokenCounter += 1
	})
	tas.server = httptest.NewServer(mux)
	return tas
}

func (tas *testAuthServer) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   tas.server.URL + "/auth",
			TokenURL:  tas.server.URL + "/token",
		},
	}
}

func (tas *testAuthServer) Close() {
	tas.server.Close()
}

func TestDefaultAuth(t *testing.T) {
	tsp := NewDefaultTokenSourceProvider()
	ts, _ := tsp.TokenSource()
	if ts != nil {
		t.Errorf("Expected to fail to create a token source\n")
	}
}

func TestAuthCode(t *testing.T) {
	tas := NewTestAuthServer()
	defer tas.Close()
	refreshTokenPath := t.TempDir() + "/refresh_token.json"
	tsp := NewDefaultTokenSourceProviderWithParams(
		tas.Config(), &tas.authBuffer, refreshTokenPath)
	ts, err := tsp.TokenSource()
	if err != nil {
		t.Fatal(err)
	}
	tok, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if testAuthToken0 != tok.AccessToken {
		t.Errorf("Expected: %s Got: %s\n", testAuthToken0, tok.AccessToken)
	}
}

func TestRefreshAuth(t *testing.T) {
	tas := NewTestAuthServer()
	defer tas.Close()
	refreshTokenPath := t.TempDir() + "/refresh_token.json"
	tsp := NewDefaultTokenSourceProviderWithParams(
		tas.Config(), &tas.authBuffer, refreshTokenPath)
	ts, err := tsp.TokenSource()
	if err != nil {
		t.Fatal(err)
	}
	tok, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if testAuthToken0 != tok.AccessToken {
		t.Errorf("Expected: %s Got: %s\n", testAuthToken0, tok.AccessToken)
	}

	tsp = NewDefaultTokenSourceProviderWithParams(
		tas.Config(), &tas.authBuffer, refreshTokenPath)
	ts, err = tsp.TokenSource()
	if err != nil {
		t.Fatal(err)
	}
	tok, err = ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if testAuthToken0 != tok.AccessToken {
		t.Errorf("Expected: %s Got: %s\n", testAuthToken0, tok.AccessToken)
	}
}
