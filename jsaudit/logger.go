package main

import (
	"fmt"
	"os"
)

type logger struct {
	enabled bool
}

func newLogger(enabled bool) *logger {
	return &logger{enabled: enabled}
}

func (l *logger) Debugf(format string, args ...interface{}) {
	if !l.enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "[DBG] "+format+"\n", args...)
}

