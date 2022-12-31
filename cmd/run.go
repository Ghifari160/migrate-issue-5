package cmd

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ghifari160/migrate/internal/exit"
	"github.com/ghifari160/migrate/internal/logger"
)

type CmdMigrate struct {
	f          *flag.FlagSet
	printFlags bool
	m          *bufio.Reader
	closeM     func() error
	c          migrateConf
	log        *logger.Logger
}

type migrateConf struct {
	dryRun bool
	util   string
	args   string
}

func NewCmdMigrate() Cmd {
	var util, args string

	util = "rsync"
	args = "-avr"

	if runtime.GOOS == "windows" {
		util = "robocopy"
		args = "/E /COPY:DAT"
	}

	return &CmdMigrate{
		f: NewFlagSet("run"),
		c: migrateConf{
			util: util,
			args: args,
		},
	}
}

func (c *CmdMigrate) Command(args []string) int {
	var err error
	var manifest, src, dest string

	c.log, err = logger.OpenLogs("logs")
	if err != nil {
		return exit.LogError
	}
	fmt.Println("Logging to " + c.log.DirAbs() + ".")

	c.f.BoolVar(&c.c.dryRun, "dryrun", c.c.dryRun, "Run in dry run mode.")
	c.f.StringVar(&c.c.util, "util", c.c.util, "Copying utility.")
	c.f.StringVar(&c.c.args, "util-args", c.c.args, "Copying utility arguments.")

	err = c.f.Parse(args)
	if err != nil {
		c.printFlags = true
		return exit.Usage
	}

	args = c.f.Args()

	if len(args) < 1 {
		manifest = ManifestName
	} else if len(args) < 2 {
		if len(args[0]) < 1 {
			return exit.Usage
		}

		manifest = args[0]
	} else {
		if len(args[0]) < 1 || len(args[1]) < 1 {
			return exit.Usage
		}

		src = args[0]
		dest = args[1]
	}

	util, err := exec.LookPath(c.c.util)
	if err != nil || len(util) < 1 {
		return exit.UtilNotFound
	}
	c.c.util = util

	if len(manifest) > 0 {
		c.m, c.closeM, err = openManifest(c.log, manifest)
	} else {
		src, dest, norm := normPaths(src, dest)
		if !norm {
			c.log.Log(logger.LevelWARN, "Cannot normalize paths for "+src+" => "+dest+".")
			return exit.ManifestRead
		}

		c.m, c.closeM, err = manifestFromArgs(c.log, src, dest)
	}

	if err != nil {
		c.log.Log(logger.LevelError, "error reading manifest: "+err.Error())
		return exit.ManifestRead
	}

	return exit.RDY
}

func (c *CmdMigrate) Task() int {
	defer c.closeM()
	defer c.log.Close()

	if c.c.dryRun {
		fmt.Println("Running in dry run mode. Check logs.")
		c.log.Log(logger.LevelINFO, "Running in dry run mode.")
	}

	c.log.Log(logger.LevelINFO, "Copying files with "+c.c.util+".")

	lineN := 1
	eof := false

	for !eof {
		src, dest, err := readManifestEntry(c.m, &lineN)
		if err != nil {
			if errors.Is(err, errManifest) {
				c.log.Log(logger.LevelWARN, "Error: "+err.Error())
			} else if errors.Is(err, io.EOF) {
				eof = true
			} else {
				return exit.ManifestRead
			}

			continue
		}

		copy(c.log, c.c, src, dest)
	}

	return exit.Norm
}

// normPaths normalizes paths by converting them to absolute paths.
// Trailing slashes are reintroduced into the paths after normalizations.
func normPaths(src, dest string) (string, string, bool) {
	var err error

	aSrc, err := filepath.Abs(src)
	if err != nil {
		return "", "", false
	}

	aDest, err := filepath.Abs(dest)
	if err != nil {
		return "", "", false
	}

	// reintroduce trailing slashes
	src = PreserveTrailingSlash(src, aSrc)
	dest = PreserveTrailingSlash(dest, aDest)

	return src, dest, true
}

