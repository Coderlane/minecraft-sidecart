package firebase

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"

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
	Delete(string)
}

type MemoryUserCache map[string]*User

func (muc MemoryUserCache) Get(userID string) (*User, error) {
	user, ok := muc[userID]
	if !ok {
		return nil, fmt.Errorf("Not found.")
	}
	return user, nil
}

func (muc *MemoryUserCache) Set(userID string, user *User) error {
	(*muc)[userID] = user
	return nil
}

func (muc *MemoryUserCache) Delete(userID string) {
	delete(*muc, userID)
}

type Auth struct {
	app          *App
	emulatorHost string
	idpConfig    *internal.IdpConfig
	userCache    UserCache

	mtx         sync.RWMutex
	currentUser *User
	token       *oauth2.Token
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
	auth.mtx.RLock()
	user := auth.currentUser
	auth.mtx.RUnlock()
	return user
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

	auth.mtx.Lock()
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
	auth.mtx.Unlock()
	return auth.currentUser, token, err
}

func (auth *Auth) signInWithUser(ctx context.Context,
	user *User) (token *oauth2.Token, err error) {
	token, err = auth.idpConfig.Refresh(ctx,
		&oauth2.Token{RefreshToken: user.RefreshToken})
	if err != nil {
		return nil, err
	}
	auth.currentUser = user
	auth.currentUser.RefreshToken = token.RefreshToken
	auth.token = token
	if auth.userCache != nil {
		err = auth.userCache.Set("default", auth.currentUser)
	}
	return token, err
}

func (auth *Auth) SignInWithUser(ctx context.Context,
	user *User) (token *oauth2.Token, err error) {
	auth.mtx.Lock()
	token, err = auth.signInWithUser(ctx, user)
	auth.mtx.Unlock()
	return token, err
}

func (auth *Auth) SignOut() {
	auth.mtx.Lock()
	if auth.currentUser != nil && auth.userCache != nil {
		auth.userCache.Delete(auth.currentUser.UserID)
	}
	auth.currentUser = nil
	auth.token = nil
	auth.mtx.Unlock()
}

func (auth *Auth) Token() (token *oauth2.Token, err error) {
	auth.mtx.RLock()
	if auth.token.Valid() {
		token = auth.token
	} else if auth.currentUser == nil {
		err = fmt.Errorf("no user is currently authenticated")
	}
	auth.mtx.RUnlock()
	if token != nil || err != nil {
		return token, err
	}
	auth.mtx.Lock()
	defer auth.mtx.Unlock()
	if auth.token.Valid() {
		token = auth.token
		return token, nil
	}
	// TODO: Get the context from elsewhere
	ctx := context.Background()
	return auth.signInWithUser(ctx, auth.currentUser)
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

func NewOAuthConfig(provider AuthProvider) *oauth2.Config {
	cfg, err := gauth.ConfigFromJSON(authProviderConfigs[provider], "")
	if err != nil {
		panic(err)
	}
	switch provider {
	case GoogleAuthProvider:
		cfg.Scopes = []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"openid",
		}
	}
	return cfg
}
