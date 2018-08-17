package anansi

import (
	"os"
	"syscall"
)

func isEWouldBlock(err error) bool {
	switch val := err.(type) {
	case *os.PathError:
		err = val.Err
	case *os.LinkError:
		err = val.Err
	case *os.SyscallError:
		err = val.Err
	}
	return err == syscall.EWOULDBLOCK
}
