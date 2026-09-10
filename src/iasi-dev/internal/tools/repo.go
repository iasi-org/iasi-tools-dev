package tools

import (
	"os"
	"path/filepath"
)

// IsRepo reports whether directory contains a Git marker.
func IsRepo(directory string) bool {
	_, err := os.Stat(filepath.Join(directory, ".git"))
	return err == nil
}

// IsFile reports whether path exists and is a regular filesystem entry rather than a directory.
func IsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// IsProject reports whether directory is the root of an IASI project.
func IsProject(directory string) bool {
	return IsFile(filepath.Join(directory, "_iasi.yml")) || IsFile(filepath.Join(directory, ".iasi.yml"))
}
