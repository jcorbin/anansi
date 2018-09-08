package platform

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func init() {
	Logs.Buffer.Grow(1024 * 1024)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(&Logs)
}

// Logs is the LogSink installed as the output for the standard "logs" package.
var Logs LogSink

// LogSink implements an in-memory log buffer.
type LogSink struct {
	bytes.Buffer // TODO capped buffer

	f *os.File
}

func (logs *LogSink) Write(p []byte) (n int, _ error) {
	n, _ = logs.Buffer.Write(p)
	if logs.f != nil {
		return logs.f.Write(p)
	}
	return n, nil
}

// SetFile sets the sink's file destination.
func (logs *LogSink) SetFile(f *os.File) error {
	if logs.f == f {
		return nil
	}
	if logs.f != nil {
		if err := logs.f.Close(); err != nil {
			return fmt.Errorf("failed to close log file %q: %v", logs.f.Name(), err)
		}
	}
	if f != nil {
		log.Printf("logging to %q", f.Name())
		logs.f = f
		return nil
	}
	if logs.f != nil {
		log.Printf("disabling log file output")
		logs.f = nil
		return nil
	}
	return nil
}
