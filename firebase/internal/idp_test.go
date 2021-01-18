package internal

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
)

const testIdpConfigData = `{
  "api_key": "test_api_key"
}`

func testIdpConfig(t *testing.T) *IdpConfig {
	t.Helper()
	emulatorHost := os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if emulatorHost == "" {
		t.Skipf("Environment variable FIREBASE_AUTH_EMULATOR_HOST is unset. " +
			"This test requires the firebase auth emulator.")
	}
	cfg := NewIdpConfig("test_api_key", emulatorHost)
	return cfg
}

func testIdToken(t *testing.T) string {
	idTokenClaim := jws.ClaimSet{
		Aud: "test",
		Exp: time.Now().Add(time.Hour).Unix(),
		Iat: time.Now().Unix(),
		Sub: "test",
	}
	idTokenData, err := json.Marshal(idTokenClaim)
	if err != nil {
		t.Fatal(err)
	}
	return string(idTokenData)
}

func testAccessToken(t *testing.T) *oauth2.Token {
	t.Helper()
	accessToken := oauth2.Token{}
	return accessToken.WithExtra(map[string]interface{}{
		"id_token": testIdToken(t),
	})
}

func TestEmulatorExchangeToken(t *testing.T) {
	cfg := testIdpConfig(t)
	ctx := context.Background()
	_, refreshToken, err := cfg.Exchange(ctx, testAccessToken(t))
	if err != nil {
		t.Fatal(err)
	}
	if refreshToken == nil {
		t.Fatal("Expected non nil token")
	}
	now := time.Now()
	if now.After(refreshToken.Expiry) {
		t.Errorf("Expected time left on token. Now: %v Token: %v\n",
			now, refreshToken.Expiry)
	}
}

func TestEmulatorExchangeInvalidToken(t *testing.T) {
	cfg := testIdpConfig(t)
	ctx := context.Background()
	_, _, err := cfg.Exchange(ctx, &oauth2.Token{})
	if err == nil {
		t.Fatal("Expected an error")
	}
	if err.(*url.Error).Temporary() {
		t.Error("Expected a permanent error")
	}
	if err.(*url.Error).Timeout() {
		t.Error("Expected a non-timeout error")
	}
}

func TestEmulatorRefreshToken(t *testing.T) {
	cfg := testIdpConfig(t)
	ctx := context.Background()
	_, refreshToken, err := cfg.Exchange(ctx, testAccessToken(t))
	if err != nil {
		t.Fatal(err)
	}
	if refreshToken == nil {
		t.Fatal("Expected non nil token")
	}
	newToken, err := cfg.Refresh(ctx, refreshToken)
	if err != nil {
		t.Fatal(err)
	}
	if newToken == nil {
		t.Fatal("Expected non nil token")
	}
	now := time.Now()
	if now.After(refreshToken.Expiry) {
		t.Errorf("Expected time left on token. Now: %v Token: %v\n",
			now, refreshToken.Expiry)
	}
}
