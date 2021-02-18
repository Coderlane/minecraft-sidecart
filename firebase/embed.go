package firebase

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

//go:embed .env/firebase_services.json
var firebaseServicesConfig []byte

//go:embed .env/google_client_secret.json
var googleAuthProviderConfig []byte

func init() {
	authProviderConfigs[GoogleAuthProvider] = googleAuthProviderConfig

	if len(firebaseServicesConfig) > 0 {
		if err := json.Unmarshal(firebaseServicesConfig, &DefaultApp); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}
	}
}
