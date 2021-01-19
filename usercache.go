package main

import (
	"encoding/json"

	"github.com/99designs/keyring"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

// LocalUserCache implements the UserCache interface by caching users
// in the local keychain.
type LocalUserCache struct {
	ring keyring.Keyring
}

// NewLocalUserCache creates a new LocalUserCache.
func NewLocalUserCache(projectID string) (*LocalUserCache, error) {
	ring, err := keyring.Open(keyring.Config{
		ServiceName:             projectID,
		LibSecretCollectionName: "login",
	})
	if err != nil {
		return nil, err
	}
	return &LocalUserCache{
		ring: ring,
	}, nil
}

// Get will fetch a user from the cache.
// If no user is found, it returns (nil, nil)
func (luc *LocalUserCache) Get(userID string) (*firebase.User, error) {
	item, err := luc.ring.Get(userID)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, nil
		}
		return nil, err
	}
	var user firebase.User
	if err = json.Unmarshal(item.Data, &user); err != nil {
		// Corrupt entry, delete it
		luc.ring.Remove(userID)
		return nil, err
	}
	return &user, nil
}

// Set adds a user to the user cache.
func (luc *LocalUserCache) Set(userID string, user *firebase.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		panic(err)
	}
	return luc.ring.Set(keyring.Item{
		Key:  userID,
		Data: data,
	})
}
