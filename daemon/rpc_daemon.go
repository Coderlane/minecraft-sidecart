package daemon

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"os/user"
	"path"

	"github.com/Coderlane/minecraft-sidecart/firebase"
)

type RPCDaemon struct {
	listener *net.UnixListener
	server   *rpc.Server
	errs     chan error
	conns    chan net.Conn
}

func NewRPCDaemon(ctx context.Context,
	app *firebase.App, auth *firebase.Auth) (*RPCDaemon, error) {
	daemon, err := NewDaemon(ctx, app, auth)
	if err != nil {
		return nil, err
	}
	addr, err := DefaultAddress()
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, err
	}

	server := rpc.NewServer()
	err = server.Register(daemon)
	if err != nil {
		return nil, err
	}

	return &RPCDaemon{
		listener: listener,
		server:   server,
		errs:     make(chan error, 1),
		conns:    make(chan net.Conn, 1),
	}, nil
}

func (dae *RPCDaemon) listen() {
	for {
		conn, err := dae.listener.Accept()
		if err != nil {
			dae.errs <- err
			continue
		}
		dae.conns <- conn
	}
}

func (dae *RPCDaemon) Run(ctx context.Context) error {
	go dae.listen()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)
	for {
		select {
		case <-sigc:
			return nil
		case <-ctx.Done():
			return nil
		case err := <-dae.errs:
			return err
		case conn := <-dae.conns:
			go dae.server.ServeConn(conn)
		}
	}
}

func (dae *RPCDaemon) Close() {
	dae.listener.Close()
}

var DefaultRootDir = os.TempDir()

func DefaultAddress() (*net.UnixAddr, error) {
	user, err := user.Current()
	if err != nil {
		return nil, err
	}
	return net.ResolveUnixAddr("unix",
		path.Join(DefaultRootDir, fmt.Sprintf("minecraft-sidecart-%s", user.Uid)))
}
