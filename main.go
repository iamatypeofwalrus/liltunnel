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

func main() {
	// todo configurable
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

	// todo configurable
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

	// todo configurable
	d := &dialer{clientConfig: sshConf, host: "138.68.203.191"}

	// todo configurable (port)
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
		DialContext:           d.DialContext,
	}
	rp.Transport = t

	// use this cache: https://github.com/lox/httpcache
	// todo configurable (port)
	log.Println("Listening :1080")
	log.Fatal(http.ListenAndServe(":1080", rp))
}
