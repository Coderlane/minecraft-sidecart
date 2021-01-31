package daemon

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	"golang.org/x/oauth2"

	"github.com/Coderlane/minecraft-sidecart/db"
	"github.com/Coderlane/minecraft-sidecart/firebase"
	"github.com/Coderlane/minecraft-sidecart/server"
)

var defaultPollInterval = time.Second * 5

type Daemon struct {
	ctx  context.Context
	auth *firebase.Auth
	db   db.Database
	mgr  *serverManager
}

func NewDaemon(ctx context.Context,
	app *firebase.App, auth *firebase.Auth) (*Daemon, error) {
	store, err := app.NewFirestore(ctx, auth)
	if err != nil {
		return nil, err
	}
	database, err := db.NewDatabase(ctx, store)
	if err != nil {
		return nil, err
	}
	mgr, err := newServerManager()
	if err != nil {
		return nil, err
	}
	dae := &Daemon{
		ctx:  ctx,
		auth: auth,
		db:   database,
		mgr:  mgr,
	}
	for id, srv := range mgr.servers {
		dae.monitorServer(srv, id)
	}
	return dae, nil
}

func (dae *Daemon) requireAuth() error {
	if dae.auth.CurrentUser() == nil {
		return fmt.Errorf("user is not authenticated")
	}
	return nil
}

func (dae *Daemon) SignIn(
	user *firebase.User, token *oauth2.Token) (err error) {
	_, err = dae.auth.SignInWithUser(dae.ctx, user)
	return err
}

type ServerSpec struct {
	Path string
	Name string
}

func (dae *Daemon) AddServer(
	spec ServerSpec, id *string) (err error) {
	if err := dae.requireAuth(); err != nil {
		return err
	}
	// Avoid collision
	if !filepath.IsAbs(spec.Path) {
		return fmt.Errorf("server path must be absolute")
	}
	if dae.mgr.hasPath(spec.Path) {
		return fmt.Errorf("server with path already exists")
	}
	// Setup a new server
	srv, err := server.NewServer(spec.Path)
	if err != nil {
		return err
	}
	tmpID, err := dae.db.CreateServer(dae.ctx,
		dae.auth.CurrentUser().UserID, spec.Name,
		server.GetType(srv), srv.GetServerInfo())
	if err != nil {
		return err
	}
	err = dae.mgr.addServer(tmpID, spec.Path, srv)
	if err != nil {
		return err
	}
	dae.monitorServer(srv, tmpID)
	*id = tmpID
	return nil
}

type serverUpdate struct {
	ID   string
	Info interface{}
}

func (dae *Daemon) monitorServer(srv server.Server, id string) {
	go func() {
		ticker := time.NewTicker(defaultPollInterval)
		var lastInfo interface{}
		for {
			select {
			case <-ticker.C:
				info := srv.GetServerInfo()
				if reflect.DeepEqual(info, lastInfo) {
					continue
				}
				fmt.Printf("Updating server info for: %s\n", id)
				dae.db.UpdateServerInfo(dae.ctx, id, info)
				lastInfo = info
			}
		}
	}()
}
