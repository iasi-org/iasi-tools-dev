package runners

import (
	"fmt"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/RC"
	iasiPkg "iasi-dev/internal/iasi"
	"iasi-dev/internal/structures"
)

// Release releases the selected IASI projects and returns the projects that remain active.
func Release(Parms *structures.Parms) []string {
	if Parms.Debug { fmt.Printf("Release: projects=%v\n", Parms.Projects) }
	targets := []string{}
	rc := RC.OK

	for _, project := range Parms.Projects {
		cli.Info(*Parms, "Generando release de %s", filepath.Base(project))

		_, ok := iasiPkg.Read(project, *Parms)
		if !ok {
			rc = checkTolerant(*Parms, RC.Release)
		} else {
			rc = releaseQuarto(project, *Parms)
		}

		if rc == RC.Skip {
			addToBlackList(Parms, project)
			continue
		}

		targets = append(targets, project)
	}

	return targets
}

// releaseQuarto releases an IASI project through iasi.quarto.
func releaseQuarto(project string, Parms structures.Parms) int {
	parameters := []string{}

	if Parms.Force { parameters = append(parameters, "force = TRUE") }

	expression := "iasi.quarto::release(" + strings.Join(parameters, ", ") + ")"

	result := commands.RunFriendlyLogged(project, Parms.LogFile, "Rscript", "-e", expression)
	if RC.IsErroneous(result.RC) { return checkTolerant(Parms, RC.Release) }

	return RC.OK
}
