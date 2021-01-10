package server

import (
	"fmt"
	"path"

	"github.com/fsnotify/fsnotify"

	config "github.com/Coderlane/go-minecraft-config"
)

type minecraftConfigType int

const (
	minecraftConfigTypeServer minecraftConfigType = iota
	minecraftConfigTypeDenyIP
	minecraftConfigTypeDenyUser
	minecraftConfigTypeOperatorUser
	minecraftConfigTypeAllowUser
)

type minecraftConfigLoadFunction func(string) (interface{}, error)

var configFileToType = map[string]minecraftConfigType{
	config.MinecraftConfigFile:       minecraftConfigTypeServer,
	config.MinecraftDenyIPFile:       minecraftConfigTypeDenyIP,
	config.MinecraftDenyUserFile:     minecraftConfigTypeDenyUser,
	config.MinecraftOperatorUserFile: minecraftConfigTypeOperatorUser,
	config.MinecraftAllowUserFile:    minecraftConfigTypeAllowUser,
}

type configEvent struct {
	Type   minecraftConfigType
	Config interface{}
}

type minecraftConfigWatcher struct {
	ConfigEvents chan configEvent
	Errors       chan error

	watcher *fsnotify.Watcher
	dir     string
	cancel  chan struct{}
}

func newMinecraftConfigWatcher(dir string) (*minecraftConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for config := range configFileToType {
		err := watcher.Add(path.Join(dir, config))
		if err != nil {
			return nil, err
		}
	}

	cfgWatcher := &minecraftConfigWatcher{
		ConfigEvents: make(chan configEvent),
		Errors:       make(chan error),
		watcher:      watcher,
		dir:          dir,
		cancel:       make(chan struct{}),
	}
	go cfgWatcher.start()
	return cfgWatcher, nil
}

func (cfgWatcher *minecraftConfigWatcher) start() {
	for {
		select {
		case event := <-cfgWatcher.watcher.Events:
			cfgEvent, err := cfgWatcher.reloadConfig(event)
			if err != nil {
				cfgWatcher.Errors <- err
			} else {
				cfgWatcher.ConfigEvents <- *cfgEvent
			}
		case err := <-cfgWatcher.watcher.Errors:
			cfgWatcher.Errors <- err
		case <-cfgWatcher.cancel:
			return
		}
	}
}

func (cfgWatcher *minecraftConfigWatcher) reloadConfig(
	event fsnotify.Event) (*configEvent, error) {
	var (
		cfgType minecraftConfigType
		cfg     interface{}
		err     error
	)
	cfgType, ok := configFileToType[path.Base(event.Name)]
	if !ok {
		return nil, fmt.Errorf("unknown config file: %s", event.Name)
	}
	switch cfgType {
	case minecraftConfigTypeServer:
		cfg, err = config.LoadConfig(cfgWatcher.dir)
	case minecraftConfigTypeDenyIP:
		cfg, err = config.LoadDenyIPList(cfgWatcher.dir)
	case minecraftConfigTypeDenyUser:
		cfg, err = config.LoadDenyUserList(cfgWatcher.dir)
	case minecraftConfigTypeOperatorUser:
		cfg, err = config.LoadOperatorUserList(cfgWatcher.dir)
	case minecraftConfigTypeAllowUser:
		cfg, err = config.LoadAllowUserList(cfgWatcher.dir)
	}
	if err != nil {
		return nil, err
	}
	return &configEvent{
		Type:   cfgType,
		Config: cfg,
	}, nil
}

func (cfgWatcher *minecraftConfigWatcher) Stop() {
	close(cfgWatcher.cancel)
}
