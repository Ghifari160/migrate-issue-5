package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const verFileName = "version_generated.go"
const winresFileName = "winres.json"

const dirPerm = fs.FileMode(0755)
const commonPerm = fs.FileMode(0644)

var verPath string
var winresPath string

var lastErrorStep string

const (
	goWinresImport = "github.com/tc-hib/go-winres"
	goWinresVer    = "0.2.3"
)

var goWinresPath string

type configs struct {
	Tool          string
	Version       string
	Authors       string
	Copyright     string
	CopyrightYear int
}

func main() {
	var confirm string
	var noninteractive bool

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s ERROR: %v\n", lastErrorStep, r)
			os.Exit(1)
		}
	}()

	verPath = filepath.Join("internal", "ver")
	winresPath = filepath.Join("winres")

	conf := getConfigs()

	conf.Version = "0.1.0"
	conf.CopyrightYear = time.Now().Year()

	flag.StringVar(&conf.Version, "ver", conf.Version, "Set version.")
	flag.IntVar(&conf.CopyrightYear, "copyyear", conf.CopyrightYear, "Set copyright year.")
	flag.BoolVar(&noninteractive, "noninteractive", false, "Noninteractive mode.")

	flag.Parse()

	if !noninteractive {
		fmt.Printf("Version: (%s) ", conf.Version)
		fmt.Scanln(&conf.Version)

		fmt.Printf("Copyright Year: (%d) ", conf.CopyrightYear)
		fmt.Scanln(&conf.CopyrightYear)

		prompt := "\nVersion: %s\nCopyright Year: %d\nCorrect? (y/n) "
		prompt = fmt.Sprintf(prompt, conf.Version, conf.CopyrightYear)
		confirm = promptUntilValid(prompt, []string{"y", "n"}, false)

		fmt.Println()
	}

	if confirm == "n" {
		os.Exit(0)
	}

	conf.Copyright = strings.Replace(conf.Copyright, `%s`, conf.Authors, -1)
	conf.Copyright = strings.Replace(conf.Copyright, `%d`,
		fmt.Sprintf("%d", conf.CopyrightYear), -1)

	fmt.Println("prerelease script")
	fmt.Println()

	stepSetupGoWinres(conf)
	stepGenVer(conf)
	stepGenWinres(conf)

	fmt.Println()
	fmt.Println("Done")
}

// handleError automatically raises panics when it encounters an error.
func handleError(step string, err error) {
	if err != nil {
		lastErrorStep = step

		panic(err)
	}
}

// promptUntilValid continually prompts for user input until the input is valid.
func promptUntilValid(prompt string, valid []string, caseSensitive bool) string {
	var input string
	var inputValid bool

	for {
		fmt.Print(prompt)
		fmt.Scanln(&input)

		inputValid = false
		for _, v := range valid {
			if inputValid {
				break
			}

			if !caseSensitive {
				input = strings.ToLower(input)
				v = strings.ToLower(v)
			}

			inputValid = v == input
		}

		if inputValid {
			break
		}
	}

	return input
}

// stepGenVer generates version data.
func stepGenVer(conf configs) {
	path := filepath.Join(verPath, verFileName)
	ts := time.Now().Format("2006/01/02 03:04:05 PM")

	fmt.Println("generate version data")
	payload := fmt.Sprintf(verFile, ts, conf.CopyrightYear, conf.Version)

	fmt.Println("write version data")
	err := os.WriteFile(path, []byte(payload), commonPerm)
	handleError("write version data", err)
}

// scanSrc recursively scans the path for valid Go sources.
func scanSrc(path string) []string {
	srcs := make([]string, 0)

	stat, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		handleError("scan path "+path, err)
	}

	if stat.IsDir() {
		contents, err := os.ReadDir(path)
		handleError("scan dir "+path, err)

		for _, content := range contents {
			srcs = append(srcs, scanSrc(filepath.Join(path, content.Name()))...)
		}
	} else if filepath.Ext(stat.Name()) == ".go" {
		srcs = append(srcs, path)
	}

	return srcs
}

// parseSrc parses configs data.
func parseSrc(path string) map[string]string {
	parser := regexp.MustCompile(`(?m)^(?:\s*)((?:[A-Z]{1})(?:[A-Za-z0-9_]*))(?:(?:\s+)(?:\={1})(?:\s+))(?:(?:\"{1})(.*)(?:\"{1}))$`)

	src, err := os.ReadFile(path)
	handleError("read source "+path, err)

	mapping := make(map[string]string)

	for _, submatches := range parser.FindAllSubmatchIndex(src, -1) {
		key := parser.ExpandString([]byte{}, "$1", string(src), submatches)
		val := parser.ExpandString([]byte{}, "$2", string(src), submatches)

		mapping[string(key)] = string(val)
	}

	return mapping
}

