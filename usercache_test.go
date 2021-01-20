package main

import (
	"reflect"
	"testing"

	"github.com/zalando/go-keyring"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

const (
	testProjectID = "minecraft-sidecart-test"
	testUserID    = "test-user"
)

func TestUserCacheGetSet(t *testing.T) {
	keyring.MockInit()
	defer func() {
		keyring.Delete(testProjectID, testUserID)
	}()

	luc := NewLocalUserCache(testProjectID)
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
	keyring.MockInit()
	defer func() {
		keyring.Delete(testProjectID, testUserID)
	}()

	luc := NewLocalUserCache(testProjectID)
	err := keyring.Set(testProjectID, testUserID, "invalid")
	if err != nil {
		t.Fatal(err)
	}

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
