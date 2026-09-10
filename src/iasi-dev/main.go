// Command iasi-dev is the IASI development command-line entry point.
package main

import (
	"os"

	"iasi-dev/internal/args"
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/runners"
)

func main() {
	if len(os.Args) == 1 { printHelp() }

	command := os.Args[1]
	Parms := args.Parse(command, os.Args[2:])
	commands.SetDebug(Parms.Debug)

	if command == "help" { Parms.Help = true }
	if Parms.Help { printHelp() }

	if Parms.Message == "" { Parms.Message = command }

	logFile, err := createLogFile(command)
	if err != nil { cli.Error(RC.Error, Parms, "No se pudo crear el log: %v", err) }
	Parms.LogFile = logFile
	defer Parms.LogFile.Close()

	switch command {
	case "build":    runners.Build(&Parms)
	case "publish":  runners.Publish(&Parms)
	case "commit":   runners.Commit(Parms)
	case "release":  runners.Release(&Parms)
	case "workflow": runners.Workflow(&Parms)
	case "sync":     runners.Sync(Parms)
	default:         cli.Error(RC.InvalidArguments, Parms, "Comando desconocido: %q", command)
	}
}
