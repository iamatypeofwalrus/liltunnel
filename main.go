package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/user"
	"path"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"

	flags "github.com/jessevdk/go-flags"
)

const (
	confFile       = ".liltunnel"
	cacheFile      = ".liltunnel-cache.json"
	knownHostsFile = "known_hosts"
)

type options struct {
	Port           string `long:"port" short:"p" required:"true" description:"all traffic from local host will be sent to this port on foreign host"`
	Host           string `long:"host-name" short:"n" required:"true" description:"valid DNS name or IP address"`
	User           string `long:"user" short:"u" required:"false" description:"DEFAULT: current user name. user name used to ssh into the host."`
	SSHKeyPath     string `long:"ssh-key" short:"k" required:"true" description:"path to the private key to be used when establishing a connection to the host"`
	KnownHostsPath string //`long:"known-hosts" short:"n" required:"false"`
	// Cache     bool   `long:"cache" short:"c"`
}

func main() {
	opts := options{}
	_, err := flags.Parse(&opts)
	if err != nil {
		fmt.Println("could not parse arguments: ", err)
		os.Exit(1)
	}

	key, err := ioutil.ReadFile(opts.SSHKeyPath)
	if err != nil {
		fmt.Println("could not open key file: ", err)
		os.Exit(1)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		fmt.Println("could not parse private key: ", err)
		os.Exit(1)
	}

	var knownHostsPath string
	if opts.KnownHostsPath == "" {
		u, err := user.Current()
		if err != nil {
			fmt.Println("could not get current user: ", err)
			os.Exit(1)
		}
		knownHostsPath = path.Join(u.HomeDir, ".ssh", knownHostsFile)

	} else {
		knownHostsPath = opts.KnownHostsPath
	}

	callback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		fmt.Println("could not create knownhosts callback: ", err)
		os.Exit(1)
	}

	var sshUser string
	if opts.User == "" {
		u, err := user.Current()
		if err != nil {
			fmt.Println("could not get current user: ", err)
			os.Exit(1)
		}

		sshUser = u.Username
	} else {
		sshUser = opts.User
	}
	sshConf := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: callback,
	}

	d := &dialer{clientConfig: sshConf, host: opts.Host}

	url, err := url.Parse(
		fmt.Sprintf("http://localhost:%v", opts.Port),
	)
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
	log.Println("listening on", opts.Port)
	log.Fatal(http.ListenAndServe(":"+opts.Port, rp))
}
