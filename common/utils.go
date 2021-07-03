package common
import "C"
import (
	"errors"
	"os"
	"syscall"
)

func NewErrorOrNil(name string, errno C.int) error {
	if errno > 0 {
		return os.NewSyscallError(name, syscall.Errno(errno))
	}
	if errno < 0 {
		// generic error for functions that don't set errno
		return errors.New(name + ": failure")
	}
	return nil
}
