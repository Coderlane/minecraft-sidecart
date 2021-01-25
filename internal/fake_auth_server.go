package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

const (
	TestClientID     = "test_client_id"
	TestClientSecret = "test_client_secret"
	TestOAuthCode    = "test_auth_code"

	TestOAuthToken = "test_auth_token"
)

type FakeAuthServer struct {
	server *httptest.Server
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

func NewFakeAuthServer() *FakeAuthServer {
	tas := &FakeAuthServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", tas.handleOAuthToken)
	tas.server = httptest.NewServer(mux)
	return tas
}

func (tas *FakeAuthServer) handleOAuthToken(
	w http.ResponseWriter, r *http.Request) {
	data, _ := ioutil.ReadAll(r.Body)
	values, _ := url.ParseQuery(string(data))
	if !validateValues("client_id", TestClientID, values, w) {
		return
	}
	if !validateValues("client_secret", TestClientSecret, values, w) {
		return
	}
	if !validateValues("code", TestOAuthCode, values, w) {
		return
	}
	tokenValues := url.Values{}
	tokenValues.Add("access_token", TestOAuthToken)
	tokenValues.Add("id_token", TestIdToken(nil))
	tokenValues.Add("token_type", "Bearer")
	w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
	w.Write([]byte(tokenValues.Encode()))
}

func (tas *FakeAuthServer) Config(t *testing.T) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     TestClientID,
		ClientSecret: TestClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthStyle: oauth2.AuthStyleInParams,
			AuthURL:   tas.server.URL + "/oauth/auth",
			TokenURL:  tas.server.URL + "/oauth/token",
		},
	}
}

func (tas *FakeAuthServer) Close() {
	tas.server.Close()
}

func TestIdToken(t *testing.T) string {
	idTokenClaim := jws.ClaimSet{
		Aud: "test",
		Exp: time.Now().Add(time.Hour).Unix(),
		Iat: time.Now().Unix(),
		Sub: "test",
	}
	idTokenData, err := json.Marshal(idTokenClaim)
	if err != nil {
		panic(err)
	}
	return string(idTokenData)
}
