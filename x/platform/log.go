package platform

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func init() {
	Logs.buf.Grow(1024 * 1024)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(&Logs)
}

// Logs is the LogSink installed as the output for the standard "logs" package.
var Logs LogSink

// LogSink implements an in-memory log buffer.
type LogSink struct {
	// TODO cap buffer size
	buf     bytes.Buffer
	bufEOLs []int

	f *os.File
}

// Contents all in-memory buffered bytes and a slice containing the index of
// all newlines within those bytes.
func (logs *LogSink) Contents() ([]byte, []int) {
	return logs.buf.Bytes(), logs.bufEOLs
}

func (logs *LogSink) Read(p []byte) (n int, err error) {
	n, err = logs.buf.Read(p)
	i := 0
	for ; i < len(logs.bufEOLs) && n > logs.bufEOLs[i]; i++ {
	}
	logs.bufEOLs = logs.bufEOLs[:copy(logs.bufEOLs, logs.bufEOLs[i:])]
	for i = 0; i < len(logs.bufEOLs); i++ {
		logs.bufEOLs[i] -= n
	}
	return n, err
}

func (logs *LogSink) Write(p []byte) (n int, _ error) {
	if len(p) == 0 {
		return 0, nil
	}
	b := logs.buf.Bytes()
	// unwind any implicit EOL-at-EOF
	if i := len(logs.bufEOLs) - 1; i >= 0 && b[logs.bufEOLs[i]] != '\n' {
		logs.bufEOLs = logs.bufEOLs[:i]
	}
	for off := 0; off < len(p); off++ {
		i := bytes.IndexByte(p[off:], '\n')
		if i < 0 {
			logs.bufEOLs = append(logs.bufEOLs, len(b)+len(p))
			break
		}
		off += i
		logs.bufEOLs = append(logs.bufEOLs, len(b)+off)
	}

	n, _ = logs.buf.Write(p)
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
