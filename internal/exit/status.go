package exit

const (
	Norm = iota
	RDY
	IncompatibleOS
	Usage
	ManifestRead
	ManifestWrite
	UtilNotFound
	NotFound
)

// Message returns the user friendly error message for the given exit code.
func Message(code int) string {
	switch code {
	case Norm, RDY:
		return ""

	case IncompatibleOS:
		return "OS not supported"

	case Usage:
		return "Usage:\n"

	case ManifestRead:
		return "Unable to read manifest"

	case ManifestWrite:
		return "Unable to write manifest"

	case UtilNotFound:
		return "Copying utility not found"

	case NotFound:
		return "File not found"

	default:
		return "Unknown error"
	}
}
