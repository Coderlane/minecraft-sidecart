package main

import (
	"reflect"
	"testing"

	"github.com/99designs/keyring"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

const (
	testProjectID = "minecraft-sidecart-test"
	testUserID    = "test-user"
)

func testNewUserCache(t *testing.T) (*LocalUserCache, keyring.Keyring) {
	t.Helper()
	ring, err := keyring.Open(keyring.Config{
		ServiceName:             testProjectID,
		LibSecretCollectionName: "login",
	})
	if err != nil {
		t.Fatal(err)
	}
	luc, err := NewLocalUserCache(testProjectID)
	if err != nil {
		t.Fatal(err)
	}
	return luc, ring
}

func TestUserCacheGetSet(t *testing.T) {
	luc, ring := testNewUserCache(t)
	defer func() {
		ring.Remove(testUserID)
	}()

	input := &firebase.User{
		RefreshToken: "test",
	}
	user, err := luc.Get(testUserID)
	if user != nil {
		t.Errorf("Expected to fail to get unknown user: %v", user)
	}
	if err != nil {
		t.Errorf("Expected nil error getting unknown user: %v", err)
	}

	err = luc.Set(testUserID, input)
	if err != nil {
		t.Fatal(err)
	}

	output, err := luc.Get(testUserID)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(input, output) {
		t.Errorf("Expected: %v Got: %v\n", input, output)
	}
}

func TestUserCacheGetCorruptRecovers(t *testing.T) {
	luc, ring := testNewUserCache(t)
	defer func() {
		ring.Remove(testUserID)
	}()

	err := ring.Set(keyring.Item{
		Key:  testUserID,
		Data: []byte("]invalid"),
	})

	_, err = luc.Get(testUserID)
	if err == nil {
		t.Errorf("Expected to fail to get user")
	}

	user, err := luc.Get(testUserID)
	if user != nil {
		t.Errorf("Expected to fail to get unknown user: %v", user)
	}
	if err != nil {
		t.Errorf("Expected nil error getting unknown user: %v", err)
	}
}
