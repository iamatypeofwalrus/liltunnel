package main

import (
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

// conn exists as a wrapper struct so when http.Server calls Close() it will close the SSH connection and
// the original net.Conn
type conn struct {
	net.Conn
	client  *ssh.Client
	verbose bool
}

func (c *conn) Close() error {
	if c.verbose {
		log.Println("conn received Close()")
	}
	// and with this hacky method both connections will wind up closed
	// returning the net.Conn error as that is what _normally_ happens.
	defer c.client.Close()
	return c.Conn.Close()
}
