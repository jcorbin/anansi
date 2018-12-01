package anansi_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcorbin/anansi"
)

func TestInput_ReadMore(t *testing.T) {
	type write struct {
		d time.Duration
		s string
	}

	type read struct {
		n int
		s string
	}

	for _, tc := range []struct {
		name     string
		steps    []write
		expected []read
	}{

		{
			name: "hello world",
			steps: []write{
				{time.Millisecond, "hello"},
				{time.Millisecond, "world"},
			},
			expected: []read{
				{5, "hello"},
				{5, "world"},
			},
		},

		{
			name: "helloworld",
			steps: []write{
				{time.Millisecond, "hello"},
				{0, "world"},
			},
			expected: []read{
				{10, "helloworld"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var wg sync.WaitGroup
			defer wg.Wait()

			r, w, err := os.Pipe()
			require.NoError(t, err)
			defer w.Close()

			in := anansi.Input{File: r}

			wg.Add(1)
			go func() {
				defer wg.Done()
				defer r.Close()
				var got bytes.Buffer

				for i := 0; ; i++ {
					var ex read
					if i < len(tc.expected) {
						ex = tc.expected[i]
					}

					t.Logf("r[%v] reading more", i)
					n, err := in.ReadMore()

					got.Reset()
					slurpInput(&got, &in)
					t.Logf("r[%v] got %v %q", i, err, got.String())

					assert.Equal(t,
						ex,
						read{n, got.String()},
						"[%v] expected read result", i)

					// assert.Equal(t, ex.n, n, "[%v] expected number of bytes", i)
					// assert.Equal(t, ex.s, got.String(), "[%v] expected string", i)

					if err == io.EOF {
						break
					}
					require.NoError(t, err, "[%v] unexpected read error", i)
				}
			}()

			for i, step := range tc.steps {
				if step.d > 0 {
					t.Logf("w[%v] sleeping %v", i, step.d)
					time.Sleep(step.d)
				}
				if len(step.s) > 0 {
					t.Logf("w[%v] writing %q", i, step.s)
					_, err = w.WriteString(step.s)
					require.NoError(t, err, "[%v] failed to write string", i)
				}
			}

			t.Logf("w[ ] closing")
			require.NoError(t, w.Close(), "failed to close write pipe")
		})
	}
}

func slurpInput(buf *bytes.Buffer, in *anansi.Input) {
	for {
		e, a := in.DecodeEscape()
		if e == 0 {
			r, ok := in.DecodeRune()
			if !ok {
				break
			}
			buf.WriteRune(r)
		} else {
			fmt.Fprintf(buf, "[ansi %v %q]", e, a)
		}
	}
}
