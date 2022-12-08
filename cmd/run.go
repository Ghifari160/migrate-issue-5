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
	"github.com/ghifari160/migrate/internal/lookfor"
)

type CmdMigrate struct {
	f          *flag.FlagSet
	printFlags bool
	m          map[string]string
	c          migrateConf
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
		c.m, err = readManifest(manifest)
		if err != nil {
			return exit.ManifestRead
		}
	} else {
		src, dest, norm := normPaths(src, dest)
		if !norm {
			return exit.ManifestRead
		}

		c.m = make(map[string]string)
		c.m[src] = dest
	}

	return exit.RDY
}

func (c *CmdMigrate) Task() int {
	for src, dest := range c.m {
		copy(c.c, src, dest)
	}

	return exit.Norm
}

// normPaths normalizes paths by converting them to absolute paths.
func normPaths(src, dest string) (string, string, bool) {
	var err error

	src, err = filepath.Abs(src)
	if err != nil {
		return "", "", false
	}

	dest, err = filepath.Abs(dest)
	if err != nil {
		return "", "", false
	}

	return src, dest, true
}

// readManifest parses the manifest and generates source->dest map.
func readManifest(manifest string) (map[string]string, error) {
	m, err := os.ReadFile(manifest)
	if err != nil {
		return nil, err
	}

	mappings := make(map[string]string)
	lines := strings.Split(string(m), "\n")

	for _, line := range lines {
		if len(line) < 1 {
			continue
		}

		mapping := strings.Split(line, ManifestSep)
		if len(mapping) < 2 || len(mapping[0]) < 1 || len(mapping[1]) < 1 {
			continue
		}

		src, dest, norm := normPaths(mapping[0], mapping[1])
		if !norm {
			continue
		}

		mappings[src] = dest
	}

	return mappings, nil
}

// copy copies the file by executing the copying utility. In dry mode, it instead prints the exec
// commands.
func copy(config migrateConf, src, dest string) {
	if config.dryRun {
		fmt.Printf("  %s %s %s %s\n", config.util, config.args, src, dest)
		return
	}

	cmd := exec.Command(config.util, config.args, src, dest)
	stdout, err := cmd.Output()
	if err != nil {
		fmt.Printf("error: %v\n", err)
		fmt.Println(string(stdout))
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
