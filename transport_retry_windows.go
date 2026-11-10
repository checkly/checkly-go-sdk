package checkly

import "syscall"

// platformConnectionErrors are the Windows socket errors for a reset or
// aborted connection. syscall.ECONNRESET and syscall.ECONNABORTED do not match
// them on Windows, where they are separate values.
var platformConnectionErrors = []error{syscall.WSAECONNRESET, syscall.WSAECONNABORTED}
