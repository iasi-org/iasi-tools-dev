package commands

import (
	"fmt"
	"os"
	"strings"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// commandGit executes Git using friendly semantics when requested.
func commandGit(directory string, friendly bool, logFile *os.File, args ...string) structures.Result {
	if debug { fmt.Printf("commandGit: directory=%s friendly=%t args=%v\n", directory, friendly, args) }
	if friendly { return commandGitFriendly(directory, logFile, args...) }

	return command(directory, false, logFile, "git", args...)
}

// commandGitFriendly executes Git with command-specific friendly semantics when available.
func commandGitFriendly(directory string, logFile *os.File, args ...string) structures.Result {
	if debug { fmt.Printf("commandGitFriendly: directory=%s args=%v\n", directory, args) }
	if len(args) == 0 { return command(directory, true, logFile, "git") }

	switch args[0] {
	case "status": return commandGitStatus(directory, logFile, args[1:]...)
	default:       return command(directory, true, logFile, "git", args...)
	}
}

// commandGitStatus checks whether the repository has pending changes.
func commandGitStatus(directory string, logFile *os.File, args ...string) structures.Result {
	if debug { fmt.Printf("commandGitStatus: directory=%s args=%v\n", directory, args) }
	result := command(directory, true, logFile, "git", "status", "--porcelain")

	if result.RC != RC.OK {
		result.RC = RC.Error
		return result
	}

	if strings.TrimSpace(result.Stdout) == "" { result.RC = RC.NothingToDo }

	return result
}
