package liltunnel

// logger is a simple logging interface that is satisfied by the standard
// library log package
type logger interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
	Fatal(v ...interface{})
}
