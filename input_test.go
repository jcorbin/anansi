package anansi_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcorbin/anansi"
	"github.com/jcorbin/anansi/ansi"
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

// Reading input in blocking mode, like a simple REPL-style program might.
//
// See cmd/decode/main.go for a more advanced example (including use of raw
// mode in addition to line-oriented as shown here).
func ExampleInput_blocking() {
	term := anansi.NewTerm(os.Stdin, os.Stdout)
	term.SetEcho(true)
	anansi.MustRun(term.RunWith(func(term *anansi.Term) error {
		for {
			n, err := term.ReadMore()

			// process any (maybe partial) input first before stopping on error
			if n > 0 {
				for i := 0; ; i++ {
					e, a := term.DecodeEscape()
					if e == 0 {
						r, ok := term.DecodeRune()
						if !ok {
							break
						}
						fmt.Printf("read[%v]: %q\n", i, r)
					} else if a != nil {
						fmt.Printf("read[%v]: %v %q\n", i, e, a)
					} else {
						fmt.Printf("read[%v]: %v\n", i, e)
					}
				}
			}

			if err != nil {
				return err // likely io.EOF
			}
		}
	}))
}

// Reading input in non-blocking mode 10 times a second, like an animated
// frame-rendering-loop program might.
func ExampleInput_nonblocking() {
	// We need to handle these signals so that we restore terminal state
	// properly (raw mode and exit the alternate screen).
	halt := anansi.Notify(syscall.SIGTERM, syscall.SIGINT)

	term := anansi.NewTerm(os.Stdin, os.Stdout, &halt)

	// run in a dedicated fullscreen, and handle input as it comes in
	term.SetRaw(true)
	term.AddMode(ansi.ModeAlternateScreen)

	anansi.MustRun(term.RunWith(func(term *anansi.Term) error {
		for range time.Tick(time.Second / 10) {
			if err := halt.AsErr(); err != nil {
				return err
			}

			n, err := term.ReadAny()

			// process any (maybe partial) input first before stopping on error
			if n > 0 {
				for i := 0; ; i++ {
					e, a := term.DecodeEscape()
					if e == 0 {
						r, ok := term.DecodeRune()
						if !ok {
							break
						}
						e = ansi.Escape(r)
					}

					// stop on Ctrl-C
					if e == 0x03 {
						return fmt.Errorf("read %v", e)
					}

					// simple character/escape-at-a-time handling
					// NOTE the need for CR below since we're in raw mode
					if !e.IsEscape() {
						fmt.Printf("read[%v]: %q\r\n", i, rune(e))
					} else if a != nil {
						fmt.Printf("read[%v]: %v %q\r\n", i, e, a)
					} else {
						fmt.Printf("read[%v]: %v\r\n", i, e)
					}
				}
			}

			if err != nil {
				return err // likely io.EOF
			}
		}
		return nil
	}))
}
