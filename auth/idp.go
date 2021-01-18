package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/oauth2"
)

const (
	// IdpUserID is an extra field attached to all to tokens. Use:
	// `token.Extra(IdpUserID)` to fetch a token's user ID.
	IdpUserID string = "UserID"

	defaultAuthHost  = "identitytoolkit.googleapis.com"
	defaultAuthPath  = "v1/accounts:signInWithIdp"
	defaultTokenHost = "securetoken.googleapis.com"
	defaultTokenPath = "v1/token"
)

// IdpConfig represents the configuration for the google identity provider
type IdpConfig struct {
	// authURL is the URL that OAuth2 Access Tokens are exchanged for IDP
	// ID Tokens
	authURL string
	// tokenURL is the URL that allows refreshing IDP ID Tokens
	tokenURL string
	// APIKey is the API key used for IDP APIs
	APIKey string `json:"api_key"`
}

type idpExchangeRequest struct {
	PostBody            string `json:"postBody"`
	ProviderID          string `json:"providerId"`
	RequestURI          string `json:"requestUri"`
	ReturnIdpCredential bool   `json:"returnIdpCredential"`
	ReturnSecureToken   bool   `json:"returnSecureToken"`
}

type idpExchangeResponse struct {
	FederatedID      string `json:"federatedId"`
	ProviderID       string `json:"providerId"`
	LocalID          string `json:"localId"`
	EmailVerified    bool   `json:"emailVerified"`
	Email            string `json:"email"`
	OAuthAccessToken string `json:"oauthAccessToken"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	FullName         string `json:"fullName"`
	DisplayName      string `json:"displayName"`
	IDToken          string `json:"idToken"`
	PhotoURL         string `json:"photoUrl"`
	RefreshToken     string `json:"refreshToken"`
	ExpiresIn        string `json:"expiresIn"`
	RawUserInfo      string `json:"rawUserInfo"`
}

type idpRefreshResponse struct {
	ExpiresIn    string `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	UserID       string `json:"user_id"`
	ProjectID    string `json:"project_id"`
}

type idpExtra struct {
	UserID string `json:"user_id"`
}

func buildAPIUrl(emulatorHost, apiHost, apiPath, apiKey string) string {
	outputQuery := url.Values{}
	outputQuery.Set("key", apiKey)
	outputURL := url.URL{
		RawQuery: outputQuery.Encode(),
	}
	if emulatorHost == "" {
		outputURL.Scheme = "https"
		outputURL.Host = apiHost
		outputURL.Path = apiPath
	} else {
		outputURL.Scheme = "http"
		outputURL.Host = emulatorHost
		outputURL.Path = fmt.Sprintf("%s/%s", apiHost, apiPath)
	}
	return outputURL.String()
}

func fixupExpiry(expiryString string) (*time.Time, error) {
	expirySeconds, err := strconv.ParseFloat(expiryString, 64)
	if err != nil {
		return nil, err
	}
	expiry := time.Now().Add(time.Second * time.Duration(expirySeconds*.75))
	return &expiry, nil
}

// ConfigFromJSON parses an IdpConfig from its JSON representation, filling in
// any necessary default values.
func ConfigFromJSON(data []byte) (*IdpConfig, error) {
	return configFromJSONWithEmulator(data, "")
}

func configFromJSONWithEmulator(data []byte, emulatorHost string) (*IdpConfig, error) {
	var cfg IdpConfig
	err := json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	cfg.authURL = buildAPIUrl(emulatorHost, defaultAuthHost, defaultAuthPath, cfg.APIKey)
	cfg.tokenURL = buildAPIUrl(emulatorHost, defaultTokenHost, defaultTokenPath, cfg.APIKey)
	return &cfg, nil
}

// Exchange exchanges an OAuth2 Access Token for an IDP ID Token
func (cfg IdpConfig) Exchange(
	ctx context.Context, accessToken *oauth2.Token) (*oauth2.Token, error) {
	postValues := url.Values{}
	postValues.Add("providerId", "google.com")
	idToken := accessToken.Extra("id_token")
	if idToken != nil {
		postValues.Add("id_token", idToken.(string))
	} else {
		postValues.Add("access_token", accessToken.AccessToken)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.Encode(idpExchangeRequest{
		PostBody:            postValues.Encode(),
		RequestURI:          "http://localhost",
		ReturnIdpCredential: true,
		ReturnSecureToken:   true,
	})

	request, err := http.NewRequestWithContext(ctx,
		http.MethodPost, cfg.authURL, &buf)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newIdpError(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, newIdpErrorFromResponse(
			fmt.Errorf("failed to exchange token"), resp.StatusCode, string(data))
	}

	var idpResp idpExchangeResponse
	if err := json.Unmarshal(data, &idpResp); err != nil {
		return nil, newIdpError(err)
	}

	expiry, err := fixupExpiry(idpResp.ExpiresIn)
	if err != nil {
		return nil, newIdpError(err)
	}

	token := &oauth2.Token{
		AccessToken:  idpResp.IDToken,
		RefreshToken: idpResp.RefreshToken,
		Expiry:       *expiry,
	}
	extra := map[string]interface{}{
		IdpUserID: idpResp.LocalID,
	}
	return token.WithExtra(extra), nil
}

// Refresh gets a new IDP ID Token with the provided refresh token.
func (cfg IdpConfig) Refresh(
	ctx context.Context, refreshToken *oauth2.Token) (*oauth2.Token, error) {

	postValues := url.Values{}
	postValues.Add("refresh_token", refreshToken.RefreshToken)
	postValues.Add("grant_type", "refresh_token")

	var buf bytes.Buffer
	buf.WriteString(postValues.Encode())

	request, err := http.NewRequestWithContext(ctx,
		http.MethodPost, cfg.tokenURL, &buf)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, newIdpError(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, newIdpErrorFromResponse(
			fmt.Errorf("failed to exchange token"), resp.StatusCode, string(data))
	}

	var idpResp idpRefreshResponse
	if err := json.Unmarshal(data, &idpResp); err != nil {
		return nil, newIdpError(err)
	}

	expiry, err := fixupExpiry(idpResp.ExpiresIn)
	if err != nil {
		return nil, newIdpError(err)
	}

	token := &oauth2.Token{
		AccessToken:  idpResp.IDToken,
		RefreshToken: idpResp.RefreshToken,
		Expiry:       *expiry,
	}
	extra := map[string]interface{}{
		IdpUserID: idpResp.UserID,
	}
	return token.WithExtra(extra), nil
}

// TokenSource creates a new IDP Token Source which provides tokens with the
// IDP ID Token as the Access Token. Wrap this class with a cache to avoid
// extra refresh calls.
func (cfg IdpConfig) TokenSource(
	ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	return newIdpTokenSource(ctx, cfg, token)
}

type idpTokenSource struct {
	ctx   context.Context
	cfg   IdpConfig
	token *oauth2.Token
}

func newIdpTokenSource(ctx context.Context,
	cfg IdpConfig, token *oauth2.Token) *idpTokenSource {
	return &idpTokenSource{
		ctx:   ctx,
		cfg:   cfg,
		token: token,
	}
}

func (ts *idpTokenSource) Token() (*oauth2.Token, error) {
	token, err := ts.cfg.Refresh(ts.ctx, ts.token)
	if err != nil {
		return nil, err
	}
	ts.token = token
	return token, nil
}
