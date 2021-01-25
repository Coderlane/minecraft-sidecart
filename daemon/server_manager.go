package daemon

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/Coderlane/minecraft-sidecart/server"
)

// ConfigPaths holds paths to look for the daemon configuration file.
// We'll try the first entry and then the next until we find a valid entry.
// If no entry is found, we'll create an empty config and save it in the last
// path.
var ConfigPaths = []string{
	"/etc/minecraft-sidecart/daemon.json",
	"$HOME/.config/minecraft-sidecart/daemon.json",
}

type serverConfig struct {
	Path string
}

type config struct {
	Servers map[string]serverConfig
}

type serverManager struct {
	cfg     config
	cfgPath string
	servers map[string]server.Server
}

func newServerManager() (*serverManager, error) {
	mgr := &serverManager{
		cfg: config{
			Servers: make(map[string]serverConfig),
		},
		servers: make(map[string]server.Server),
	}
	if err := mgr.loadConfig(); err != nil {
		return nil, err
	}
	for id, srvCfg := range mgr.cfg.Servers {
		srv, err := server.NewServer(srvCfg.Path)
		if err != nil {
			continue
		}
		mgr.servers[id] = srv
	}
	return mgr, nil
}

func (mgr *serverManager) loadConfig() error {
	for _, cfgPath := range ConfigPaths {
		mgr.cfgPath = os.ExpandEnv(cfgPath)
		data, err := ioutil.ReadFile(mgr.cfgPath)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(data, &mgr.cfg); err != nil {
			return err
		}
	}
	return nil
}

func (mgr *serverManager) saveConfig() error {
	data, err := json.MarshalIndent(mgr.cfg, "", "    ")
	if err != nil {
		return err
	}
	if _, err := os.Stat(mgr.cfgPath); os.IsNotExist(err) {
		os.Mkdir(path.Dir(mgr.cfgPath), 0600)
	}
	return ioutil.WriteFile(mgr.cfgPath, data, 0600)
}

func (mgr *serverManager) hasPath(path string) bool {
	for _, existingServer := range mgr.cfg.Servers {
		if existingServer.Path == path {
			return true
		}
	}
	return false
}

func (mgr *serverManager) addServer(id, path string, srv server.Server) error {
	mgr.cfg.Servers[id] = serverConfig{
		Path: path,
	}
	mgr.servers[id] = srv
	return mgr.saveConfig()
}
