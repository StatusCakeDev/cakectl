package cmd

import (
	"io/fs"
	"os"
)

// osFS is an [fs.FS] backed directly by the operating system's file system.
//
// [iooption.OpenFile] takes an [fs.FS] so that callers can substitute their
// own, but [os.DirFS] cannot serve the paths people type on a command line: it
// rejects absolute paths, paths beginning with "./" and any path containing
// "..". Names are therefore passed to the operating system unchanged.
type osFS struct{}

// Open opens the named file using the operating system's path rules.
func (osFS) Open(name string) (fs.File, error) {
	return os.Open(name)
}
