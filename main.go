package main

import (
	"fmt"
	"os"
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

	args := os.Args
	if len(args) < 2 {
		handleExit(exit.Usage)
	}
	args = sliceShift(args)

	c, valid := validCommands[args[0]]
	if !valid {
		switch strings.ToLower(args[0]) {
		case "version", "-version", "--version":
			version()
			handleExit(exit.Norm)
		}

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
	fmt.Printf("Copyright (C) GHIFARI160 %d. Distributed under MIT License\n", ver.Copyright)
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
