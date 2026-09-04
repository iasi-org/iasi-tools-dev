// Command iasi-dev is the IASI development command-line entry point.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"iasi-dev/internal/RC"
	"iasi-dev/internal/cli"
)

type Arguments struct {
	Verbose  int
	Full     bool
	Tolerant bool
	Install  bool
	Message  string
	Format   string
	LogFile  string
	Params   map[string]string
}

func main() {
	if len(os.Args) == 1 {
		printHelp()
		os.Exit(RC.OK)
	}

	command := os.Args[1]
	arguments, positionalArgs := parseArguments(command, os.Args[2:])

	logFile, err := createLogFile(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No se pudo crear el log: %v\n", err)
		os.Exit(RC.Error)
	}

	arguments.LogFile = logFile
	cli.ConfigureOutput(cli.OutputOptions{Verbose: arguments.Verbose, LogFile: arguments.LogFile})
	if command == "sync" {
		runSync(arguments, positionalArgs)
		os.Exit(RC.OK)
	}

	directories := selectDirectories(positionalArgs)
	directories = selectIASIDirectories(directories)

	runCommand(command, arguments, directories)
	os.Exit(RC.OK)
}

// createLogFile creates the log for the complete iasi-dev execution.
func createLogFile(command string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	logDir := filepath.Join(cwd, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", err
	}

	logFile := filepath.Join(logDir, fmt.Sprintf("iasi-%s-%s.log", command, time.Now().Format("20060102150405")))
	file, err := os.Create(logFile)
	if err != nil {
		return "", err
	}

	if err := file.Close(); err != nil {
		return "", err
	}

	return logFile, nil
}

// parseArguments converts command-line options into the common arguments structure.
func parseArguments(command string, args []string) (Arguments, []string) {
	arguments := Arguments{Verbose: 1, Message: "iasi-dev " + command, Params: map[string]string{}}
	directories := []string{}
	silent := false

	for len(args) > 0 {
		argument := args[0]

		switch argument {
		case "-s":
			silent = true
			args = args[1:]
		case "-v":
			if arguments.Verbose < 3 {
				arguments.Verbose = 3
			}
			args = args[1:]
		case "-V":
			arguments.Verbose = 7
			args = args[1:]
		case "-f":
			arguments.Full = true
			args = args[1:]
		case "-t":
			arguments.Tolerant = true
			args = args[1:]
		case "-i":
			arguments.Install = true
			args = args[1:]
		default:
			if strings.HasPrefix(argument, "--") {
				key := strings.TrimPrefix(argument, "--")
				value := ""

				if len(args) > 1 {
					value = args[1]
					args = args[2:]
				} else {
					args = args[1:]
				}

				if key == "message" {
					arguments.Message = value
					continue
				}

				if key == "format" {
					arguments.Format = value
					continue
				}

				arguments.Params[key] = value
				continue
			}

			directories = append(directories, argument)
			args = args[1:]
		}
	}

	if silent {
		arguments.Verbose = 0
	}

	return arguments, directories
}

// selectDirectories resolves the directories that will be processed.
// Without explicit directories, it returns the direct children of the current directory.
func selectDirectories(candidates []string) []string {
	if len(candidates) == 0 {
		return firstLevelDirectories(".")
	}

	directories := []string{}

	for _, candidate := range candidates {
		if filepath.Clean(candidate) == "." {
			directories = append(directories, firstLevelDirectories(candidate)...)
			continue
		}

		directory, err := filepath.Abs(candidate)
		if err != nil {
			cli.Warning("Se ignora %q: no se puede resolver la ruta.", candidate)
			continue
		}

		if isIgnoredDirectory(directory) {
			continue
		}

		info, err := os.Stat(directory)
		if err != nil || !info.IsDir() {
			cli.Warning("Se ignora %q: no existe o no es un directorio.", candidate)
			continue
		}

		directories = append(directories, filepath.Clean(directory))
	}

	return directories
}

// firstLevelDirectories returns only the direct child directories of root.
func firstLevelDirectories(root string) []string {
	root, err := filepath.Abs(root)
	if err != nil {
		cli.Warning("Se ignora %q: no se puede resolver la ruta.", root)
		return nil
	}

	if isIgnoredDirectory(root) {
		return nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		cli.Warning("Se ignora %q: no existe o no se puede leer.", root)
		return nil
	}

	directories := []string{}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "tests" {
			continue
		}

		directories = append(directories, filepath.Join(root, entry.Name()))
	}

	return directories
}

