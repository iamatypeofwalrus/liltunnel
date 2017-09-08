package main

// conf looks like this on disk:
// {
//   "host_scheme": "http",
//   "host_name": "feeneyj.aka.corp.amazon.com",
//   "host_port": "2010",
//   "local_port": "2009"
// }
//
// Interacting with a struct is idiomatic. in main() i'll parse the conf file
// into this struct with go magic.
type config struct {
	LocalPort string `json:"local_port"`
	Cache     cache  `json:"cache"`
}

type cache map[string]response

type response struct {
	Body string `json:"body"`
}
