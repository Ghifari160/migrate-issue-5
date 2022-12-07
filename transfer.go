package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const defaultManifest = "COPY"
const defaultLogs = "LOGS"

const rsyncExec = "rsync"
const rsyncArgs = "-arv"

func main() {
	var manifest string
	var logsFile string
	var dest string
	var dryRun bool

	flag.StringVar(&manifest, "manifest", defaultManifest, "manifest")
	flag.StringVar(&logsFile, "logs", defaultLogs, "logs file")
	flag.StringVar(&dest, "dest", "", "destination")
	flag.BoolVar(&dryRun, "dryrun", false, "dryrun")
	flag.Parse()

	files := readManifest(manifest)

	logs := make([]string, 0, len(files))

	rsync, found := lookForExe(rsyncExec)
	if !found {
		panic("cannot find rsync")
	}

	for i, file := range files {
		if len(file) < 1 || file == "." || file == ".." {
			continue
		}

		fmt.Printf("Copying (%d of %d) %s...", i+1, len(files), file)

		if dryRun {
			fmt.Printf("\n  %s %s %s %s\n", rsync, rsyncArgs, file, dest)
		} else {
			cmd := exec.Command(rsync, rsyncArgs, file, dest)
			stdout, err := cmd.Output()
			if err != nil {
				logs = append(logs, fmt.Sprintf("ERROR: %v", err))
				fmt.Printf(" ERROR! %v\n", err)
			} else {
				fmt.Println(" DONE")
			}
			logs = append(logs, string(stdout))
		}
	}

	if !dryRun {
		writeLogs(logsFile, logs)
	}
}

func lookForExe(exe string) (string, bool) {
	var pathVariableSep string

	switch runtime.GOOS {
	case "windows":
		pathVariableSep = ";"
		break

	case "linux", "darwin":
		fallthrough

	default:
		pathVariableSep = ":"
		break
	}

	usrPath := os.Getenv("PATH")
	paths := strings.Split(usrPath, pathVariableSep)

	var exePath string

	for _, p := range paths {
		files, err := ioutil.ReadDir(p)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.Name() == exe && !file.IsDir() {
				exePath = filepath.Join(p, file.Name())
			}
		}
	}

	return exePath, len(exePath) > 0
}

func readManifest(manifest string) []string {
	m, err := os.ReadFile(manifest)
	if err != nil {
		panic(err)
	}

	files := strings.Split(string(m), "\n")
	sanitized := make([]string, 0, len(files))

	for _, file := range files {
		if len(file) > 0 && file != "." && file != ".." {
			sanitized = append(sanitized, file)
		}
	}

	return sanitized
}

func writeLogs(file string, entries []string) {
	var contents strings.Builder

	for _, entry := range entries {
		contents.WriteString(entry)
		contents.WriteString("\n")
	}

	err := os.WriteFile(file, []byte(contents.String()), 0644)
	if err != nil {
		panic(err)
	}
}
