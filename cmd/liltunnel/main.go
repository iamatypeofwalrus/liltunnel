package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"github.com/iamatypeofwalrus/liltunnel"
	flags "github.com/jessevdk/go-flags"
)

const (
	confFile       = ".liltunnel.conf"
	cacheFile      = ".liltunnel.cache"
	knownHostsFile = "known_hosts"
	usage          = "--port 1080 --host-name the.best.example.com --ssh-key ~/.ssh/liltunnel_rsa"
)

type options struct {
	Port           string `long:"port" short:"p" required:"true" description:"all traffic from this port on localhost will be sent to the same port on the foreign host"`
	Host           string `long:"host-name" short:"n" required:"true" description:"valid DNS name or IP address"`
	SSHKeyPath     string `long:"ssh-key" short:"k" required:"true" description:"path to the private key to be used when establishing a connection to the host"`
	User           string `long:"user" short:"u" required:"false" description:"DEFAULT: current posix user name. User used to SSH into the foreign host"`
	KnownHostsPath string `long:"known-hosts" short:"o" required:"false" description:"DEFAULT: ~/.ssh/known_hosts. Path to known hosts file"`
	Verbose        bool   `long:"verbose" short:"v" description:"DEFAULT: false. Peak under the hood"`
	Cache          bool   `long:"cache" short:"c" description:"DEFAULT: false. Cache responses to disk"`
	TCP            bool
	HTTP           bool
	HTTPCache      bool
}

// TODO: save options to .liltunnel.conf and pick those up if liltunnel is run
//       without arguments again
func main() {
	opts := parseArgs()
	l := log.New(os.Stdout, "", log.Lshortfile)

	d, err := liltunnel.NewDialer(
		opts.SSHKeyPath,
		opts.KnownHostsPath,
		opts.User,
		opts.Host,
		opts.Verbose,
	)
	if err != nil {
		fmt.Println("could not init dialer", err)
		os.Exit(1)
	}

	t, err := liltunnel.NewHTTPTunnler(d, opts.Port, opts.Port, l)
	if err != nil {
		l.Fatal(err)
	}

	l.Fatal(t.Tunnel())
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

	return setDefaults(opts)
}

// setDefaults accepts an options struct and sets the default paramaters
// to valid values
func setDefaults(opts options) options {
	currUser, err := user.Current()
	if err != nil {
		// TODO: would be nice to keep everything that can terminate the program
		//       in the main function
		fmt.Println("could not get current user", err)
		os.Exit(1)
	}

	if opts.KnownHostsPath == "" {
		opts.KnownHostsPath = path.Join(currUser.HomeDir, ".ssh", knownHostsFile)
	}

	if opts.User == "" {
		opts.User = currUser.Username
	}

	return opts
}
