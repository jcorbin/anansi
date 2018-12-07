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

func TestInput_ReadAny(t *testing.T) {
	type write struct {
		d time.Duration
		s string
	}

	type read struct {
		n int
		s string
	}

	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer w.Close()
	defer r.Close()

	in := anansi.Input{File: r}

	steps := []struct {
		read
		write
	}{
		{write: write{s: "hello"}},
		{read: read{s: "hello", n: 5}},
		{write: write{s: "world"}},
		{read: read{s: "world", n: 5}},
	}

	var got bytes.Buffer
	for i, step := range steps {
		if step.write.s != "" {
			_, err := w.WriteString(step.write.s)
			require.NoError(t, err, "w[%v] unexpected error", i)
			if step.read.s == "" {
				continue
			}
		}

		n, err := in.ReadAny()

		got.Reset()
		slurpInput(&got, &in)
		t.Logf("r[%v] got %v %q", i, err, got.String())

		assert.Equal(t, step.read, read{n, got.String()}, "r[%v] expected read result", i)

		if err == io.EOF {
			i++
			assert.True(t, i <= len(steps)+1, "r[ ] expected all chunks")
			break
		}
		require.NoError(t, err, "r[%v] unexpected error", i)
	}

}

func slurpInput(buf *bytes.Buffer, in *anansi.Input) {
	for {
		if e, a, ok := in.Decode(); !ok {
			break
		} else if e.IsEscape() {
			fmt.Fprintf(buf, "[ansi %v %q]", e, a)
		} else {
			buf.WriteRune(rune(e))
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
	anansi.MustRun(term.RunWithFunc(func(term *anansi.Term) error {
		for {
			// process any (maybe partial) input first before stopping on error
			_, err := term.ReadMore()

			i := 0
			for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
				if !e.IsEscape() {
					fmt.Printf("read[%v]: %q\n", i, rune(e))
				} else if a != nil {
					fmt.Printf("read[%v]: %v %q\n", i, e, a)
				} else {
					fmt.Printf("read[%v]: %v\n", i, e)
				}
				i++
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

	anansi.MustRun(term.RunWithFunc(func(term *anansi.Term) error {
		for range time.Tick(time.Second / 10) {
			// poll for halting signal before reading input
			if err := halt.AsErr(); err != nil {
				return err
			}

			// process any (maybe partial) input first before stopping on error
			// NOTE the need for CR below since we're in raw mode
			_, err := term.ReadAny()

			i := 0
			for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
				if e == 0x03 {
					return fmt.Errorf("read %v", e) // stop on Ctrl-C
				} else if !e.IsEscape() {
					fmt.Printf("read[%v]: %q\r\n", i, rune(e))
				} else if a != nil {
					fmt.Printf("read[%v]: %v %q\r\n", i, e, a)
				} else {
					fmt.Printf("read[%v]: %v\r\n", i, e)
				}
				i++
			}

			if err != nil {
				return err // likely io.EOF
			}
		}
		return nil
	}))
}

// Reading input driven by asynchronous notifications (and doing so in the
// normative non-blocking way). This is another option for a fullscreen program
// when there's no need for an animation frame rendering loop or when input is
// processed independently from one.
func ExampleInput_nonblockingAsync() {
	// We need to handle these signals so that we restore terminal state
	// properly (raw mode and exit the alternate screen).
	halt := anansi.Notify(syscall.SIGTERM, syscall.SIGINT)

	term := anansi.NewTerm(os.Stdin, os.Stdout, &halt)

	// run in a dedicated fullscreen, and handle input as it comes in
	term.SetRaw(true)
	term.AddMode(ansi.ModeAlternateScreen)

	anansi.MustRun(term.RunWithFunc(func(term *anansi.Term) error {
		canRead := make(chan os.Signal, 1)
		if err := term.Notify(canRead); err != nil {
			return err
		}

		for {
			select {

			case sig := <-halt.C:
				return anansi.SigErr(sig)

			case <-canRead:
				// process any (maybe partial) input first before stopping on error
				// NOTE the need for CR below since we're in raw mode
				_, err := term.ReadAny()

				i := 0
				for e, a, ok := term.Decode(); ok; e, a, ok = term.Decode() {
					if e == 0x03 {
						return fmt.Errorf("read %v", e) // stop on Ctrl-C
					} else if !e.IsEscape() {
						fmt.Printf("read[%v]: %q\r\n", i, rune(e))
					} else if a != nil {
						fmt.Printf("read[%v]: %v %q\r\n", i, e, a)
					} else {
						fmt.Printf("read[%v]: %v\r\n", i, e)
					}
					i++
				}

				if err != nil {
					return err // likely io.EOF
				}
			}
		}
	}))
}
