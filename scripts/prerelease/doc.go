/*
Prerelease prepares the environment for building release binaries of Migrate.
The script can be used interactively or non-interactively.
All interactive prompts can be overriden with their appropriate flag.

This tool uses [go-winres] (automatically downloaded) to generate resources for
Windows binaries.

Usage:

	go run ./scripts/prerelease [FLAGS]

Flags:

	-copyyear string
		Sets the copyright year.
	-ver string
		Sets the version.
	-noninteractive bool
		Disable interactive prompts.

[go-winres]: https://github.com/tc-hib/go-winres
*/
package main
