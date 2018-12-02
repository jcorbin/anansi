package anansi_test

import (
	"log"
	"os"
	"syscall"

	"github.com/jcorbin/anansi"
)

func ExampleSignal() {

	var (
		done   = anansi.Notify(syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
		resize = anansi.Notify(syscall.SIGWINCH)

		term = anansi.NewTerm(
			os.Stdin, os.Stdout,
			&done,
			&resize,
		)
	)

	anansi.MustRun(term.RunWith(func(term *anansi.Term) error {
		for {
			select {

			case sig := <-done.C:
				return anansi.SigErr(sig)

			case <-resize.C:
				if sz, err := term.Size(); err != nil {
					log.Printf("Terminal resized; failed to get size: %v", err)
				} else {
					log.Printf("Terminal resized to %v", sz)
				}

			}
		}
	}))
}
