package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Run executes a command using its specific implementation when available.
func Run(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug { fmt.Printf("Run: directory=%s name=%s args=%v\n", directory, name, args) }
	switch name {
	case "git": return commandGit(directory, false, logFile, args...)
	default:    return command(directory, false, logFile, name, args...)
	}
}

// RunFriendly executes a command producing friendly output when supported.
func RunFriendly(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug { fmt.Printf("RunFriendly: directory=%s name=%s args=%v\n", directory, name, args) }
	switch name {
	case "git": return commandGit(directory, true, logFile, args...)
	default:    return command(directory, true, logFile, name, args...)
	}
}

// RunDirect executes a command without using a specific implementation.
func RunDirect(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug { fmt.Printf("RunDirect: directory=%s name=%s args=%v\n", directory, name, args) }
	return command(directory, false, logFile, name, args...)
}

// RunDirectLogged executes a direct command and writes its output to the log.
func RunDirectLogged(directory string, logFile *os.File, name string, args ...string) structures.Result {
	if debug { fmt.Printf("RunDirectLogged: directory=%s name=%s args=%v\n", directory, name, args) }
	return commandLogged(directory, false, logFile, name, args...)
}

// command executes a generic command, logs it and captures its output.
func command(directory string, friendly bool, logFile *os.File, name string, args ...string) structures.Result {
	if debug { fmt.Printf("command: directory=%s friendly=%t name=%s args=%v\n", directory, friendly, name, args) }
	writeCommand(logFile, directory, name, args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := structures.Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		RC:     RC.OK,
	}

	if err != nil { result.RC = RC.Error }

	return result
}

// commandLogged executes a generic command, logs it and writes its output directly to the log.
func commandLogged(directory string, friendly bool, logFile *os.File, name string, args ...string) structures.Result {
	if debug { fmt.Printf("commandLogged: directory=%s friendly=%t name=%s args=%v\n", directory, friendly, name, args) }
	writeCommand(logFile, directory, name, args...)

	cmd := exec.Command(name, args...)
	cmd.Dir = directory
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err := cmd.Run()

	result := structures.Result{RC: RC.OK}
	if err != nil { result.RC = RC.Error }

	return result
}

// writeCommand writes the executed command to the log.
func writeCommand(logFile *os.File, directory string, name string, args ...string) {
	if logFile == nil { return }
	fmt.Fprintf(logFile, "%s > %s %s\n", directory, name, strings.Join(args, " "))
}
