// +build ignore

package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	out, err := os.Create("auth_config.go")
	if err != nil {
		fmt.Println(err)
		return
	}
	out.Write([]byte("package auth\n\nconst defaultAuthJSON = `"))
	f, err := os.Open("../.env/client_secret.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	io.Copy(out, f)
	out.Write([]byte("`\n"))
}
