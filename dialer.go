package liltunnel

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Dialer abstracts any interface that can dial a remote machine and setup a
// connection
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// NewDialer returns a struct where the DialContext func can be used anywhere
// that expects that function signature.
//
// This dialer creates a connection to the remote machine over SSH.
func NewDialer(sshKey string, known string, username string, host string, l logger) (Dialer, error) {
	key, err := ioutil.ReadFile(sshKey)
	if err != nil {
		return nil, errors.Wrap(err, "could not open key file")
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse private key")
	}

	callback, err := knownhosts.New(known)
	if err != nil {
		return nil, errors.Wrap(err, "could not create knownhosts callback")
	}

	sshConf := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: callback,
	}

	return &dialer{clientConfig: sshConf, host: host, log: l}, nil
}

// Dialer is a wrapper around a DialContext func that dials remote machines over
// SSH
type dialer struct {
	clientConfig *ssh.ClientConfig
	host         string
	log          logger
}

// DialContext matches the signature of net.DialContext and can be used anywhere
// that expectes a net.DialContext func
func (d *dialer) DialContext(ctx context.Context, n, addr string) (net.Conn, error) {
	d.log.Println("dialing host via SSH on port 22")

	client, err := ssh.Dial("tcp", fmt.Sprintf("%v:22", d.host), d.clientConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "could not open connection to %v on port 22", d.host)
	}

	c, err := client.Dial(n, addr)
	if err != nil {
		return nil, errors.Wrap(err, "failed dialing the remote server")
	}

	d.log.Println("passing SSH connection")
	return &conn{c, client, d.log}, nil
}
