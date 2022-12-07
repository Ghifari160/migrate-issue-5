package cmd

const ManifestName = "manifest.txt"
const ManifestSep = ";"

type Cmd interface {
	Command(args []string) int
	Task() int
	Usage() string

	private() // prevent external functions from meeting the interface criteria
}
