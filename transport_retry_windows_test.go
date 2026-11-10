package checkly

import (
	"net"
	"os"
	"syscall"
	"testing"
)

func TestConnectionRetryRecognizesWindowsResets(t *testing.T) {
	for name, errno := range map[string]syscall.Errno{
		"reset":   syscall.WSAECONNRESET,
		"aborted": syscall.WSAECONNABORTED,
	} {
		err := &net.OpError{Op: "read", Net: "tcp", Err: os.NewSyscallError("wsarecv", errno)}
		if !brokenConnection(err) {
			t.Errorf("%s: want a broken connection", name)
		}
	}
}
