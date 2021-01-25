package daemon

import (
	"context"
	"testing"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

func TestRPCDaemonRunAndCancel(t *testing.T) {
	DefaultRootDir = t.TempDir()
	app := &firebase.App{}
	auth := app.NewAuth()
	ctx := context.Background()
	rpcDaemon, err := NewRPCDaemon(ctx, app, auth)
	if err != nil {
		t.Fatal(err)
	}
	defer rpcDaemon.Close()

	ctx, cancel := context.WithCancel(ctx)
	cancel()
	err = rpcDaemon.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestRPCDaemonDuplicateFails(t *testing.T) {
	DefaultRootDir = t.TempDir()
	app := &firebase.App{}
	auth := app.NewAuth()
	ctx := context.Background()
	rpcDaemon, err := NewRPCDaemon(ctx, app, auth)
	if err != nil {
		t.Fatal(err)
	}
	defer rpcDaemon.Close()

	_, err = NewRPCDaemon(ctx, app, auth)
	if err == nil {
		t.Errorf("Expected to fail to create duplicate daemon.")
	}

	ctx, cancel := context.WithCancel(ctx)
	cancel()
	err = rpcDaemon.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
}
