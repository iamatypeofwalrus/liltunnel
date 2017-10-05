package main

import "io"

type verboseWriter struct {
	verbose bool
	writer  io.Writer
}

func (v *verboseWriter) Write(p []byte) (n int, err error) {
	if !v.verbose {
		return len(p), nil
	}
	return v.writer.Write(p)
}
