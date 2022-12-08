package cmd

import (
	"flag"
	"strings"
)

const ManifestName = "manifest.txt"
const ManifestSep = ";"

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
