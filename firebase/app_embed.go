package firebase

// DefaultApp is the application generated from the services file
var DefaultApp = &App{
	APIKey:            "test",
	AuthDomain:        "",
	DatabaseURL:       "",
	ProjectID:         "minecraft-sidecart",
	StorageBucket:     "",
	MessagingSenderID: "",
	AppID:             "",
	MeasurementID:     "",
}

func init() {
	authProviderConfigs[GoogleAuthProvider] = []byte(`{
  "installed": {
    "client_id": "test",
    "project_id": "minecraft-sidecart",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "test",
    "redirect_uris": [
      "urn:ietf:wg:oauth:2.0:oob",
      "http://localhost"
    ]
  }
}
`)
}
