package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"iasi-dev/internal/RC"
	"iasi-dev/internal/cli"
)

// runSync propagates entries from iasi-common to copies that already exist in the workspace.
func runSync(arguments Arguments, entries []string) {
	if len(entries) == 0 {
		cli.Error("Debes indicar al menos un archivo o directorio.")
		os.Exit(RC.InvalidArguments)
	}

	workspace, err := os.Getwd()
	if err != nil {
		cli.Error("No se pudo resolver el workspace: %v", err)
		os.Exit(RC.Sync)
	}

	common := os.Getenv("IASI_COMMON_DIR")
	if common == "" {
		common = filepath.Join(workspace, "iasi-common")
	} else if !filepath.IsAbs(common) {
		common = filepath.Join(workspace, common)
	}
	common = filepath.Clean(common)

	info, err := os.Stat(common)
	if err != nil || !info.IsDir() {
		cli.Error("No se encontró iasi-common: %s", common)
		os.Exit(RC.Sync)
	}

	synced := 0

	for _, entry := range entries {
		count, err := syncEntry(workspace, common, entry, arguments)
		if err == nil {
			synced += count
			continue
		}

		if arguments.Tolerant {
			cli.Warning("%v", err)
			continue
		}

		cli.Error("%v", err)
		cli.Error("Consulta el log: %s", arguments.LogFile)
		os.Exit(RC.Sync)
	}

	cli.Success("%d copia(s) sincronizada(s) desde iasi-common.", synced)
}

// syncEntry resolves one canonical source and updates every existing copy outside iasi-common.
func syncEntry(workspace string, common string, entry string, arguments Arguments) (int, error) {
	source, pathEntry, err := resolveSyncSource(common, entry)
	if err != nil {
		return 0, err
	}

	displayName := filepath.Base(source)
	if pathEntry {
		displayName = filepath.Clean(entry)
	}

	cli.Info("Sincronizando %s", displayName)

	targets, err := findSyncTargets(workspace, common, source, entry, pathEntry)
	if err != nil {
		return 0, err
	}

	if len(targets) == 0 {
		cli.Warning("No existen copias de %s fuera de iasi-common.", displayName)
		return 0, nil
	}

	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return 0, fmt.Errorf("no se pudo leer %s: %w", source, err)
	}

	synced := 0

	for _, target := range targets {
		cli.Verbose("\tSincronizando %s", target)

		if sourceInfo.IsDir() {
			err = mergeDirectory(source, target)
		} else {
			err = copyFile(source, target)
		}
		if err != nil {
			syncErr := fmt.Errorf("no se pudo sincronizar %s: %w", target, err)
			if arguments.Tolerant {
				cli.Warning("%v", syncErr)
				continue
			}
			return synced, syncErr
		}

		synced++
		cli.VeryVerbose("\t%s: sync completado.", target)
	}

	return synced, nil
}

// resolveSyncSource resolves a name or a path relative to iasi-common to one canonical source.
func resolveSyncSource(common string, entry string) (string, bool, error) {
	cleanEntry := filepath.Clean(entry)
	pathEntry := strings.ContainsAny(entry, `/\\`)

	if pathEntry {
		if filepath.IsAbs(cleanEntry) {
			return "", true, fmt.Errorf("la ruta debe ser relativa a iasi-common: %s", entry)
		}

		source := filepath.Clean(filepath.Join(common, cleanEntry))
		inside, err := pathInside(common, source)
		if err != nil || !inside {
			return "", true, fmt.Errorf("la ruta debe ser relativa a iasi-common: %s", entry)
		}

		if _, err := os.Lstat(source); err != nil {
			return "", true, fmt.Errorf("no se encontró %s en iasi-common", entry)
		}

		return source, true, nil
	}

	matches := []string{}
	err := filepath.WalkDir(common, func(path string, item os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path != common && item.IsDir() && ignoredSyncDirectory(item.Name()) {
			return filepath.SkipDir
		}

		if path != common && item.Name() == cleanEntry {
			matches = append(matches, path)
		}

		return nil
	})
	if err != nil {
		return "", false, err
	}

	if len(matches) == 0 {
		return "", false, fmt.Errorf("no se encontró %s en iasi-common", cleanEntry)
	}
	if len(matches) > 1 {
		return "", false, fmt.Errorf("hay más de una entrada llamada %s en iasi-common", cleanEntry)
	}

	return matches[0], false, nil
}

// findSyncTargets finds only copies that already exist outside iasi-common.
func findSyncTargets(workspace string, common string, source string, entry string, pathEntry bool) ([]string, error) {
	sourceInfo, err := os.Lstat(source)
	if err != nil {
		return nil, err
	}

	targets := []string{}
	cleanEntry := filepath.Clean(entry)
	entryName := filepath.Base(source)

	err = filepath.WalkDir(workspace, func(path string, item os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if path == common {
			return filepath.SkipDir
		}
		if path != workspace && item.IsDir() && ignoredSyncDirectory(item.Name()) {
			return filepath.SkipDir
		}
		if path == workspace {
			return nil
		}

		if sourceInfo.IsDir() != item.IsDir() {
			return nil
		}

		matched := item.Name() == entryName
		if pathEntry {
			matched = pathEndsWith(path, cleanEntry)
		}
		if !matched {
			return nil
		}

		targets = append(targets, path)
		if item.IsDir() {
			return filepath.SkipDir
		}

		return nil
	})

	return targets, err
}

// ignoredSyncDirectory reports directories that sync must never traverse.
func ignoredSyncDirectory(name string) bool {
	return name == ".git" || name == "tests"
}

// pathEndsWith reports whether path ends with the same path components as suffix.
func pathEndsWith(path string, suffix string) bool {
	path = filepath.Clean(path)
	suffix = filepath.Clean(suffix)

	if filepath.Base(path) != filepath.Base(suffix) {
		return false
	}

	pathParts := splitPath(path)
	suffixParts := splitPath(suffix)
	if len(suffixParts) > len(pathParts) {
		return false
	}

	start := len(pathParts) - len(suffixParts)
	for i := range suffixParts {
		if !strings.EqualFold(pathParts[start+i], suffixParts[i]) {
			return false
		}
	}

	return true
}

// splitPath splits a cleaned path into platform-independent path components.
func splitPath(path string) []string {
	path = filepath.ToSlash(filepath.Clean(path))
	parts := strings.Split(path, "/")
	result := []string{}

	for _, part := range parts {
		if part != "" && part != "." {
			result = append(result, part)
		}
	}

	return result
}

// pathInside reports whether path is common itself or a descendant of common.
func pathInside(common string, path string) (bool, error) {
	relative, err := filepath.Rel(common, path)
	if err != nil {
		return false, err
	}

	return relative == "." || (relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))), nil
}

// mergeDirectory copies source into an existing target without deleting target-only entries.
func mergeDirectory(source string, target string) error {
	return filepath.WalkDir(source, func(path string, item os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, relative)

		info, err := item.Info()
		if err != nil {
			return err
		}

		if item.IsDir() {
			return os.MkdirAll(destination, info.Mode().Perm())
		}

		if info.Mode()&os.ModeSymlink != 0 {
			return copySymlink(path, destination)
		}

		return copyFile(path, destination)
	})
}

// copyFile replaces one file with the canonical source contents.
func copyFile(source string, target string) error {
	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

// copySymlink replaces one symbolic link while preserving its link target.
func copySymlink(source string, target string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}

	if err := os.RemoveAll(target); err != nil {
		return err
	}

	return os.Symlink(link, target)
}
