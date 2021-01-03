package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	testClientID     = "test_client_id"
	testClientSecret = "test_client_secret"
	testOAuthCode    = "test_auth_code"

	testOAuthToken = "test_auth_token"

	testIdpIDToken      = "test_idp_token"
	testIdpIDToken1     = "test_idp_token1"
	testIdpIDToken2     = "test_idp_token2"
	testIdpRefreshToken = "test_idp_refresh"
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
	mux.HandleFunc("/idp/auth", tas.handleIdpAuth)
	mux.HandleFunc("/idp/token", tas.handleIdpToken)
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
	tokenValues.Add("token_type", "Bearer")
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Write([]byte(tokenValues.Encode()))
}

func (tas *testAuthServer) handleIdpAuth(
	w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	var idpReq idpExchangeRequest
	_ = json.Unmarshal(data, &idpReq)
	values, _ := url.ParseQuery(idpReq.PostBody)
	if !validateValues("access_token", testOAuthToken, values, w) {
		return
	}
	if !validateValues("providerId", "google.com", values, w) {
		return
	}
	idpResp := idpExchangeResponse{
		IDToken:      fmt.Sprintf("%s%d", testIdpIDToken, tas.idpTokenCounter),
		RefreshToken: testIdpRefreshToken,
		ExpiresIn:    "3600",
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(idpResp)
	tas.idpTokenCounter++
}

func (tas *testAuthServer) handleIdpToken(
	w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	values, _ := url.ParseQuery(string(data))
	if !validateValues("refresh_token", testIdpRefreshToken, values, w) {
		return
	}
	if !validateValues("grant_type", "refresh_token", values, w) {
		return
	}
	idpResp := idpRefreshResponse{
		IDToken:      fmt.Sprintf("%s%d", testIdpIDToken, tas.idpTokenCounter),
		RefreshToken: testIdpRefreshToken,
		ExpiresIn:    "3600",
	}
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.Encode(idpResp)
	tas.idpTokenCounter++
}

func (tas *testAuthServer) Config() (*oauth2.Config, *IdpConfig) {
	return &oauth2.Config{
			ClientID:     testClientID,
			ClientSecret: testClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthStyle: oauth2.AuthStyleInParams,
				AuthURL:   tas.server.URL + "/oauth/auth",
				TokenURL:  tas.server.URL + "/oauth/token",
			},
		}, &IdpConfig{
			AuthURL:  tas.server.URL + "/idp/auth",
			TokenURL: tas.server.URL + "/idp/token",
		}
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

	oauthConfig, idpConfig := tas.Config()
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
	if testIdpIDToken1 != tok.AccessToken {
		t.Errorf("Expected: %s Got: %s\n", testIdpIDToken1, tok.AccessToken)
	}
}

func TestRefreshAuth(t *testing.T) {
	tas := newTestAuthServer()
	defer tas.Close()

	oauthConfig, idpConfig := tas.Config()
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
	if testIdpIDToken1 != tok.AccessToken {
		t.Errorf("Expected: %s Got: %s\n", testIdpIDToken1, tok.AccessToken)
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
	if testIdpIDToken2 != tok.AccessToken {
		t.Errorf("Expected: %s Got: %s\n", testIdpIDToken2, tok.AccessToken)
	}
}