// isIgnoredDirectory reports whether a path is inside a tests directory.
func isIgnoredDirectory(path string) bool {
	path = filepath.Clean(path)

	for {
		if filepath.Base(path) == "tests" {
			return true
		}

		parent := filepath.Dir(path)
		if parent == path {
			return false
		}

		path = parent
	}
}

// selectIASIDirectories keeps only directories marked as IASI repositories.
func selectIASIDirectories(directories []string) []string {
	selected := []string{}

	for _, directory := range directories {
		if !hasIASIMarker(directory) {
			continue
		}

		selected = append(selected, directory)
	}

	return selected
}

// hasIASIMarker reports whether a directory contains an IASI marker in its root.
func hasIASIMarker(directory string) bool {
	return iasiMarkerPath(directory) != ""
}

// fileExists reports whether path exists and is a regular filesystem entry rather than a directory.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// isGitRepository reports whether the directory is a Git repository.
func isGitRepository(directory string) bool {
	_, err := os.Stat(filepath.Join(directory, ".git"))
	return err == nil
}

// runCommand dispatches the selected command.
func runCommand(command string, arguments Arguments, directories []string) {
	switch command {
	case "help":
		printHelp()
	case "build":
		runBuild(arguments, directories)
	case "publish":
		runPublish(arguments, directories)
	case "commit":
		runCommit(arguments, directories)
	case "release":
		runRelease(arguments, directories)
	case "deploy":
		runDeploy(arguments, directories)
	default:
		cli.Error("Comando desconocido: %q", command)
		os.Exit(RC.InvalidArguments)
	}
}

// findArtifacts returns the IASI artifacts declared below a repository.
// Directories named tests are never traversed.
func findArtifacts(repository string) []string {
	artifacts := []string{}

	filepath.WalkDir(repository, func(path string, entry os.DirEntry, err error) error {
		if err != nil || !entry.IsDir() {
			return nil
		}

		if entry.Name() == "tests" {
			return filepath.SkipDir
		}

		if path == repository {
			return nil
		}

		if hasIASIMarker(path) {
			artifacts = append(artifacts, path)
		}

		return nil
	})

	return artifacts
}

// runBuild builds the selected repositories and returns those completed successfully.
func runBuild(arguments Arguments, repositories []string) []string {
	built := []string{}

	for _, repository := range repositories {
		if !isGitRepository(repository) {
			continue
		}

		cli.Info("Construyendo %s", filepath.Base(repository))
		repositoryBuilt := true
		artifacts := findArtifacts(repository)

		for _, artifact := range artifacts {
			err := buildArtifact(artifact, arguments)
			if err == nil {
				continue
			}

			repositoryBuilt = false
			if arguments.Tolerant {
				cli.Warning("%v", err)
				continue
			}

			cli.Error("%v", err)
			cli.Error("Consulta el log: %s", arguments.LogFile)
			os.Exit(RC.Build)
		}

		if repositoryBuilt {
			built = append(built, repository)
		}
	}

	return built
}

// buildArtifact builds one artifact according to its declared IASI type.
func buildArtifact(artifact string, arguments Arguments) error {
	iasi, ok := readIASI(artifact)
	if !ok {
		return nil
	}

	cli.Verbose("\tConstruyendo %s", filepath.Base(artifact))

	var err error
	switch iasi.Type {
	case "r-package":
		err = buildPackage(artifact, arguments, iasi)
	case "quarto":
		err = buildQuarto(artifact, arguments, iasi)
	default:
		cli.Warning("Se ignora %s: type ausente o no soportado: %q.", artifact, iasi.Type)
		return nil
	}

	if err != nil {
		return err
	}

	cli.VeryVerbose("\t%s: build completado.", filepath.Base(artifact))
	return nil
}

// buildPackage builds the source and binary distributions of an R package artifact.
func buildPackage(artifact string, arguments Arguments, iasi IASI) error {
	if err := buildSourcePackage(artifact, arguments); err != nil {
		return err
	}

	if err := buildBinaryPackage(artifact, arguments); err != nil {
		return err
	}

	if arguments.Install {
		return installPackage(artifact, arguments)
	}

	return nil
}

// buildSourcePackage builds the source distribution of an R package.
func buildSourcePackage(artifact string, arguments Arguments) error {
	if err := runLoggedCommand(artifact, arguments, "R", "CMD", "build", "."); err != nil {
		return fmt.Errorf("no se pudo construir el paquete fuente %s: %w", artifact, err)
	}

	return nil
}

