package internal

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
}

type IdpUser struct {
	LocalID       string `json:"localId"`
	EmailVerified bool   `json:"emailVerified"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	PhotoURL      string `json:"photoUrl"`
	RefreshToken  string `json:"refreshToken"`
}

type idpExchangeRequest struct {
	PostBody            string `json:"postBody"`
	ProviderID          string `json:"providerId"`
	RequestURI          string `json:"requestUri"`
	ReturnIdpCredential bool   `json:"returnIdpCredential"`
	ReturnSecureToken   bool   `json:"returnSecureToken"`
}

type idpExchangeResponse struct {
	IdpUser
	FederatedID      string `json:"federatedId"`
	ProviderID       string `json:"providerId"`
	OAuthAccessToken string `json:"oauthAccessToken"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	FullName         string `json:"fullName"`
	IDToken          string `json:"idToken"`
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

func NewIdpConfig(apiKey, emulatorHost string) *IdpConfig {
	return &IdpConfig{
		authURL:  buildAPIUrl(emulatorHost, defaultAuthHost, defaultAuthPath, apiKey),
		tokenURL: buildAPIUrl(emulatorHost, defaultTokenHost, defaultTokenPath, apiKey),
	}
}

// Exchange exchanges an OAuth2 Access Token for an IDP ID Token
func (cfg IdpConfig) Exchange(
	ctx context.Context, accessToken *oauth2.Token) (*IdpUser, *oauth2.Token, error) {
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
	if err != nil {
		return nil, nil, err
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, newIdpError(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, newIdpErrorFromResponse(
			fmt.Errorf("failed to exchange token"), resp.StatusCode, string(data))
	}

	var idpResp idpExchangeResponse
	if err := json.Unmarshal(data, &idpResp); err != nil {
		return nil, nil, newIdpError(err)
	}

	expiry, err := fixupExpiry(idpResp.ExpiresIn)
	if err != nil {
		return nil, nil, newIdpError(err)
	}

	token := &oauth2.Token{
		AccessToken:  idpResp.IDToken,
		RefreshToken: idpResp.RefreshToken,
		Expiry:       *expiry,
	}
	return &idpResp.IdpUser, token, nil
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
	if err != nil {
		return nil, err
	}
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
			fmt.Errorf("failed to refresh token"), resp.StatusCode, string(data))
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
	return token, nil
}