// getConfigs reads and parses configs from valid Go sources in internal package ver.
func getConfigs() configs {
	verPath := filepath.Join("internal", "ver")

	conf := configs{}

	srcs := scanSrc(verPath)
	for _, src := range srcs {
		mapping := parseSrc(src)

		if v, ok := mapping["Tool"]; ok {
			conf.Tool = v
		}

		if v, ok := mapping["Authors"]; ok {
			conf.Authors = v
		}

		if v, ok := mapping["Copyright"]; ok {
			conf.Copyright = v
		}
	}

	return conf
}

// stepSetupGowinres looks for go-winres and downloads and installs it if not found.
func stepSetupGoWinres(conf configs) {
	var err error
	var goPath string

	fmt.Println("look for go-winres")
	goWinresPath, err = exec.LookPath("go-winres")
	if err == nil {
		fmt.Println("go-winres found: " + goWinresPath)
		return
	} else if err != nil && !errors.Is(err, exec.ErrNotFound) {
		handleError("look for go-winres", err)
	}

	fmt.Println("look for Go")
	goPath, err = exec.LookPath("go")
	handleError("look for Go", err)

	fmt.Println("install go-winres@v" + goWinresVer)
	cmd := exec.Command(goPath, "install", goWinresImport+"@v"+goWinresVer)
	_, err = cmd.Output()
	handleError("install go-winres@v"+goWinresVer, err)

	fmt.Println("look for go-winres after install")
	goWinresPath, err = exec.LookPath("go-winres")
	if err == nil {
		fmt.Println("go-winres found: " + goWinresPath)
		return
	}
	handleError("look for go-winres after install", err)
}

// stepGenWinres generates resources for Windows binaries.
func stepGenWinres(conf configs) {
	fmt.Println("create winres dir")
	err := os.MkdirAll(winresPath, dirPerm)
	if err != nil && !os.IsExist(err) {
		handleError("create winres dir", err)
	}

	genWinresManifest(conf)

	fmt.Println("execute go-winres")
	cmd := exec.Command(goWinresPath, "make")
	_, err = cmd.Output()
	handleError("execute go-winres", err)
}

// genWinresManifest generates winres manifest.
func genWinresManifest(conf configs) {
	path := filepath.Join(winresPath, winresFileName)

	fmt.Println("generate winres parse map")
	parser := winresParseMap(conf)
	m := winresFile

	fmt.Println("parse winres manifest")
	for key, val := range parser {
		m = winresGenerator(m, key, val)
	}

	fmt.Println("write winres manifest")
	err := os.WriteFile(path, []byte(m), commonPerm)
	handleError("write winres manifest", err)
}

// winresParseMap returns mappings for the winrest manifest generator.
func winresParseMap(conf configs) map[string]string {
	parseMap := make(map[string]string)

	parseMap["info.productname"] = conf.Tool
	parseMap["info.companyname"] = conf.Authors
	parseMap["info.legalcopyright"] = conf.Copyright

	parseMap["fixed.file.version"] = conf.Version + ".0"
	parseMap["info.fileversion"] = conf.Version

	parseMap["fixed.product.version"] = conf.Version + ".0"
	parseMap["info.productversion"] = conf.Version

	return parseMap
}

// winresGenerator replaces winres manifest keys with its appropriate value.
func winresGenerator(manifest, key, val string) string {
	return strings.Replace(manifest, "${{"+key+"}}", val, -1)
}

const verFile string = `//go:build prod
// +build prod

// Generated by scripts/prerelease on %v

package ver

const (
	CopyrightYear = %d
	Version   = "%s"
)
`

// See https://github.com/tc-hib/go-winres#json-format
const winresFile string = `{
	"RT_MANIFEST": {
		"#1": {
			"0409": {
				"identity": {
					"name": "",
					"version": ""
				},
				"description": "",
				"minimum-os": "win7",
				"execution-level": "as invoker",
				"ui-access": false,
				"auto-elevate": false,
				"dpi-awareness": "system",
				"disable-theming": false,
				"disable-window-filtering": false,
				"high-resolution-scrolling-aware": false,
				"ultra-high-resolution-scrolling-aware": false,
				"long-path-aware": false,
				"printer-driver-isolation": false,
				"gdi-scaling": false,
				"segment-heap": false,
				"use-common-controls-v6": false
			}
		}
	},
	"RT_VERSION": {
		"#1": {
			"0000": {
				"fixed": {
					"file_version": "${{fixed.file.version}}",
					"product_version": "${{fixed.product.version}}"
				},
				"info": {
					"0409": {
						"Comments": "",
						"CompanyName": "${{info.companyname}}",
						"FileDescription": "",
						"FileVersion": "${{info.fileversion}}",
						"InternalName": "",
						"LegalCopyright": "${{info.legalcopyright}}",
						"LegalTrademarks": "",
						"OriginalFilename": "",
						"PrivateBuild": "",
						"ProductName": "${{info.productname}}",
						"ProductVersion": "${{info.productversion}}",
						"SpecialBuild": ""
					}
				}
			}
		}
	}
}
`
