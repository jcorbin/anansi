package platform

import (
	"bytes"
	"log"
)

func init() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(&Logs)
}

// Logs is the LogSink installed as the output for the standard "logs" package.
var Logs LogSink

// LogSink implements an in-memory log buffer.
type LogSink struct {
	bytes.Buffer // TODO capped buffer
}
