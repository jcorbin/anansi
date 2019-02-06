package anansi_test

import (
	"bufio"
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
)

func TestCursor(t *testing.T) {
	type step struct {
		run    func(*VirtualCursor)
		expect string
	}
	for _, tc := range []struct {
		name  string
		steps []step
	}{

		{"hello", []step{
			{func(cur *VirtualCursor) {
				cur.To(ansi.Pt(5, 5))
				cur.WriteSGR(ansi.SGRRed.FG() | ansi.SGRGreen.BG())
				cur.WriteString("hello")
			}, "\x1b[5;5H\x1B[0;31;42mhello"},
			{func(cur *VirtualCursor) {
				cur.To(ansi.Pt(5, 6))
				cur.WriteSGR(ansi.SGRBlue.FG() | ansi.SGRGreen.BG())
				cur.WriteString("world")
			}, "\x1b[6;5H\x1b[34mworld"},
		}},
	} {
		t.Run(tc.name, logBuf.With(func(t *testing.T) {
			var out bytes.Buffer
			var cur VirtualCursor
			for i, step := range tc.steps {
				out.Reset()
				step.run(&cur)
				_, err := cur.WriteTo(&out)
				require.NoError(t, err)
				assert.Equal(t, step.expect, out.String(), "[%d] expected output", i)
			}
		}))
	}
}

type _logBuf struct {
	buf bytes.Buffer
	t   *testing.T
}

var logBuf _logBuf

func init() {
	log.SetOutput(&logBuf)
}

func (lb *_logBuf) With(f func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		lb.t = t
		defer func() {
			for sc := bufio.NewScanner(&lb.buf); sc.Scan(); {
				lb.t.Logf(sc.Text())
			}
			lb.t = nil
		}()
		f(t)
	}
}

func (lb *_logBuf) Write(p []byte) (n int, err error) {
	n, _ = lb.buf.Write(p)
	if lb.t != nil {
		for p := lb.buf.Bytes(); len(p) > 0; {
			i := bytes.IndexByte(p, '\n')
			if i < 0 {
				break
			}
			lb.t.Logf(string(p[:i]))
			p = p[i+1:]
			lb.buf.Next(i + 1)
		}
	}
	return n, nil
}
