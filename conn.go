package liltunnel

import (
	"net"

	"golang.org/x/crypto/ssh"
)

// conn exists as a wrapper struct around the ssh connection so when the consuming
// go routine has finished with the connection we also close the underlying ssh
// connection. This stops us from leaking SSH connections everytime a request
// is made.
type conn struct {
	net.Conn
	client *ssh.Client
	log    logger
}

func (c *conn) Close() error {
	c.log.Println("conn received Close()")
	// We're not really that interested in what the SSH connection has to say
	// when we close it so let's return the net.Conn's error message, if any.
	defer c.client.Close()
	return c.Conn.Close()
}
