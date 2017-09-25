package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

const (
	confFile  = ".liltunnel"
	cacheFile = ".liltunnel-cache.json"
)

// Tunnel strategy
// create a closure that can use SSH inputs and returns a
// DialContext func(ctx context.Context, network, addr string) (net.Conn, error)
// Copy the http.DefaultTransport init and pass it to the HTTP server
//
// use transport in http.RoundTripper
//
// This may be fine for reverse proxy, but can it cache responses?
//
// possilby use this inconjuction with https://github.com/lox/httpcache
func main() {
	key, err := ioutil.ReadFile("/Users/joe/.ssh/digital_ocean_rsa")
	if err != nil {
		fmt.Println("could not open key file: ", err)
		os.Exit(1)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println("could not parse private key: ", err)
		os.Exit(1)
	}

	callback, err := knownhosts.New("/Users/joe/.ssh/known_hosts")
	if err != nil {
		fmt.Println("could not create knownhosts")
		os.Exit(1)
	}

	sshConf := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: callback,
	}

	client, err := ssh.Dial("tcp", "138.68.203.191:22", sshConf)
	if err != nil {
		fmt.Println("could not Dial via ssh the host: ", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("at some point we successfully dialed the host")

	// Construct HTTP proxy using the dialed client
	url, err := url.Parse("http://localhost:1080")
	if err != nil {
		fmt.Println("could not parse url: ", err)
		os.Exit(1)
	}
	rp := httputil.NewSingleHostReverseProxy(url)
	t := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		Dial: client.Dial,
	}
	rp.Transport = t

	log.Println("Listening :1080")
	log.Fatal(http.ListenAndServe(":1080", rp))
}
