package main

type options struct {
	LocalPort string `long:"local-port" short:"p"`
	Host      string `long:"host" short:"h"`
	User      string `long:"user" short:"u"`
	Cache     bool   `long:"cache" short:"c"`
}
