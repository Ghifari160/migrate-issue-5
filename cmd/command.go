package cmd

import (
	"flag"
	"path/filepath"
	"strings"
)

const ManifestName = "manifest.txt"
const ManifestSep = ";"
const PathSep = string(filepath.Separator)

type Cmd interface {
	Command(args []string) int
	Task() int
	Usage() string

	private() // prevent external functions from meeting the interface criteria
}

func NewFlagSet(name string) *flag.FlagSet {
	flag := flag.NewFlagSet(name, flag.ContinueOnError)
	flag.Usage = func() {}

	return flag
}

func PrintDefaults(f *flag.FlagSet) string {
	var s strings.Builder
	output := f.Output()

	f.SetOutput(&s)
	f.PrintDefaults()
	f.SetOutput(output)

	return s.String()
}

// PreserveTrailingSlash reintroduces trailing slash based on the original path string.
// If the original string ends with a trailing slash, it is reintroduced to the normalized string.
// Otherwise, the normalized string is returned as is.
func PreserveTrailingSlash(original, normalized string) string {
	if original[len(original)-1:] == PathSep {
		return normalized + PathSep
	}

	return normalized
}
