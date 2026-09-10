package commands

import (
	"fmt"
	"os"

	"iasi-dev/internal/structures"
)

// RunLogged executes a command writing its output directly to the log.
func RunLogged(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug { fmt.Printf("RunLogged: directory=%s name=%s args=%v\n", directory, name, args) }
	return commandLogged(directory, false, logFile, name, args...)
}

// RunFriendlyLogged executes a friendly command writing its output directly to the log.
func RunFriendlyLogged(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug { fmt.Printf("RunFriendlyLogged: directory=%s name=%s args=%v\n", directory, name, args) }
	switch name {
	case "rinstall": return commandRInstall(directory, logFile, args...)
	default:         return commandLogged(directory, true, logFile, name, args...)
	}
}
