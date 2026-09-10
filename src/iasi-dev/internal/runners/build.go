package runners

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/R"
	"iasi-dev/internal/consts/RC"
	iasiPkg "iasi-dev/internal/iasi"
	"iasi-dev/internal/structures"
)

// Build builds the selected IASI projects and returns the projects that remain active.
func Build(Parms *structures.Parms) []string {
	if Parms.Debug { fmt.Printf("Build: projects=%v\n", Parms.Projects) }
	targets := []string{}
	rc := RC.OK

	for _, project := range Parms.Projects {
		cli.Info(*Parms, "Construyendo %s", filepath.Base(project))

		iasi, ok := iasiPkg.Read(project, *Parms)
		if !ok {
			rc = checkTolerant(*Parms, RC.Build)
		} else {
			switch iasi.Type {
			case R.PACKAGE: rc = buildPackage(project, *Parms)
			case R.WEBSITE, R.BOOK: rc = buildQuarto(project, *Parms)
			default:
				cli.Warning(*Parms, "********** TIPO IASI NO SOPORTADO: %q (%s) **********", iasi.Type, filepath.Base(project))
				rc = RC.Skip
			}
		}

		if rc == RC.Skip {
			addToBlackList(Parms, project)
			continue
		}

		targets = append(targets, project)
	}

	return targets
}

// buildPackage builds the source and binary distributions of an R package.
func buildPackage(project string, Parms structures.Parms) int {
	result := commands.RunFriendlyLogged(project, Parms.LogFile, "rinstall", "source")
	if RC.IsErroneous(result.RC) { return checkTolerant(Parms, RC.Build) }

	result = commands.RunFriendlyLogged(project, Parms.LogFile, "rinstall", "binary")
	if RC.IsErroneous(result.RC) { return checkTolerant(Parms, RC.Build) }

	return RC.OK
}

// buildQuarto builds an IASI project through iasi.quarto.
func buildQuarto(project string, Parms structures.Parms) int {
	parameters := []string{}

	if Parms.Format != "" {
		formats := strings.Split(Parms.Format, ",")
		for i, format := range formats { formats[i] = strconv.Quote(strings.TrimSpace(format)) }

		parameters = append(parameters, "format = c("+strings.Join(formats, ", ")+")")
	}

	if Parms.Force { parameters = append(parameters, "force = TRUE") }

	expression := "iasi.quarto::build(" + strings.Join(parameters, ", ") + ")"

	result := commands.RunFriendlyLogged(project, Parms.LogFile, "Rscript", "-e", expression)
	if RC.IsErroneous(result.RC) { return checkTolerant(Parms, RC.Build) }

	return RC.OK
}
