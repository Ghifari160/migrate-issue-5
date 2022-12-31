package cmd

import "fmt"

// errManifest is a dummy manifest error for use with error checking.
var errManifest = newManifestErr(-1, "")

// manifestErr is an error type for manifest reading and parsing.
type manifestErr struct {
	line int
	msg  string
}

// newManifestError returns an error from the given line number and text.
// Each call returns a distinct error value
func newManifestErr(line int, msg string) error {
	return &manifestErr{line: line, msg: msg}
}

func (e *manifestErr) Error() string {
	return fmt.Sprintf("%s at line %d", e.msg, e.line)
}

func (e *manifestErr) Is(target error) bool {
	if _, ok := target.(*manifestErr); ok {
		return true
	}

	return false
}