// buildBinaryPackage builds the binary distribution without installing it in the user library.
func buildBinaryPackage(artifact string, arguments Arguments) error {
	library, err := os.MkdirTemp("", "iasi-r-library-")
	if err != nil {
		return fmt.Errorf("no se pudo crear la librería temporal para %s: %w", artifact, err)
	}
	defer os.RemoveAll(library)

	if err := runLoggedCommand(artifact, arguments, "R", "CMD", "INSTALL", "--build", "-l", library, "."); err != nil {
		return fmt.Errorf("no se pudo construir el paquete binario %s: %w", artifact, err)
	}

	return nil
}

// installPackage installs an R package when explicitly requested with -i.
func installPackage(artifact string, arguments Arguments) error {
	if err := runLoggedCommand(artifact, arguments, "R", "CMD", "INSTALL", "."); err != nil {
		return fmt.Errorf("no se pudo instalar el paquete %s: %w", artifact, err)
	}

	return nil
}

// buildQuarto builds a Quarto artifact with iasi.quarto.
func buildQuarto(artifact string, arguments Arguments, iasi IASI) error {
	expression := quartoBuildExpression(arguments)
	if err := runLoggedCommand(artifact, arguments, "Rscript", "-e", expression); err != nil {
		return fmt.Errorf("no se pudo construir %s: %w", artifact, err)
	}

	return nil
}

// runLoggedCommand executes a command in a directory and sends all output to the execution log.
func runLoggedCommand(directory string, arguments Arguments, name string, args ...string) error {
	command := exec.Command(name, args...)
	command.Dir = directory

	log, err := os.OpenFile(arguments.LogFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("no se puede abrir el log %s: %w", arguments.LogFile, err)
	}
	defer log.Close()

	command.Stdout = log
	command.Stderr = log

	return command.Run()
}

// quartoBuildExpression builds the R expression for iasi.quarto::build().
func quartoBuildExpression(arguments Arguments) string {
	parameters := []string{}

	if arguments.Format != "" {
		parameters = append(parameters, "format = "+strconv.Quote(arguments.Format))
	}

	if arguments.Full {
		parameters = append(parameters, "force = TRUE")
	}

	return "iasi.quarto::build(" + strings.Join(parameters, ", ") + ")"
}

// runPublish publishes the selected repositories and returns those completed successfully.
func runPublish(arguments Arguments, repositories []string) []string {
	if arguments.Full {
		repositories = runBuild(arguments, repositories)
	}

	published := []string{}

	for _, repository := range repositories {
		if !isGitRepository(repository) {
			continue
		}

		cli.Info("Publicando %s", filepath.Base(repository))
		repositoryPublished := true
		artifacts := findArtifacts(repository)

		for _, artifact := range artifacts {
			err := publishArtifact(artifact, arguments)
			if err == nil {
				continue
			}

			repositoryPublished = false
			if arguments.Tolerant {
				cli.Warning("%v", err)
				continue
			}

			cli.Error("%v", err)
			cli.Error("Consulta el log: %s", arguments.LogFile)
			os.Exit(RC.Publish)
		}

		if repositoryPublished {
			published = append(published, repository)
		}
	}

	return published
}

// publishArtifact publishes one artifact when its declared IASI type is publishable.
func publishArtifact(artifact string, arguments Arguments) error {
	iasi, ok := readIASI(artifact)
	if !ok {
		return nil
	}

	switch iasi.Type {
	case "quarto":
		return publishQuarto(artifact, arguments, iasi)
	case "book":
		return publishQuarto(artifact, arguments, iasi)
	default:
		return nil
	}
}

// publishQuarto publishes a Quarto artifact with iasi.quarto.
func publishQuarto(artifact string, arguments Arguments, iasi IASI) error {
	cli.Verbose("\tPublicando %s", filepath.Base(artifact))

	expression := quartoPublishExpression(arguments)
	if err := runLoggedCommand(artifact, arguments, "Rscript", "-e", expression); err != nil {
		return fmt.Errorf("no se pudo publicar %s: %w", artifact, err)
	}

	cli.VeryVerbose("\t%s: publish completado.", filepath.Base(artifact))
	return nil
}

// quartoPublishExpression builds the R expression for iasi.quarto::publish().
func quartoPublishExpression(arguments Arguments) string {
	parameters := []string{}

	if arguments.Full {
		parameters = append(parameters, "force = TRUE")
	}

	return "iasi.quarto::publish(" + strings.Join(parameters, ", ") + ")"
}

// runCommit commits and pushes the selected repositories and returns those completed successfully.
func runCommit(arguments Arguments, repositories []string) []string {
	if arguments.Full {
		repositories = runPublish(arguments, repositories)
	}

	committed := []string{}

	for _, repository := range repositories {
		if !isGitRepository(repository) {
			continue
		}

		cli.Info("Commit %s", filepath.Base(repository))
		err := commitRepository(repository, arguments)
		if err == nil {
			committed = append(committed, repository)
			continue
		}

		if arguments.Tolerant {
			cli.Warning("%v", err)
			continue
		}

		cli.Error("%v", err)
		cli.Error("Consulta el log: %s", arguments.LogFile)
		os.Exit(RC.Commit)
	}

	return committed
}

