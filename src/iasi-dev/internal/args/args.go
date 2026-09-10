package args

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
	"iasi-dev/internal/tools"
)

// Parse parses command-line arguments and prepares the effective targets.
func Parse(command string, values []string) structures.Parms {
	subcommand := ""

	if command == "workflow" {
		if len(values) == 0 { cli.Error(RC.InvalidArguments, structures.Parms{}, "Falta el comando del workflow.") }

		subcommand = values[0]
		values = values[1:]
	}

	Parms := parseArguments(values)
	Parms.RequestedTargets = append([]string{}, Parms.Targets...)
	Parms.Subcommand = subcommand
	Prepare(&Parms)

	return Parms
}

func parseArguments(args []string) structures.Parms {
	Parms := structures.Parms{
		Verbose:    1,
		Exclusions: append([]string{}, consts.RequiredExclusions...),
	}

	for i := 0; i < len(args); i++ {
		if len(args[i]) == 0 { invalidArgument(&Parms, args[i]) }

		switch args[i][0] {
		case '-': parseFlagOrParameter(args, &i, &Parms)
		default:  parseTarget(args, i, &Parms)
		}
	}

	return Parms
}

func parseTarget(args []string, i int, Parms *structures.Parms) {
	Parms.Targets = append(Parms.Targets, args[i])
}

func parseFlagOrParameter(args []string, i *int, Parms *structures.Parms) {
	switch len(args[*i]) {
	case 1:  invalidArgument(Parms, args[*i])
	case 2:  parseFlag(args, *i, Parms)
	default: parseParameter(args, i, Parms)
	}
}

func parseFlag(args []string, i int, Parms *structures.Parms) {
	if args[i][1] == '-' { invalidArgument(Parms, args[i]) }

	switch args[i][1] {
	case 'a': Parms.All = true
	case 'c': Parms.Checkpoints = true
	case 'd': Parms.Debug = true
	case 'f': Parms.Force = true
	case 'h': Parms.Help = true
	case 'i': Parms.Install = true
	case 'l': Parms.Local = true
	case 's': Parms.Verbose = 0
	case 't': Parms.Tolerant = true
	case 'v': Parms.Verbose = 3
	case 'V': Parms.Verbose = 7
	default:  invalidArgument(Parms, args[i])
	}
}

func parseParameter(args []string, i *int, Parms *structures.Parms) {
	if args[*i][1] != '-' { invalidArgument(Parms, args[*i]) }
	if *i + 1 >= len(args) { missingParameterValue(Parms, args[*i]) }

	name := args[*i][2:]
	(*i)++
	value := args[*i]

	validateParameter(Parms, name, value)
}

// validateParameter validates and applies one long parameter.
func validateParameter(Parms *structures.Parms, name string, value string) {
	switch name {
	case "exclude": processExclusions(Parms, value)
	case "format":  Parms.Format = value
	case "message": Parms.Message = value
	default:        invalidArgument(Parms, "--"+name)
	}
}

func invalidArgument(Parms *structures.Parms, argument string) {
	cli.Error(RC.InvalidArguments, *Parms, "Argumento no válido: %q", argument)
}

func missingParameterValue(Parms *structures.Parms, parameter string) {
	cli.Error(RC.InvalidArguments, *Parms, "Falta el valor del parámetro: %q", parameter)
}

// Prepare discovers the effective targets.
func Prepare(Parms *structures.Parms) {
	processTargets(Parms)
}

// processExclusions adds exclusions supplied directly or through files.
func processExclusions(Parms *structures.Parms, values string) {
	for _, value := range strings.Split(values, ",") {
		value = strings.TrimSpace(value)
		if value == "" { continue }

		if tools.IsFile(value) {
			addExclusionsFile(Parms, value)
			continue
		}

		Parms.Exclusions = append(Parms.Exclusions, value)
	}

	Parms.Exclusions = uniqueStrings(Parms.Exclusions)
}

// addExclusionsFile adds one exclusion per non-empty line of path.
func addExclusionsFile(Parms *structures.Parms, path string) {
	file, err := os.Open(path)
	if err != nil { cli.Error(RC.InvalidArguments, *Parms, "No se puede leer el fichero de exclusiones: %q", path) }
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		value := strings.TrimSpace(scanner.Text())
		if value == "" { continue }
		Parms.Exclusions = append(Parms.Exclusions, value)
	}

	if err := scanner.Err(); err != nil { cli.Error(RC.InvalidArguments, *Parms, "Error leyendo el fichero de exclusiones: %q", path) }
}

// processTargets resolves discovery roots and finds repositories and IASI projects recursively.
func processTargets(Parms *structures.Parms) {
	roots := Parms.Targets
	if len(roots) == 0 { roots = []string{"."} }

	Parms.Targets = []string{}
	Parms.Repos = []string{}
	Parms.Projects = []string{}
	Parms.ProjectRepos = map[string]string{}
	Parms.BlackList = []string{}

	for _, root := range roots {
		path, err := filepath.Abs(root)
		if err != nil {
			cli.Warning(*Parms, "Se ignora %q: no se puede resolver la ruta.", root)
			continue
		}

		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			cli.Warning(*Parms, "Se ignora %q: no existe o no es un directorio.", root)
			continue
		}

		path = filepath.Clean(path)
		Parms.Targets = append(Parms.Targets, path)
		discoverTargets(Parms, path, "")
	}

	Parms.Targets = uniqueStrings(Parms.Targets)
	Parms.Repos = uniqueStrings(Parms.Repos)
	Parms.Projects = uniqueStrings(Parms.Projects)
}

// discoverTargets recursively discovers repositories and IASI projects below path.
func discoverTargets(Parms *structures.Parms, path string, repository string) {
	if isExcluded(Parms, filepath.Base(path)) { return }

	if tools.IsRepo(path) {
		repository = path
		Parms.Repos = append(Parms.Repos, path)
	}

	if tools.IsProject(path) {
		Parms.Projects = append(Parms.Projects, path)
		if repository != "" { Parms.ProjectRepos[path] = repository }
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		cli.Warning(*Parms, "Se ignora %q: no se puede leer.", path)
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || isExcluded(Parms, entry.Name()) { continue }
		discoverTargets(Parms, filepath.Join(path, entry.Name()), repository)
	}
}

// isExcluded reports whether name is excluded from target discovery.
func isExcluded(Parms *structures.Parms, name string) bool {
	for _, exclusion := range Parms.Exclusions {
		if name == exclusion { return true }
	}
	return false
}

// uniqueStrings returns values without duplicates, preserving their order.
func uniqueStrings(values []string) []string {
	unique := []string{}
	seen := map[string]bool{}

	for _, value := range values {
		key := filepath.Clean(value)
		if seen[key] { continue }

		seen[key] = true
		unique = append(unique, value)
	}

	return unique
}
