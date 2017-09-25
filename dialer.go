package main

import (
	"context"
	"fmt"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type dialer struct {
	clientConfig *ssh.ClientConfig
	host         string
}

func (d *dialer) DialContext(ctx context.Context, n, addr string) (net.Conn, error) {
	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:22", d.host), d.clientConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open connection to %v on :22", d.host)
	}

	c, err := client.Dial(n, addr)
	if err != nil {
		return nil, err
	}

	return &conn{c, client}, nil
}
