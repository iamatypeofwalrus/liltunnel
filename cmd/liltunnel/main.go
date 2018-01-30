package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/iamatypeofwalrus/liltunnel"
	flags "github.com/jessevdk/go-flags"
)

const (
	knownHostsFile      = "known_hosts"
	sshKey              = "id_rsa"
	usage               = "--port-mapping 80:8080 --remote me@remote.example.com --identity ~/.ssh/liltunnel_rsa"
	protocolHTTP        = "http"
	protocolTCP         = "tcp"
	defaultHTTPCacheTTL = 12 * time.Hour
)

// options represents the raw cli options exposed to the user
type options struct {
	PortMapping         string        `long:"port-mapping" short:"p" required:"true" description:"local:remote or port. If remote is not specified local port is used"`
	Remote              string        `long:"remote" short:"r" required:"true" description:"username@remote.example.com or remote.example.com. If username is not specified the current $USER is used"`
	SSHKeyPath          string        `long:"identity" short:"i" required:"true" description:"private key to be used when establishing a connection to the remote (default: ~/.ssh/id_rsa)"`
	KnownHostsPath      string        `long:"known-hosts" short:"o" required:"false" description:"known hosts file (default: ~/.ssh/known_hosts)"`
	Protocol            string        `long:"protocol" short:"n" required:"false" description:"network protocol to use when tunneling" default:"tcp" choice:"http" choice:"tcp"`
	HTTPCache           bool          `long:"http-cache" short:"c" description:"HTTP only. Cache all succesful responses to GET requests to disk"`
	HTTPCacheTTL        time.Duration `long:"http-cache-ttl" short:"t" description:"HTTP only. Expressed in seconds. Length of time to keep successful responses in cache. Defaults to 12 hours"`
	HTTPCacheServeStale bool          `long:"http-cache-serve-stale" short:"s" description:"HTTP only. Always return return a stale read from the cache. Handy if you need an offline mode"`
	Verbose             bool          `long:"verbose" short:"v"`
}

// config holds all of the parsed and default values sanely set from the options
// the user provided.
type config struct {
	sshKeyPath          string
	user                string
	localPort           string
	remotePort          string
	remote              string
	knownHosts          string
	protocol            string
	httpCache           bool
	httpCacheTTL        time.Duration
	httpCacheServeStale bool
	verbose             bool
}

func main() {
	conf, err := newConf()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse CLI options: %v\n", err)
		os.Exit(1)
	}

	l := log.New(
		&verboseWriter{verbose: conf.verbose, writer: os.Stdout},
		"",
		log.Lshortfile,
	)

	d, err := liltunnel.NewDialer(
		conf.sshKeyPath,
		conf.knownHosts,
		conf.user,
		conf.remote,
		l,
	)

	if err != nil {
		fmt.Println("could not init dialer", err)
		os.Exit(1)
	}

	var t liltunnel.Tunneler
	var tunnelErr error
	if conf.protocol == "http" {
		l.Printf("initalizing http tunnler")
		t, err = liltunnel.NewHTTPTunnler(
			d,
			conf.localPort,
			conf.remotePort,
			conf.httpCache,
			conf.httpCacheTTL,
			conf.httpCacheServeStale,
			l,
		)
	} else {
		l.Printf("initalizing tcp tunnler")
		t, err = liltunnel.NewTCPTunneler(d, ":2009", ":2009", l)
	}

	if tunnelErr != nil {
		fmt.Fprintf(os.Stderr, "could not initalize tunnel: %v", err)
		os.Exit(1)
	}

	l.Printf(
		"forwarding all %v traffic sent to localhost%v to %v@%v%v",
		conf.protocol,
		conf.localPort,
		conf.user,
		conf.remote,
		conf.remotePort,
	)
	l.Fatal(t.Tunnel())
}

func newConf() (config, error) {
	opts := parseArgs()
	conf := config{}

	currUser, err := user.Current()
	if err != nil {
		return conf, err
	}

	// PortMapping can look like "8080:80" or "8080". In the later case we implicitly
	// forward set the remote port to the same as local
	ports := strings.SplitN(opts.PortMapping, ":", 2)
	if len(ports) == 1 {
		conf.localPort = ":" + ports[0]
		conf.remotePort = ":" + ports[0]
	} else {
		conf.localPort = ":" + ports[0]
		conf.remotePort = ":" + ports[1]
	}
	// User and Remote address. User's pass in a string like "username@remote.example.com"
	// or "remote.example.com". If "username" is not provided we'll default to the
	// OS' current user.
	ur := strings.Split(opts.Remote, "@")
	if len(ur) == 1 {
		conf.remote = ur[0]
		conf.user = currUser.Username
	} else {
		conf.user = ur[0]
		conf.remote = ur[1]
	}

	// Set default private key
	if opts.SSHKeyPath == "" {
		conf.sshKeyPath = path.Join(
			currUser.HomeDir,
			".ssh",
			sshKey,
		)
	} else {
		conf.sshKeyPath = opts.SSHKeyPath
	}

	// Set known hosts
	if opts.KnownHostsPath == "" {
		conf.knownHosts = path.Join(
			currUser.HomeDir,
			".ssh",
			knownHostsFile,
		)
	} else {
		conf.knownHosts = opts.KnownHostsPath
	}

	conf.protocol = opts.Protocol
	if conf.protocol == protocolTCP && (opts.HTTPCache || opts.HTTPCacheTTL != 0 || opts.HTTPCacheServeStale) {
		return conf, errors.New("protocol TCP does not accept arguments http-cache, http-cache-ttl, http-cache-serve-stale")
	}

	if opts.Protocol == protocolHTTP {
		conf.httpCache = opts.HTTPCache
		conf.httpCacheTTL = opts.HTTPCacheTTL * time.Second
		conf.httpCacheServeStale = opts.HTTPCacheServeStale
	}

	if conf.httpCache && conf.httpCacheTTL == 0*time.Nanosecond {
		conf.httpCacheTTL = defaultHTTPCacheTTL
	}

	conf.verbose = opts.Verbose

	return conf, nil
}

func parseArgs() options {
	p := flags.NewNamedParser("liltunnel", flags.HelpFlag)
	p.Usage = usage

	opts := options{}
	_, err := p.AddGroup("Options", "", &opts)
	if err != nil {
		panic(
			fmt.Sprintf("could not add application options to group: %v", err),
		)
	}

	_, err = p.ParseArgs(os.Args)
	if err != nil {
		fmt.Printf("%v\n\n", err)
		p.WriteHelp(os.Stdout)
		os.Exit(1)
	}

	return opts
}
