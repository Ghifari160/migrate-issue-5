package cmd

import (
	"flag"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/ghifari160/migrate/internal/exit"
)

type CmdGenerate struct {
	f          *flag.FlagSet
	printFlags bool
	c          generateConf
	m          map[string]string
	src        string
	dest       string
	manifest   string
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

	args = flag.Args()

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

	c.manifest = ManifestName
	if len(args) > 2 && len(args[2]) > 0 {
		c.manifest = args[2]
	}

	c.manifest, err = filepath.Abs(c.manifest)
	if err != nil {
		return exit.NotFound
	}

	return exit.RDY
}

func (c *CmdGenerate) Task() int {
	var err error

	c.m = make(map[string]string)

	generate(c.c, c.m, c.src, c.dest)

	flag := os.O_CREATE | os.O_WRONLY
	if !c.c.overwrite {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	file, err := os.OpenFile(c.manifest, flag, fs.FileMode(0644))
	if err != nil {
		return exit.ManifestWrite
	}
	defer file.Close()

	rel := filepath.Dir(c.manifest)
	for src, dest := range c.m {
		if c.c.relSrc {
			src, err = filepath.Rel(rel, src)
			if err != nil {
				continue
			}
		}

		if c.c.relDest {
			dest, err = filepath.Rel(rel, dest)
			if err != nil {
				continue
			}
		}

		_, err = file.WriteString(src + ManifestSep + dest + "\n")
		if err != nil {
			continue
		}
	}

	return exit.Norm
}

// generate generates a source->dest mapping for the given src and dest. If src is a dir, it is
// scanned and the mappings will be for its children instead. This is NOT done recursively. It is
// up to the copying utility to handle directories.
func generate(config generateConf, m map[string]string, src, dest string) {
	var err error
	var files []string

	s, err := os.Stat(src)
	if err != nil {
		return
	}

	if s.IsDir() {
		f, err := os.ReadDir(src)
		if err != nil {
			return
		}

		files = make([]string, 0, len(f))
		for _, file := range f {
			files = append(files, filepath.Join(src, file.Name()))
		}
	} else {
		files = make([]string, 1)
		files[0] = src
	}

	for _, file := range files {
		m[file] = dest
	}
}

func (c *CmdGenerate) Usage() string {
	usage := "  migrate generate SRC DEST [MANIFEST]\n"

	if c.printFlags {
		usage += "\nFLAGS:\n\n" + PrintDefaults(c.f)
	}

	return usage
}

func (c *CmdGenerate) private() {}
