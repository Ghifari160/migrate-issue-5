# Migrate

Dead simple file migration tool.

## How it works

Migrate doesn't actually move data by itself.
It's simply a wrapper over whatever copying utility you want to use.
It reads a file manifest calls the copying utility to actually move the files.
This approach means that you can use whatever copying utility you normally use (see [below](#copy-files)).

As long as the utility accepts their input in the correct format (`utility -flags source dest`), it will work with Migrate.

## Why?

I'm lazy and I wanted like to move files from _all over my drives_ to _their own destinations_ in a single command.
At the same time, I wanted to be able to preview and modify the paths before executing the command.

## How to use

Migrate can operate on files (and directories) or manifest files.
The tool can also generate the manifest file.

### Generate manifest

``` shell
migrate generate [flags] <source> <destination> [manifest]
```

By default, the manifest will be saved as `manifest.txt` in the current directory.
If `source` is a directory, the tool will scan its contents and instead add its children as entries to the manifest.
This is **NOT** done recursively, so it's up to the copying utility to handle directories.

The generate subcommand supports the following flags:

| Flag        | Type   | Default | Descriptions                                                             |
|-------------|--------|---------|--------------------------------------------------------------------------|
| `overwrite` | `bool` | `false` | Overwrite the manifest instead of appending to it.                       |
| `rel-src`   | `bool` | `true`  | Use paths relative to the manifest's path when listing the sources.      |
| `rel-dest`  | `bool` | `false` | Use paths relative tot he manifest's path when listing the destinations. |

### Copy files

You can provide a source path and a destination path,

``` shell
migrate run [flags] <source> <destination>
```

or provide a path to a manifest.

``` shell
migrate run [flags] <manifest>
```

When a manifest is provided, Migrate will read the manifest and copy files to their individual destination as listed in the manifest.

The run subcommand supports the following flags:

| Flag        | Type     | Default                                                  | Descriptions                                                                                                                                           |
|-------------|----------|----------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| `dryrun`    | `bool`   | `false`                                                  | Run in dry run mode. Simply print the execution commands without executing them.                                                                       |
| `util`      | `string` | `rsync` on Linux and macOS, and `robocopy` on Windows    | Path to copying utility or its name. Migrate will search for this utility and use it to copy files. On Windows, `.exe` will be automatically appended. |
| `util-args` | `string` | `-avr` on Linux and macOS, and `/E /COPY:DAT` on Windows | Arguments for the copying utility.                                                                                                                     |

## Building

Migrate requires [Go](https://go.dev/) 1.18 as it makes use of [generics](https://go.dev/blog/intro-generics).
There are no other dependency requirements.

### Windows

``` shell
go build -o out/migrate.exe main.go
```

### Linux

``` shell
go build -o out/migrate main.go
```

### macOS

You can build for your current architecture,

``` shell
go build -o out/migrate main.go
```

or build a universal binary file.

``` shell
GOOS=darwin GOARCH=amd64 go build -o out/migrate.amd64 main.go
GOOS=darwin GOARCH=arm64 go build -o out/migrate.arm64 main.go
lipo -create -output out/migrate out/migrate.amd64 out/migrate.arm64
```
