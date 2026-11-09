//go:build !windows

package checkly

// platformConnectionErrors are the platform's own errors for a broken
// connection beyond the portable ones brokenConnection checks. Outside
// Windows, the portable ones cover them.
var platformConnectionErrors []error
