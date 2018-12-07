package anansi_test

import (
	"errors"
	"log"
	"os"
	"syscall"

	"github.com/jcorbin/anansi"
)

// Handling some of the usual terminal lifecycle signals.
func ExampleSignal() {

	var (
		halt   = anansi.Notify(syscall.SIGTERM, syscall.SIGHUP)
		stop   = anansi.Notify(syscall.SIGINT)
		resize = anansi.Notify(syscall.SIGWINCH)
	)

	term := anansi.NewTerm(
		os.Stdin, os.Stdout,
		&halt,
		&resize,
	)

	anansi.MustRun(term.RunWithFunc(func(term *anansi.Term) error {
		for {
			select {

			case sig := <-halt.C:
				// Terminate program asap on a halting signal; wrapping it in
				// anansi.SigErr provides transparency both internally, when
				// anansi.MustRun logs, and externally by setting the normative
				// exit code "killed by signal X" status code.
				return anansi.SigErr(sig)

			case <-stop.C:
				// Interrupt, on the other hand, may not always result in
				// immediate halt, it may only mean "stop / cancel whatever
				// operation is being currently run, but don't halt". Here we
				// just return a regular error, which will cause a normal "exit
				// code 1", externally hiding the fact that we stopped due to
				// SIGINT.
				return errors.New("stop")

			case <-resize.C:
				sz, _ := term.Size()
				log.Printf("Terminal resized to %v", sz)

			}
		}
	}))
}