// openManifest opens the manifest and creates a buffered reader.
func openManifest(log *logger.Logger, manifest string) (*bufio.Reader, func() error, error) {
	m, err := os.Open(manifest)
	if err != nil {
		return nil, nil, err
	}

	return bufio.NewReader(m), m.Close, nil
}

// manifestFromArgs returns a manifest buffered reader from the given src and dest paths.
func manifestFromArgs(log *logger.Logger, src, dest string) (*bufio.Reader, func() error, error) {
	var buffer bytes.Buffer
	closeFn := func() error {
		buffer.Reset()

		return nil
	}

	buffer.WriteString(src + ManifestSep + dest)

	return bufio.NewReader(&buffer), closeFn, nil
}

// readManifestEntry reads and parses the next line of the manifest.
// The normalized source path and destination path are returned.
//
// readManifestEntry reads whole lines from the buffered reader when possible.
// If the lines are too long for a single read, multiple reads executed until the whole line has
// been read.
//
// lineN is advanced after a successful read and parse.
func readManifestEntry(m *bufio.Reader, lineN *int) (string, string, error) {
	var lineBuilder strings.Builder
	var lineBuffer []byte
	var err error
	isPrefix := true

	for isPrefix {
		lineBuffer, isPrefix, err = m.ReadLine()
		if err != nil {
			return "", "", err
		}

		lineBuilder.Write(lineBuffer)
	}

	defer func() { (*lineN)++ }()

	if lineBuilder.Len() < 1 {
		return "", "", newManifestErr(*lineN, "empty line")
	}

	mapping := strings.Split(lineBuilder.String(), ManifestSep)
	if len(mapping) < 2 || len(mapping[0]) < 1 || len(mapping[1]) < 1 {
		return "", "", newManifestErr(*lineN, "syntax error")
	}

	src, dest, norm := normPaths(mapping[0], mapping[1])
	if !norm {
		return "", "", newManifestErr(*lineN, "cannot normalize paths for "+src+" => "+dest)
	}

	return src, dest, nil
}

// readManifest parses the manifest and generates source->dest map.
//
// Deprecated: readManifest has been refactored to partially fix [#2], but it still stores the
// whole mappings into memory.
// Parsing the manifest should truly be done line-by-line by calling openManifest and
// readManifestEntry.
//
// [#2]: https://github.com/Ghifari160/migrate/issues/2
func readManifest(log *logger.Logger, manifest string) (map[string]string, error) {
	m, closeM, err := openManifest(log, manifest)
	if err != nil {
		return nil, err
	}
	defer closeM()

	mappings := make(map[string]string)
	lineN := 1
	eof := false

	for !eof {
		src, dest, err := readManifestEntry(m, &lineN)
		if err != nil {
			if errors.Is(err, errManifest) {
				log.Log(logger.LevelWARN, fmt.Sprintf("Error: %s. Skipping.", err))
			} else if errors.Is(err, io.EOF) {
				eof = true
			} else {
				return nil, err
			}

			continue
		}

		mappings[src] = dest
	}

	return mappings, nil
}

// copy copies the file by executing the copying utility. In dry mode, it instead prints the exec
// commands.
func copy(log *logger.Logger, config migrateConf, src, dest string) {
	if config.dryRun {
		entry := fmt.Sprintf("  %s %s %s %s", config.util, config.args, src, dest)
		log.Log(logger.LevelINFO, entry)
		return
	}

	log.Log(logger.LevelINFO, "Copying "+src+" to "+dest+".")

	cmd := exec.Command(config.util, config.args, src, dest)
	stdout, err := cmd.Output()
	if err != nil {
		entry := fmt.Sprintf("Error copying %s", src)

		log.Log(logger.LevelError, entry+".")

		log.File(src).Log(logger.LevelError, entry+": "+err.Error())
		log.File(src).Log(logger.LevelError, config.util+" output:")
		log.File(src).Write(stdout)
	}
}

func (c *CmdMigrate) Usage() string {
	usage := "  migrate run [FLAGS] SRC DEST\n  migrate run [FLAGS] [MANIFEST]\n"

	if c.printFlags {
		usage += "\nFLAGS:\n\n" + PrintDefaults(c.f)
	}

	return usage
}

func (c *CmdMigrate) private() {}
