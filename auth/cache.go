package auth

import (
	"encoding/json"
	"io/ioutil"

	"golang.org/x/oauth2"
)

type cachedToken struct {
	oauth2.Token
	UserID string `json:"user_id,omitempty"`
}

// InsecureDiskReuseTokenSource writes tokens out to cachePath so they can be
// reused in subsiquent runs. This is insecure as we are writing user
// credentials to disk. Ideally, we'd write these to the OS's credential store.
func InsecureDiskReuseTokenSource(
	cachePath string, src oauth2.TokenSource) oauth2.TokenSource {
	return &insecureDiskReuseTokenSource{
		cachePath: cachePath,
		src:       src,
	}
}

// LoadDiskToken attempts to load a cached token from disk.
func LoadDiskToken(cachePath string) *oauth2.Token {
	data, err := ioutil.ReadFile(cachePath)
	if err == nil {
		var cacheToken cachedToken
		err = json.Unmarshal(data, &cacheToken)
		if err == nil {
			return cacheToken.Token.WithExtra(map[string]interface{}{
				IdpUserID: cacheToken.UserID,
			})
		}
	}
	return nil
}

type insecureDiskReuseTokenSource struct {
	cachePath string
	token     *oauth2.Token
	src       oauth2.TokenSource
}

func (idr *insecureDiskReuseTokenSource) Token() (*oauth2.Token, error) {
	// If we don't have a token, try to load one from disk.
	if idr.token == nil {
		idr.token = LoadDiskToken(idr.cachePath)
	}
	// If we now have a valid token, bail early and return it.
	if idr.token != nil && idr.token.Valid() {
		return idr.token, nil
	}
	// No valid token, fetch a new one from the source.
	token, err := idr.src.Token()
	if err != nil {
		return nil, err
	}
	idr.token = token
	// Save the token for later use.
	userID, _ := token.Extra(IdpUserID).(string)
	cacheToken := cachedToken{
		Token:  *token,
		UserID: userID,
	}
	data, err := json.Marshal(cacheToken)
	if err == nil {
		ioutil.WriteFile(idr.cachePath, data, 0600)
	}
	return token, nil
}
