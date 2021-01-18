// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

type EmbedData struct {
	firebase.App
	GoogleAuthData string
}

var embedDataTemplate = template.Must(template.New("embed").Parse(
	`package firebase

// DefaultApp is the application generated from the services file
var DefaultApp = &App{
	APIKey:            "{{.APIKey}}",
	AuthDomain:        "{{.AuthDomain}}",
	DatabaseURL:       "{{.DatabaseURL}}",
	ProjectID:         "{{.ProjectID}}",
	StorageBucket:     "{{.StorageBucket}}",
	MessagingSenderID: "{{.MessagingSenderID}}",
	AppID:             "{{.AppID}}",
	MeasurementID:     "{{.MeasurementID}}",
}

func init() {
	authProviderConfigs[GoogleAuthProvider] = []byte(` + "`{{.GoogleAuthData}}`" + `)
}
`))

func main() {
	cfgData, err := ioutil.ReadFile("../.env/firebase_services.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	embedData := EmbedData{}
	if err := json.Unmarshal(cfgData, &embedData.App); err != nil {
		fmt.Println(err)
		return
	}

	authData, _ := ioutil.ReadFile("../.env/google_client_secret.json")
	embedData.GoogleAuthData = string(authData)

	embedFile, err := os.Create("app_embed.go")
	if err != nil {
		fmt.Println(err)
		return
	}

	if err := embedDataTemplate.Execute(embedFile, embedData); err != nil {
		fmt.Println(err)
		return
	}
}