// commitRepository stages, commits and pushes one Git repository.
func commitRepository(repository string, arguments Arguments) error {
	if err := runLoggedCommand(repository, arguments, "git", "add", "-A", "--", "."); err != nil {
		return fmt.Errorf("no se pudieron preparar los cambios de %s: %w", repository, err)
	}

	hasChanges, err := gitHasChanges(repository, arguments)
	if err != nil {
		return err
	}

	if hasChanges {
		if err := runLoggedCommand(repository, arguments, "git", "commit", "-m", arguments.Message); err != nil {
			return fmt.Errorf("no se pudo crear el commit de %s: %w", repository, err)
		}
	}

	hasOutgoing, err := gitHasOutgoing(repository, arguments)
	if err != nil {
		return err
	}

	if hasOutgoing {
		if err := runLoggedCommand(repository, arguments, "git", "push"); err != nil {
			return fmt.Errorf("no se pudo hacer push de %s: %w", repository, err)
		}
	}

	cli.VeryVerbose("\t%s: commit completado.", filepath.Base(repository))
	return nil
}

// gitHasChanges reports whether a repository has staged or unstaged changes.
func gitHasChanges(repository string, arguments Arguments) (bool, error) {
	output, err := runLoggedOutput(repository, arguments, "git", "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("no se pudo consultar el estado de %s: %w", repository, err)
	}

	return strings.TrimSpace(output) != "", nil
}

// gitHasOutgoing reports whether HEAD contains commits not present in its upstream branch.
func gitHasOutgoing(repository string, arguments Arguments) (bool, error) {
	upstream, err := runLoggedOutput(repository, arguments, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	if err != nil {
		return false, nil
	}

	count, err := runLoggedOutput(repository, arguments, "git", "rev-list", "--count", strings.TrimSpace(upstream)+"..HEAD")
	if err != nil {
		return false, fmt.Errorf("no se pudo comprobar el estado remoto de %s: %w", repository, err)
	}

	return strings.TrimSpace(count) != "0", nil
}

// runLoggedOutput executes a command, logs its combined output and returns that output.
func runLoggedOutput(directory string, arguments Arguments, name string, args ...string) (string, error) {
	command := exec.Command(name, args...)
	command.Dir = directory

	output, err := command.CombinedOutput()
	if logErr := appendLog(arguments.LogFile, output); logErr != nil {
		return "", logErr
	}

	return string(output), err
}

// appendLog appends raw command output to the execution log.
func appendLog(logFile string, data []byte) error {
	log, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("no se puede abrir el log %s: %w", logFile, err)
	}
	defer log.Close()

	_, err = log.Write(data)
	return err
}

// runRelease promotes the selected repositories and returns those completed successfully.
// With -f, each repository is committed and pushed only after its release succeeds.
func runRelease(arguments Arguments, repositories []string) []string {
	released := []string{}

	for _, repository := range repositories {
		if !isGitRepository(repository) {
			continue
		}

		cli.Info("Generando release de %s", filepath.Base(repository))
		err := releaseRepository(repository, arguments)
		if err == nil && arguments.Full {
			err = commitRepository(repository, arguments)
		}
		if err == nil {
			released = append(released, repository)
			continue
		}

		if arguments.Tolerant {
			cli.Warning("%v", err)
			continue
		}

		cli.Error("%v", err)
		cli.Error("Consulta el log: %s", arguments.LogFile)
		os.Exit(RC.Release)
	}

	return released
}

// runDeploy executes the complete build/publish/release/commit cycle when full.
func runDeploy(arguments Arguments, repositories []string) []string {
	if arguments.Full {
		repositories = runPublish(arguments, repositories)

		releaseArguments := arguments
		releaseArguments.Full = false
		repositories = runRelease(releaseArguments, repositories)

		arguments.Full = false
	}

	return runCommit(arguments, repositories)
}

// printHelp shows the command-line help.
func printHelp() {
	cli.Direct(`IASI Dev

Usage:
  iasi-dev <command> [-s] [-v|-V] [-f] [-t] [-i] [--format value] [--key value] [directory...]

Commands:
  help       Show help
  build      Build
  publish    Publish
  release    Release
  commit     Commit and push
  deploy     Deploy
  sync       Sync shared files from iasi-common
`)
}
