package firebase

type AuthOption interface {
	Apply(auth *Auth)
}

type withUserCache struct {
	uc UserCache
}

func (wuc withUserCache) Apply(auth *Auth) {
	auth.userCache = wuc.uc
}

// Attach a user cache to the authlication client
func WithUserCache(uc UserCache) AuthOption {
	return withUserCache{uc}
}

type withEmulatorHost struct {
	emulatorHost string
}

func (weh withEmulatorHost) Apply(auth *Auth) {
	auth.emulatorHost = weh.emulatorHost
}

// Attach an emulator to the authlication client
func WithEmulatorHost(emulatorHost string) AuthOption {
	return withEmulatorHost{emulatorHost}
}
