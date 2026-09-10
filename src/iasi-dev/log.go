package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// createLogFile creates and keeps open the log for the complete execution.
func createLogFile(command string) (*os.File, error) {
	cwd, err := os.Getwd()
	if err != nil { return nil, err }

	logDir := filepath.Join(cwd, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil { return nil, err }

	path := filepath.Join(logDir, fmt.Sprintf("iasi-%s-%s.log", command, time.Now().Format("20060102150405")))
	return os.Create(path)
}
