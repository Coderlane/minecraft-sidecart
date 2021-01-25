package main

import (
	"encoding/json"

	"github.com/zalando/go-keyring"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

// LocalUserCache implements the UserCache interface by caching users
// in the local keychain.
type LocalUserCache struct {
	projectID string
}

// NewLocalUserCache creates a new LocalUserCache.
func NewLocalUserCache(projectID string) *LocalUserCache {
	return &LocalUserCache{
		projectID: projectID,
	}
}

// Get will fetch a user from the cache.
// If no user is found, it returns (nil, nil)
func (luc *LocalUserCache) Get(userID string) (*firebase.User, error) {
	data, err := keyring.Get(luc.projectID, userID)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	var user firebase.User
	if err = json.Unmarshal([]byte(data), &user); err != nil {
		// Corrupt entry, delete it
		keyring.Delete(luc.projectID, userID)
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
	return keyring.Set(luc.projectID, userID, string(data))
}

// Delete removes a user from the user cache.
func (luc *LocalUserCache) Delete(userID string) {
	keyring.Delete(luc.projectID, userID)
}
