package auth

import (
	"fmt"
	"io/ioutil"
	"path"
	"reflect"
	"testing"

	"golang.org/x/oauth2"
)

type fakeTokenSource struct {
	tokens []*oauth2.Token
}

func newFakeTokenSource(tokens []*oauth2.Token) oauth2.TokenSource {
	return &fakeTokenSource{
		tokens: tokens,
	}
}

func (ts *fakeTokenSource) Token() (*oauth2.Token, error) {
	if len(ts.tokens) == 0 {
		return nil, fmt.Errorf("Synthetic error")
	}
	token := ts.tokens[0]
	ts.tokens = ts.tokens[1:]
	return token, nil
}

func newTestInsecureDiskReuseTokenSource(
	cachePath string, tokens []*oauth2.Token) oauth2.TokenSource {
	ts := newFakeTokenSource(tokens)
	return InsecureDiskReuseTokenSource(cachePath, ts)
}

func newTestToken(version int) *oauth2.Token {
	tok := oauth2.Token{
		AccessToken:  fmt.Sprintf("access_token%d", version),
		TokenType:    "Bearer",
		RefreshToken: fmt.Sprintf("refresh_token%d", version),
	}
	return tok.WithExtra(map[string]interface{}{IdpUserID: "test_id"})
}

func TestInsecureDiskNewToken(t *testing.T) {
	cachePath := path.Join(t.TempDir(), "token.json")
	testTokens := []*oauth2.Token{
		newTestToken(1),
	}
	ts := newTestInsecureDiskReuseTokenSource(cachePath, testTokens)
	token, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testTokens[0], token) {
		t.Errorf("Expected: %v Got: %v\n", testTokens[0], token)
	}
}

func TestInsecureDiskNewTokenFails(t *testing.T) {
	cachePath := path.Join(t.TempDir(), "token.json")
	ts := newTestInsecureDiskReuseTokenSource(cachePath, []*oauth2.Token{})
	_, err := ts.Token()
	if err == nil {
		t.Errorf("Expected to fail.")
	}
}

func TestInsecureDiskReusesMemoryToken(t *testing.T) {
	cachePath := path.Join(t.TempDir(), "token.json")
	testTokens := []*oauth2.Token{
		newTestToken(1),
	}
	ts := newTestInsecureDiskReuseTokenSource(cachePath, testTokens)
	token, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testTokens[0], token) {
		t.Errorf("Expected: %v Got: %v\n", testTokens[0], token)
	}
	token, err = ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testTokens[0], token) {
		t.Errorf("Expected: %v Got: %v\n", testTokens[0], token)
	}
}

func TestInsecureDiskReusesDiskToken(t *testing.T) {
	cachePath := path.Join(t.TempDir(), "token.json")
	testTokens := []*oauth2.Token{
		newTestToken(1),
	}
	ts := newTestInsecureDiskReuseTokenSource(cachePath, testTokens)
	_, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}

	ts = newTestInsecureDiskReuseTokenSource(cachePath, []*oauth2.Token{})
	token, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testTokens[0], token) {
		t.Errorf("Expected: %v Got: %v\n", testTokens[0], token)
	}
}

func TestInsecureDiskReusesDiskTokenFails(t *testing.T) {
	cachePath := path.Join(t.TempDir(), "token.json")
	err := ioutil.WriteFile(cachePath, []byte("invalid"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	testTokens := []*oauth2.Token{
		newTestToken(1),
	}
	ts := newTestInsecureDiskReuseTokenSource(cachePath, testTokens)
	token, err := ts.Token()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testTokens[0], token) {
		t.Errorf("Expected: %v Got: %v\n", testTokens[0], token)
	}
}
