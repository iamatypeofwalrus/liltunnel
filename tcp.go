package liltunnel

import (
	"context"
	"io"
	"net"

	"github.com/pkg/errors"
)

// NewTCPTunneler constructs a tunnler that handles any kind of TCP traffic
func NewTCPTunneler() (Tunneler, error) {
	return &tcp{}, nil
}

type tcp struct {
	address string
	log     Logger
	dialer  Dialer
}

func (t *tcp) Tunnel() error {
	l, err := net.Listen("tcp", t.address)
	if err != nil {
		t.log.Println("could not start listener:", err)
		return errors.Wrap(err, "could start listener")
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			t.log.Println("couldn't accept listener:", err)
			return errors.Wrap(err, "could not accept listener")
		}
		go t.handle(conn)
	}
}

func (t *tcp) handle(local net.Conn) {
	remote, err := t.dialer.DialContext(context.Background(), "tcp", t.address)
	if err != nil {
		t.log.Println("could not dial remote server:", err)
		return
	}

	// Error handling?
	go io.Copy(local, remote)
	go io.Copy(remote, local)
}
