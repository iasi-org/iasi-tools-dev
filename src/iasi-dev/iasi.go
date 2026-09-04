package main

import (
	"os"
	"path/filepath"

	"iasi-dev/internal/cli"

	"gopkg.in/yaml.v3"
)

// IASI contains the processed configuration of an IASI marker.
type IASI struct {
	Type       string `yaml:"type"`
	ReleaseDir string `yaml:"release-dir"`
}

// readIASI reads and processes the IASI marker of a directory.
func readIASI(directory string) (IASI, bool) {
	path := iasiMarkerPath(directory)
	if path == "" {
		cli.Warning("Se ignora %s: no tiene marcador IASI.", directory)
		return IASI{}, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		cli.Warning("Se ignora %s: no se puede leer %s.", directory, filepath.Base(path))
		return IASI{}, false
	}

	iasi := IASI{}
	if err := yaml.Unmarshal(data, &iasi); err != nil {
		cli.Warning("Se ignora %s: %s no contiene YAML válido.", directory, filepath.Base(path))
		return IASI{}, false
	}

	return iasi, true
}

// iasiMarkerPath returns the IASI marker found in the root of a directory.
func iasiMarkerPath(directory string) string {
	for _, name := range []string{"_iasi.yml", ".iasi.yml"} {
		path := filepath.Join(directory, name)
		if fileExists(path) {
			return path
		}
	}

	return ""
}
