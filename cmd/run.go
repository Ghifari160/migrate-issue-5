package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ghifari160/migrate/internal/exit"
	"github.com/ghifari160/migrate/internal/logger"
	"github.com/ghifari160/migrate/internal/lookfor"
)

type CmdMigrate struct {
	f          *flag.FlagSet
	printFlags bool
	m          map[string]string
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

	util, found := lookfor.Exe(c.c.util)
	if !found {
		return exit.UtilNotFound
	}
	c.c.util = util

	if len(manifest) > 0 {
		c.m, err = readManifest(c.log, manifest)
		if err != nil {
			c.log.Log(logger.LevelError, "error reading manifest: "+err.Error())
			return exit.ManifestRead
		}
	} else {
		src, dest, norm := normPaths(src, dest)
		if !norm {
			c.log.Log(logger.LevelWARN, "Cannot normalize paths for "+src+" => "+dest+".")
			return exit.ManifestRead
		}

		c.m = make(map[string]string)
		c.m[src] = dest
	}

	return exit.RDY
}

func (c *CmdMigrate) Task() int {
	defer c.log.Close()

	if c.c.dryRun {
		fmt.Println("Running in dry run mode. Check logs.")
		c.log.Log(logger.LevelINFO, "Running in dry run mode.")
	}

	c.log.Log(logger.LevelINFO, "Copying files with "+c.c.util+".")

	for src, dest := range c.m {
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

// readManifest parses the manifest and generates source->dest map.
func readManifest(log *logger.Logger, manifest string) (map[string]string, error) {
	m, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}

	mappings := make(map[string]string)
	lines := strings.Split(string(m), "\n")

	for i, line := range lines {
		if len(line) < 1 {
			continue
		}

		mapping := strings.Split(line, ManifestSep)
		if len(mapping) < 2 || len(mapping[0]) < 1 || len(mapping[1]) < 1 {
			log.Log(logger.LevelWARN, fmt.Sprintf("Manifest syntax error at line %d. Skipping.", i))
			continue
		}

		src, dest, norm := normPaths(mapping[0], mapping[1])
		if !norm {
			log.Log(logger.LevelWARN, "Cannot normalize paths for "+src+" => "+dest+". Skipping.")
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
