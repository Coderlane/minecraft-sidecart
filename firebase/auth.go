package firebase

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

	"github.com/Coderlane/minecraft-sidecart/firebase/internal"
)

type AuthProvider int

const (
	GoogleAuthProvider AuthProvider = iota
)

var authProviderConfigs = map[AuthProvider][]byte{}

type UserCache interface {
	Get(string) (*User, error)
	Set(string, *User) error
}

type Auth struct {
	app         *App
	currentUser *User

	token     *oauth2.Token
	idpConfig *internal.IdpConfig

	userCache    UserCache
	emulatorHost string
}

type User struct {
	UserID        string `json:"userId"`
	EmailVerified bool   `json:"emailVerified"`
	Email         string `json:"email"`
	DisplayName   string `json:"displayName"`
	PhotoURL      string `json:"photoUrl"`
	RefreshToken  string `json:"refreshToken"`
}

func (auth *Auth) CurrentUser() *User {
	return auth.currentUser
}

func (auth *Auth) SignInWithConsole(
	ctx context.Context, cfg *oauth2.Config) (*User, *oauth2.Token, error) {
	return auth.SignInWithConsoleWithInput(ctx, cfg, os.Stdin)
}

func (auth *Auth) SignInWithConsoleWithInput(ctx context.Context,
	cfg *oauth2.Config, tokenReader io.Reader) (*User, *oauth2.Token, error) {
	// Force OOB auth for console auth
	tmpCfg := *cfg
	tmpCfg.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"
	state, err := createRandomState()
	if err != nil {
		return nil, nil, err
	}
	url := tmpCfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Printf("Visit this url to authenticate: %v\n", url)
	fmt.Printf("Input code: ")
	reader := bufio.NewReader(tokenReader)
	var code string
	_, err = fmt.Fscanf(reader, "%s", &code)
	if err != nil {
		return nil, nil, err
	}
	token, err := tmpCfg.Exchange(ctx, code)
	if err != nil {
		return nil, nil, err
	}
	return auth.SignInWithToken(ctx, token)
}

func (auth *Auth) SignInWithToken(
	ctx context.Context, token *oauth2.Token) (*User, *oauth2.Token, error) {
	user, token, err := auth.idpConfig.Exchange(ctx, token)
	if err != nil {
		return nil, nil, err
	}
	auth.currentUser = &User{
		UserID:        user.LocalID,
		EmailVerified: user.EmailVerified,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		PhotoURL:      user.PhotoURL,
		RefreshToken:  token.RefreshToken,
	}
	if auth.userCache != nil {
		err = auth.userCache.Set("default", auth.currentUser)
	}
	return auth.currentUser, token, err
}

func (auth *Auth) Token() (*oauth2.Token, error) {
	if auth.token.Valid() {
		return auth.token, nil
	}
	if auth.currentUser == nil {
		return nil, fmt.Errorf("no user is currently authenticated")
	}
	// TODO: Get the context from elsewhere
	ctx := context.Background()
	token, err := auth.idpConfig.Refresh(ctx,
		&oauth2.Token{RefreshToken: auth.currentUser.RefreshToken})
	if err != nil {
		return nil, err
	}
	auth.token = token
	if auth.userCache != nil {
		// Refresh the refresh token
		auth.currentUser.RefreshToken = token.RefreshToken
		err = auth.userCache.Set("default", auth.currentUser)
	}
	return auth.token, nil
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

func NewOAuthConfig(provider AuthProvider) (*oauth2.Config, error) {
	cfg, err := gauth.ConfigFromJSON(authProviderConfigs[provider], "")
	if err != nil {
		return nil, err
	}
	switch provider {
	case GoogleAuthProvider:
		cfg.Scopes = []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"openid",
		}
	}
	return cfg, nil
}
