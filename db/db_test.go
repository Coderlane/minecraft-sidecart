package db

import (
	"context"
	"os"
	"testing"

	firestore "cloud.google.com/go/firestore"

	"github.com/Coderlane/minecraft-sidecart/server"
)

func testRequiresEmulators(t *testing.T) {
	t.Helper()
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skipf("Test requires firestore emulator.")
	}
	host = os.Getenv("FIREBASE_AUTH_EMULATOR_HOST")
	if host == "" {
		t.Skipf("Test requires auth emulator.")
	}
}

func testNewDatabase(t *testing.T, ctx context.Context) Database {
	testRequiresEmulators(t)
	t.Helper()
	store, err := firestore.NewClient(ctx, "minecraft-sidecart")
	if err != nil {
		t.Fatal(err)
	}

	db, err := NewDatabase(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestDatabaseCreateAndUpdate(t *testing.T) {
	ctx := context.Background()
	db := testNewDatabase(t, ctx)

	id, err := db.CreateServer(ctx, "test",
		server.ServerTypeMinecraft, server.MinecraftServerInfo{})
	if err != nil {
		t.Fatal(err)
	}

	err = db.UpdateServerInfo(ctx, id, server.MinecraftServerInfo{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDatabaseCreateHandlesFailure(t *testing.T) {
	ctx := context.Background()
	db := testNewDatabase(t, ctx)

	ctx, cancel := context.WithCancel(ctx)
	cancel()
	_, err := db.CreateServer(ctx, "test",
		server.ServerTypeMinecraft, server.MinecraftServerInfo{})
	if err == nil {
		t.Error("Expected an error")
	}
	t.Log(err)
}

func TestDatabaseUpdateHandlesFailure(t *testing.T) {
	ctx := context.Background()
	db := testNewDatabase(t, ctx)

	err := db.UpdateServerInfo(ctx, "unknown", server.MinecraftServerInfo{})
	if err == nil {
		t.Error("Expected an error")
	}
	t.Log(err)
}
