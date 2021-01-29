package firebase

// AuthOption is a generic interface used to apply options when creating a
// new authentication client.
type AuthOption interface {
	Apply(auth *Auth)
}

type withUserCache struct {
	uc UserCache
}

func (wuc withUserCache) Apply(auth *Auth) {
	auth.userCache = wuc.uc
}

// WithUserCache provides a user cache for the authentication client. This can
// be used to preserve users across a session.
func WithUserCache(uc UserCache) AuthOption {
	return withUserCache{uc}
}

type withEmulatorHost struct {
	emulatorHost string
}

func (weh withEmulatorHost) Apply(auth *Auth) {
	auth.emulatorHost = weh.emulatorHost
}

// WithEmulatorHost specifies the authentication emulator host to use in
// testing. Loading from environment variables is not supported.
func WithEmulatorHost(emulatorHost string) AuthOption {
	return withEmulatorHost{emulatorHost}
}
