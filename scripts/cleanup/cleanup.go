package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var latestStep string

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s ERROR: %v\n", latestStep, r)
			os.Exit(1)
		}
	}()

	fmt.Println("cleanup")
	fmt.Println()

	stepRmGeneratedFiles()
	stepRmWinresDir()
	stepRmSysos()
	stepRmBuildDir()

	fmt.Println()
}

// handleError automatically raises panics when it encounters an error.
func handleError(step string, err error) {
	if err != nil {
		latestStep = step
		panic(err)
	}
}

// removePattern removes all files matching the pattern.
func removePattern(pattern string) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, f := range files {
		err = os.RemoveAll(f)
	}

	return err
}

// stepRmGeneratedFiles removes generated files.
func stepRmGeneratedFiles() {
	fmt.Println("rm generated files")
	err := removePattern("*/***/*_generated.*")
	handleError("rm generated file", err)
}

// stepRmWinresDir removes winres config dir.
func stepRmWinresDir() {
	fmt.Println("rm winres dir")
	err := os.RemoveAll("winres")
	if err != nil && !os.IsNotExist(err) {
		handleError("rm winres dir", err)
	}
}

// stepRmSysos removes *.syso files.
func stepRmSysos() {
	fmt.Println("rm sysos")
	err := removePattern("*/***/*.syso")
	handleError("rm sysos", err)
}

// stepRmBuildDir removes the build directory.
func stepRmBuildDir() {
	fmt.Println("rm build dir")
	err := os.RemoveAll("out")
	if err != nil && !os.IsNotExist(err) {
		handleError("rm build dir", err)
	}
}