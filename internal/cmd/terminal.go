package cmd

import (
	"io"
	"os"
)

// isTerminal reports whether r is an interactive terminal rather than a pipe
// or a file.
func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
