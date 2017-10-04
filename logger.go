package liltunnel

// Logger is a simple logging interface that is satisfied by the standard
// library log package
type Logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
	Fatal(v ...interface{})
}
