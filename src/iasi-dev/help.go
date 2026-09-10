package main

import (
	"os"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
)

func printHelp(rc ...int) {
	exitCode := RC.OK
	if len(rc) > 0 { exitCode = rc[0] }

	cli.Direct(`IASI Dev

Usage:
  iasi-dev <command> [-h] [-s] [-v|-V] [-a] [-c] [-d] [-f] [-l] [-t] [-i] [--exclude value[,value]*] [--format value] [--message value] [target...]
  iasi-dev workflow <build|publish|release> [-h] [-s] [-v|-V] [-a] [-c] [-d] [-f] [-l] [-t] [-i] [--exclude value[,value]*] [--format value] [--message value] [target...]

Commands:
  help       Show help
  build      Build
  publish    Publish
  release    Release
  commit     Commit
  workflow   Run a workflow
  sync       Sync shared files from iasi-common

Options:
  -h         Show help.
  -s         Silent output.
  -v         Verbose output.
  -V         Very verbose output.
  -a         Execute previous stages too.
  -c         Use intermediate workflows as checkpoints.
  -d         Show debug messages.
  -f         Force the operation.
  -i         Install the artifact when applicable.
  -l         Commit locally without push.
  -t         Continue when an operation fails.

Parameters:
  --exclude value[,value]*  Add exclusions. Existing files are read one exclusion per line; .git and tests are always excluded.
  --format value            Output format.
  --message value           Commit message.
`)

	os.Exit(exitCode)
}
