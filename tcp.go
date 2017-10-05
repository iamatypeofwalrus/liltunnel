package liltunnel

import (
	"context"
	"io"
	"net"

	"github.com/pkg/errors"
)

// NewTCPTunneler constructs a tunnler that handles all TCP traffic from localPort
// to remotePort using the given Dialer to establish the connection between the
// two machines.
func NewTCPTunneler(d Dialer, localPort string, remotePort string, l logger) (Tunneler, error) {
	t := &tcp{
		dialer:     d,
		localPort:  localPort,
		remotePort: remotePort,
		log:        l,
	}
	return t, nil
}

type tcp struct {
	log        logger
	dialer     Dialer
	localPort  string
	remotePort string
}

func (t *tcp) Tunnel() error {
	l, err := net.Listen("tcp", t.localPort)
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
	remote, err := t.dialer.DialContext(context.Background(), "tcp", t.remotePort)
	if err != nil {
		t.log.Println("could not dial remote server:", err)
		return
	}

	// Error handling?
	go io.Copy(local, remote)
	go io.Copy(remote, local)
}
