package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"iasi-dev/internal/cli"
)

type releasePublish struct {
	Source      string
	ArtifactDir string
	Destination string
}

// releaseRepository promotes one repository according to the type declared
// by the IASI marker in the repository root.
func releaseRepository(repository string, arguments Arguments) error {
	iasi, ok := readIASI(repository)
	if !ok {
		return fmt.Errorf("no se pudo leer la configuración IASI raíz de %s", repository)
	}

	switch iasi.Type {
	case "book":
		return releaseBook(repository, arguments, iasi)
	default:
		return fmt.Errorf("type no soportado para release en %s: %q", repository, iasi.Type)
	}
}

// releaseBook promotes every _publish found in a book repository.
// The root artifact is copied to the release root. Nested artifacts preserve
// their artifact directory path, removing leading numeric prefixes from each
// path component.
func releaseBook(repository string, arguments Arguments, iasi IASI) error {
	releaseDir, err := resolveReleaseDirectory(repository, iasi.ReleaseDir)
	if err != nil {
		return err
	}

	publishes, err := findReleasePublishes(repository, releaseDir)
	if err != nil {
		return fmt.Errorf("no se pudieron localizar _publish en %s: %w", repository, err)
	}
	if len(publishes) == 0 {
		return fmt.Errorf("no se encontró ningún _publish en %s", repository)
	}

	workDir := releaseDir + ".work"
	if err := os.RemoveAll(workDir); err != nil {
		return fmt.Errorf("no se pudo limpiar %s: %w", workDir, err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("no se pudo crear %s: %w", workDir, err)
	}

	completed := false
	defer func() {
		if !completed {
			_ = os.RemoveAll(workDir)
		}
	}()

	// Root publish first. This also lets us detect a root-published directory
	// colliding with a named nested artifact before overwriting anything.
	sort.Slice(publishes, func(i, j int) bool {
		if publishes[i].Destination == "" {
			return true
		}
		if publishes[j].Destination == "" {
			return false
		}
		return publishes[i].Destination < publishes[j].Destination
	})

	seenDestinations := map[string]string{}
	for _, publish := range publishes {
		destination := workDir
		if publish.Destination != "" {
			destination = filepath.Join(workDir, publish.Destination)
		}

		key := strings.ToLower(filepath.Clean(destination))
		if previous, exists := seenDestinations[key]; exists {
			return fmt.Errorf("%s y %s producirían el mismo destino de release: %s", previous, publish.Source, destination)
		}
		seenDestinations[key] = publish.Source

		if publish.Destination != "" {
			if _, err := os.Stat(destination); err == nil {
				return fmt.Errorf("el destino %s ya existe al incorporar %s", destination, publish.Source)
			} else if !os.IsNotExist(err) {
				return err
			}
		}

		cli.Verbose("\t%s -> %s", publish.Source, filepath.Clean(destination))
		if err := copyDirectoryContents(publish.Source, destination); err != nil {
			return fmt.Errorf("no se pudo copiar %s: %w", publish.Source, err)
		}
	}

	if err := os.RemoveAll(releaseDir); err != nil {
		return fmt.Errorf("no se pudo reemplazar %s: %w", releaseDir, err)
	}
	if err := os.Rename(workDir, releaseDir); err != nil {
		return fmt.Errorf("no se pudo materializar %s: %w", releaseDir, err)
	}

	completed = true
	cli.VeryVerbose("\t%s: release completada en %s.", filepath.Base(repository), releaseDir)
	return nil
}

// resolveReleaseDirectory returns the configured release directory or
// <project>/release by default.
func resolveReleaseDirectory(repository string, configured string) (string, error) {
	configured = strings.TrimSpace(configured)
	if configured == "" {
		configured = "release"
	}
	if filepath.IsAbs(configured) {
		return "", fmt.Errorf("release-dir debe ser relativo al proyecto: %s", configured)
	}

	releaseDir := filepath.Clean(filepath.Join(repository, configured))
	inside, err := pathInside(repository, releaseDir)
	if err != nil || !inside || releaseDir == filepath.Clean(repository) {
		return "", fmt.Errorf("release-dir debe estar dentro del proyecto y no puede ser su raíz: %s", configured)
	}

	return releaseDir, nil
}

// findReleasePublishes finds every _publish in a repository and resolves
// where each one belongs in the release.
func findReleasePublishes(repository string, releaseDir string) ([]releasePublish, error) {
	publishes := []releasePublish{}
	workDir := releaseDir + ".work"

	err := filepath.WalkDir(repository, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}

		cleanPath := filepath.Clean(path)
		if cleanPath == filepath.Clean(releaseDir) || cleanPath == filepath.Clean(workDir) {
			return filepath.SkipDir
		}
		if path != repository && ignoredReleaseDirectory(entry.Name()) {
			return filepath.SkipDir
		}
		if entry.Name() != "_publish" {
			return nil
		}

		artifactDir := owningArtifactDirectory(repository, filepath.Dir(path))
		destination := ""
		if filepath.Clean(artifactDir) != filepath.Clean(repository) {
			relative, err := filepath.Rel(repository, artifactDir)
			if err != nil {
				return err
			}
			destination = normalizeReleaseArtifactPath(relative)
		}

		publishes = append(publishes, releasePublish{
			Source:      cleanPath,
			ArtifactDir: artifactDir,
			Destination: destination,
		})
		return filepath.SkipDir
	})

	return publishes, err
}

// owningArtifactDirectory returns the nearest ancestor with an IASI marker.
// Nested markers identify artifact ownership only. Release configuration and
// the release strategy always come from the repository root marker.
func owningArtifactDirectory(repository string, start string) string {
	repository = filepath.Clean(repository)
	current := filepath.Clean(start)

	for {
		if hasIASIMarker(current) {
			return current
		}
		if current == repository {
			return repository
		}

		parent := filepath.Dir(current)
		inside, err := pathInside(repository, parent)
		if err != nil || !inside || parent == current {
			return repository
		}
		current = parent
	}
}

// normalizeReleaseArtifactPath keeps the artifact directory structure while
// removing leading numeric prefixes from every directory name.
func normalizeReleaseArtifactPath(path string) string {
	parts := splitPath(path)
	for i, part := range parts {
		parts[i] = stripNumericPrefix(part)
	}
	return filepath.Join(parts...)
}

// stripNumericPrefix removes leading digits and the separators that follow
// them. Examples: 01-user-guide -> user-guide, 002_api -> api.
func stripNumericPrefix(name string) string {
	index := 0
	for index < len(name) && name[index] >= '0' && name[index] <= '9' {
		index++
	}
	if index == 0 {
		return name
	}

	for index < len(name) {
		switch name[index] {
		case '-', '_', '.', ' ':
			index++
		default:
			if index == len(name) {
				return name
			}
			return name[index:]
		}
	}

	// A directory made only of a numeric prefix has no useful normalized
	// identity, so keep its original name rather than producing an empty path.
	return name
}

func ignoredReleaseDirectory(name string) bool {
	return name == ".git" || name == "tests"
}

// copyDirectoryContents copies the complete contents of source into destination
// preserving the internal directory structure.
func copyDirectoryContents(source string, destination string) error {
	if err := os.MkdirAll(destination, 0755); err != nil {
		return err
	}

	return filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == source {
			return nil
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, target)
		}
		return copyFile(path, target)
	})
}
