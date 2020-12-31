// Package auth provides utilities for authenticating the client.
//go:generate go run embed.go

package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/oauth2"
	gauth "golang.org/x/oauth2/google"
)

type TokenSourceProvider interface {
	TokenSource() (oauth2.TokenSource, error)
}

type defaultTokenSourceProvider struct {
	oauth2Config     *oauth2.Config
	authTokenReader  io.Reader
	refreshTokenPath string
}

func NewDefaultTokenSourceProvider() TokenSourceProvider {
	conf, err := gauth.ConfigFromJSON([]byte(defaultAuthJSON), "")
	if err != nil {
		panic(err)
	}

	return NewDefaultTokenSourceProviderWithParams(
		conf, os.Stdin, "./refresh_token.json")
}

func NewDefaultTokenSourceProviderWithParams(
	conf *oauth2.Config, reader io.Reader, refreshTokenPath string) TokenSourceProvider {
	// The only scopes we use are the scopes required for OAuth2:
	// https://developers.google.com/identity/protocols/oauth2/scopes#oauth2
	conf.Scopes = []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"openid",
	}
	// We use OOB authentication as the server may not have a browser installed.
	conf.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
	return &defaultTokenSourceProvider{
		oauth2Config:     conf,
		authTokenReader:  reader,
		refreshTokenPath: refreshTokenPath,
	}
}

func (tsp *defaultTokenSourceProvider) TokenSource() (oauth2.TokenSource, error) {
	ctx := context.Background()
	token, err := tsp.getRefreshToken()
	if token == nil || err != nil {
		if err != nil {
			fmt.Println("Failed to read refresh token:", err)
		}
		token, err = tsp.getNewToken(ctx)
	}
	if err != nil {
		return nil, err
	}
	ts := tsp.oauth2Config.TokenSource(ctx, token)
	return ts, nil
}

func (tsp *defaultTokenSourceProvider) getRefreshToken() (*oauth2.Token, error) {
	tokenData, err := ioutil.ReadFile(tsp.refreshTokenPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var token oauth2.Token
	err = json.Unmarshal(tokenData, &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (tsp *defaultTokenSourceProvider) getNewToken(
	ctx context.Context) (*oauth2.Token, error) {
	state, err := createRandomState()
	if err != nil {
		return nil, err
	}
	url := tsp.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Visit this url to authenticate: %v\n", url)
	fmt.Printf("Input code: ")
	reader := bufio.NewReader(tsp.authTokenReader)
	var code string
	_, err = fmt.Fscanf(reader, "%s", &code)
	if err != nil {
		return nil, err
	}
	token, err := tsp.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	tokenData, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(tsp.refreshTokenPath, tokenData, 0600)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func createRandomState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(b)
	return encoded, nil
}
