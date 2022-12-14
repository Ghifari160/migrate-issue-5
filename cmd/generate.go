package cmd

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ghifari160/migrate/internal/exit"
	"github.com/ghifari160/migrate/internal/logger"
)

type CmdGenerate struct {
	f          *flag.FlagSet
	printFlags bool
	c          generateConf
	m          map[string]string
	src        string
	dest       string
	manifest   string
	log        *logger.Logger
}

type generateConf struct {
	overwrite bool
	relSrc    bool
	relDest   bool
}

func NewCmdGenerate() Cmd {
	return &CmdGenerate{
		f: NewFlagSet("generate"),
		c: generateConf{
			overwrite: false,
			relSrc:    true,
			relDest:   false,
		},
	}
}

func (c *CmdGenerate) Command(args []string) int {
	var err error

	c.f.BoolVar(&c.c.overwrite, "overwrite", c.c.overwrite, "Overwrite manifest.")
	c.f.BoolVar(&c.c.relSrc, "rel-src", c.c.relSrc, "Relative source path in the manifest.")
	c.f.BoolVar(&c.c.relDest, "rel-dest", c.c.relDest, "relative destination path in the manifest.")

	err = c.f.Parse(args)
	if err != nil {
		c.printFlags = true
		return exit.Usage
	}

	args = c.f.Args()

	if len(args) < 2 || len(args[0]) < 1 || len(args[1]) < 1 {
		return exit.Usage
	}

	c.src, err = filepath.Abs(args[0])
	if err != nil {
		return exit.NotFound
	}

	c.dest, err = filepath.Abs(args[1])
	if err != nil {
		return exit.NotFound
	}

	// reintroduce trailing slashes
	c.src = PreserveTrailingSlash(args[0], c.src)
	c.dest = PreserveTrailingSlash(args[1], c.dest)

	c.manifest = ManifestName
	if len(args) > 2 && len(args[2]) > 0 {
		c.manifest = args[2]
	}

	c.manifest, err = filepath.Abs(c.manifest)
	if err != nil {
		return exit.NotFound
	}

	c.log, err = logger.OpenLogs("logs")
	if err != nil {
		return exit.LogError
	}
	fmt.Println("Logging to " + c.log.Dir() + ".")

	return exit.RDY
}

func (c *CmdGenerate) Task() int {
	defer c.log.Close()

	var err error

	c.m = make(map[string]string)

	status := generate(c.log, c.c, c.m, c.src, c.dest)
	if status != exit.Norm {
		return status
	}

	flag := os.O_CREATE | os.O_WRONLY
	if !c.c.overwrite {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	c.log.Log(logger.LevelINFO, "Opening manifest file at "+c.manifest)
	file, err := os.OpenFile(c.manifest, flag, fs.FileMode(0644))
	if err != nil {
		c.log.Log(logger.LevelError, "Error opening manifest: "+err.Error())
		return exit.ManifestWrite
	}
	defer file.Close()

	rel := filepath.Dir(c.manifest)
	for src, dest := range c.m {
		c.log.Log(logger.LevelINFO, "Writing manifest entry for "+src)

		if c.c.relSrc {
			relSrc, err := filepath.Rel(rel, src)
			if err != nil {
				c.log.Log(logger.LevelError, "Unable to create relative src path for "+src+". Skipping.")
				c.log.File(src).Log(logger.LevelError, "Error creating src relative path: "+err.Error())
				continue
			}

			// reintroduce trailing slash
			src = PreserveTrailingSlash(src, relSrc)
		}

		if c.c.relDest {
			relDest, err := filepath.Rel(rel, dest)
			if err != nil {
				c.log.Log(logger.LevelError, "Unable to create relative dest path for "+src+". Skipping.")
				c.log.File(src).Log(logger.LevelError, "Error creating dest relative path: "+err.Error())
				continue
			}

			// reintroduce trailing slash
			dest = PreserveTrailingSlash(dest, relDest)
		}

		_, err = file.WriteString(src + ManifestSep + dest + "\n")
		if err != nil {
			c.log.Log(logger.LevelError, "Unable to create manifest entry for "+src+". Skipping.")
			c.log.File(src).Log(logger.LevelError, "Error creating manifest entry: "+err.Error())
			continue
		}
	}

	return exit.Norm
}

// generate generates a source->dest mapping for the given src and dest and returns status code.
// If src is a dir, it is scanned and the mappings will be for its children instead.
// This is NOT done recursively.
// It is up to the copying utility to handle directories.
func generate(log *logger.Logger, config generateConf, m map[string]string, src, dest string) int {
	var err error
	var files []string

	log.Log(logger.LevelINFO, "Generating mapping for "+src)

	s, err := os.Stat(src)
	if err != nil {
		log.Log(logger.LevelError, "Error generating mapping for "+src)
		log.File(src).Log(logger.LevelError, "Error generating mapping: "+err.Error())

		return exit.ManifestWrite
	}

	var isDot bool
	if s.IsDir() {
		isDot, err = isCwd(src)
		if err != nil {
			log.Log(logger.LevelError, "Error checking directory "+src)
			log.File(src).Log(logger.LevelError, "Error checking directory: "+err.Error())

			return exit.ManifestWrite
		}

		if hasTrailingSlash(src) || isDot {
			f, err := os.ReadDir(src)
			if err != nil {
				log.Log(logger.LevelError, "Error reading directory contents for "+src)
				log.File(src).Log(logger.LevelError, "Error reading directory: "+err.Error())

				return exit.ManifestWrite
			}

			files = make([]string, 0, len(f))
			for _, file := range f {
				files = append(files, filepath.Join(src, file.Name()))
			}
		}
	}

	if len(files) < 1 && !hasTrailingSlash(src) {
		files = make([]string, 1)
		files[0] = src
	}

	for _, file := range files {
		log.Log(logger.LevelINFO, "Found "+file)
		m[file] = dest
	}

	return exit.Norm
}

func (c *CmdGenerate) Usage() string {
	usage := "  migrate generate SRC DEST [MANIFEST]\n"

	if c.printFlags {
		usage += "\nFLAGS:\n\n" + PrintDefaults(c.f)
	}

	return usage
}

func (c *CmdGenerate) private() {}
