package lookfor

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Exe attempts to find an executable from user paths. If the given path is an executable path,
// it is converted to an absolute path.
func Exe(exe string) (string, bool) {
	if runtime.GOOS == "windows" && filepath.Ext(exe) != ".exe" {
		exe += ".exe"
	}

	if filepath.IsAbs(exe) {
		return exe, true
	}

	_, err := os.Stat(exe)
	if err == nil {
		abs, err := filepath.Abs(exe)
		if err != nil {
			return "", false
		}

		return abs, true
	}

	return lookForExe(exe)
}

// lookForExe attempts to find an executable from user paths.
func lookForExe(exe string) (string, bool) {
	var pathVar, pathVarSep string

	switch runtime.GOOS {
	case "windows":
		pathVar = "PATH"
		pathVarSep = ";"

	case "linux", "darwin":
		fallthrough

	default:
		pathVar = "PATH"
		pathVarSep = ":"
	}

	usrPath := os.Getenv(pathVar)
	paths := strings.Split(usrPath, pathVarSep)

	var exePath string

	for _, p := range paths {
		files, err := os.ReadDir(p)
		if err != nil {
			continue
		}

		for _, file := range files {
			if NameMatches(exe, file.Name()) && !file.IsDir() {
				exePath = filepath.Join(p, file.Name())
			}
		}
	}

	return exePath, len(exePath) > 0
}
