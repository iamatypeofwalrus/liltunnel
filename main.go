package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path"
)

const (
	confFile = ".lilcache"
)

func main() {
	u, err := user.Current()
	if err != nil {
		error("could not get current user: %s", err)
	}
	cacheFile := path.Join(u.HomeDir, confFile)

	// TODO: this should be CLI options that write out to this cache file
	//       next invocations of the binary without options will use conf
	f, err := ioutil.ReadFile(cacheFile)
	if err != nil {
		error("could not open cache config %s: %s", cacheFile, err)
	}

	var conf config
	err = json.Unmarshal(f, &conf)
	if err != nil {
		error("could not parse JSON in conf (%s): %s", cacheFile, err)
	}

	http.HandleFunc("/", handler(conf))

	log.Println("listening on ", conf.LocalPort)
	http.ListenAndServe(":"+conf.LocalPort, nil)
}

func handler(conf config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.RequestURI()
		if v, ok := conf.Cache[p]; ok {
			fmt.Fprint(w, v.Body)
		} else {
			err := fmt.Sprintf("could not find response body in cache for %s", p)
			http.Error(w, err, http.StatusNotFound)
			log.Println(err)
		}
	}
}

func error(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
