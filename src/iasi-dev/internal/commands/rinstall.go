package commands

import (
	"fmt"
	"os"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// commandRInstall builds an R package distribution using semantic source or binary modes.
func commandRInstall(directory string, logFile *os.File, args ...string) structures.Result {
	if debug { fmt.Printf("commandRInstall: directory=%s args=%v\n", directory, args) }
	if len(args) == 0 { return structures.Result{RC: RC.Error} }

	switch args[0] {
	case "source": return commandLogged(directory, true, logFile, "R", "CMD", "build", ".")
	case "binary": return commandRInstallBinary(directory, logFile)
	default:       return structures.Result{RC: RC.Error}
	}
}

// commandRInstallBinary builds a binary R package using an isolated temporary library.
func commandRInstallBinary(directory string, logFile *os.File) structures.Result {
	library, err := os.MkdirTemp("", "iasi-r-library-")
	if err != nil { return structures.Result{RC: RC.Error} }
	defer os.RemoveAll(library)

	return commandLogged(directory, true, logFile, "R", "CMD", "INSTALL", "--build", "-l", library, ".")
}
