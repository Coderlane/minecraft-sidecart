// Package auth provides utilities for authenticating the client.
//go:generate go run embed.go
package auth

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"

	"golang.org/x/oauth2"
	gauth "golang.org/x/oauth2/google"
)

// TokenSourceProvider is implemented by anything which can provide TokenSources
type TokenSourceProvider interface {
	TokenSource(context.Context, *oauth2.Token) (oauth2.TokenSource, error)
}

type firebaseTokenSourceProvider struct {
	idpConfig       *IdpConfig
	oauth2Config    *oauth2.Config
	authTokenReader io.Reader
}

// NewFirebaseTokenSourceProvider creates a new TokenSourceProvider for
// Firebase based on the configs build in to the binary.
func NewFirebaseTokenSourceProvider() TokenSourceProvider {
	oauth2Config, err := gauth.ConfigFromJSON([]byte(defaultOAuthJSON), "")
	if err != nil {
		panic(err)
	}
	idpConfig, err := ConfigFromJSON([]byte(defaultIDPJSON))
	if err != nil {
		panic(err)
	}
	return NewFirebaseTokenSourceProviderWithParams(
		oauth2Config, idpConfig, os.Stdin)
}

// NewFirebaseTokenSourceProviderWithParams creates a new TokenSourceProvider
// for Firebase based on the input parameters.
func NewFirebaseTokenSourceProviderWithParams(
	oauth2Config *oauth2.Config, idpConfig *IdpConfig,
	reader io.Reader) TokenSourceProvider {
	// The only scopes we use are the scopes required for OAuth2:
	// https://developers.google.com/identity/protocols/oauth2/scopes#oauth2
	oauth2Config.Scopes = []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
		"openid",
	}
	// We use OOB authentication as the server may not have a browser installed.
	oauth2Config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
	return &firebaseTokenSourceProvider{
		oauth2Config:    oauth2Config,
		idpConfig:       idpConfig,
		authTokenReader: reader,
	}
}

func (tsp *firebaseTokenSourceProvider) TokenSource(
	ctx context.Context, token *oauth2.Token) (oauth2.TokenSource, error) {
	var err error
	if token == nil {
		token, err = tsp.getNewToken(ctx)
		if err != nil {
			return nil, err
		}
	}

	ts := tsp.idpConfig.TokenSource(ctx, token)
	return ts, nil
}

func (tsp *firebaseTokenSourceProvider) getNewToken(
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
	token, err = tsp.idpConfig.Exchange(ctx, token)
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
