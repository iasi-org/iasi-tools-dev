package runners

import (
	"fmt"
	"path/filepath"
	"strings"

	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/consts/R"
	"iasi-dev/internal/consts/RC"
	iasiPkg "iasi-dev/internal/iasi"
	"iasi-dev/internal/structures"
)

// Publish publishes the selected IASI projects and returns the projects that remain active.
func Publish(Parms *structures.Parms) []string {
	if Parms.Debug { fmt.Printf("Publish: projects=%v\n", Parms.Projects) }
	targets := []string{}
	rc := RC.OK

	for _, project := range Parms.Projects {
		cli.Info(*Parms, "Publicando %s", filepath.Base(project))

		iasi, ok := iasiPkg.Read(project, *Parms)
		if !ok {
			rc = checkTolerant(*Parms, RC.Publish)
		} else {
			if !isPublishable(iasi.Type) {
				targets = append(targets, project)
				continue
			}
			rc = publishQuarto(project, *Parms)
		}

		if rc == RC.Skip {
			addToBlackList(Parms, project)
			continue
		}

		targets = append(targets, project)
	}

	return targets
}

// publishQuarto publishes an IASI project through iasi.quarto.
func publishQuarto(project string, Parms structures.Parms) int {
	parameters := []string{}

	if Parms.Force { parameters = append(parameters, "force = TRUE") }

	expression := "iasi.quarto::publish(" + strings.Join(parameters, ", ") + ")"

	result := commands.RunFriendlyLogged(project, Parms.LogFile, "Rscript", "-e", expression)
	if RC.IsErroneous(result.RC) { return checkTolerant(Parms, RC.Publish) }

	return RC.OK
}

// isPublishable reports whether an IASI project type can be published.
func isPublishable(projectType string) bool {
	switch projectType {
	case R.PACKAGE: return false
	default:        return true
	}
}
