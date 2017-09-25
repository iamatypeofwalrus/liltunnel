package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
)

type dialer struct {
	clientConfig *ssh.ClientConfig
	host         string
	verbose      bool
}

func (d *dialer) DialContext(ctx context.Context, n, addr string) (net.Conn, error) {
	if d.verbose {
		log.Println("dialing host via SSH on port 22")
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:22", d.host), d.clientConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open connection to %v on :22", d.host)
	}

	c, err := client.Dial(n, addr)
	if err != nil {
		return nil, err
	}

	if d.verbose {
		log.Println("passing SSH connection to HTTP server")
	}
	return &conn{c, client, d.verbose}, nil
}
