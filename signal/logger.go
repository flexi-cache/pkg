package signal

import "log"

// Logger represents the logging APIs required by this package.
type Logger interface {
	Info(...interface{})
	Debug(...interface{})
}

type stdLogger struct{}

func (stdLogger) Debug(...interface{}) {}

func (stdLogger) Info(args ...interface{}) {
	log.Println(args...)
}
