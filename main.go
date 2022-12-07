package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/ghifari160/migrate/cmd"
	"github.com/ghifari160/migrate/internal/exit"
	"github.com/ghifari160/migrate/internal/ver"
)

var validCommands map[string]cmd.Cmd

func main() {
	validCommands = make(map[string]cmd.Cmd)

	validCommands["run"] = cmd.NewCmdMigrate()
	validCommands["generate"] = cmd.NewCmdGenerate()

	if runtime.GOOS == "windows" {
		fmt.Println("Windows is not supported")
		os.Exit(exit.IncompatibleOS)
	}

	var ver bool
	flag.BoolVar(&ver, "version", false, "Print tool version.")
	flag.Parse()

	if ver {
		version()
		handleExit(exit.Norm)
	}

	args := flag.Args()
	if len(args) < 1 {
		handleExit(exit.Usage)
	}

	c, valid := validCommands[args[0]]
	if !valid {
		handleExit(exit.Usage)
	}
	args = sliceShift(args)

	version()

	status := c.Command(args)
	if status != exit.RDY {
		handleCmdExit(status, c)
	}

	handleCmdExit(c.Task(), c)
}

func version() {
	fmt.Printf("%s v%s\n", ver.Tool, ver.Version)
	fmt.Printf("Copyright (C) GHIFARI160 %d. Distributed under MIT License", ver.Copyright)
}

func usage() string {
	var msg strings.Builder

	for _, c := range validCommands {
		msg.WriteString(c.Usage())
	}

	return msg.String()
}

func sliceShift[T any](s []T) []T {
	if len(s) < 2 {
		return make([]T, 0)
	}

	return s[1:]
}

func handleExit(exitCode int) {
	msg := exit.Message(exitCode)

	if exitCode == exit.Usage {
		msg += "\n" + usage()
	}

	fmt.Println(msg)
	os.Exit(exitCode)
}

func handleCmdExit(exitCode int, cmd cmd.Cmd) {
	msg := exit.Message(exitCode)

	if exitCode == exit.Usage {
		msg += "\n" + cmd.Usage() + "\n"
	}

	fmt.Println(msg)
	os.Exit(exitCode)
}
