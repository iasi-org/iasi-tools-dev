package runners

import (
	"iasi-dev/internal/cli"
	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

// Workflow executes the selected workflow.
func Workflow(Parms *structures.Parms) {
    projects := append([]string{}, Parms.Projects...)

    for _, project := range projects {
        Parms.Projects = []string{project}

        switch Parms.Subcommand {
            case "build":   workflowBuild(true, Parms)
            case "publish": workflowPublish(true, Parms)
            case "release": workflowRelease(true, Parms)
            default:        cli.Error(RC.InvalidArguments, *Parms, "Workflow desconocido: %q", Parms.Subcommand)
        }
    }
}
// workflowBuild builds projects and commits when standalone or used as a checkpoint.
func workflowBuild(standalone bool, Parms *structures.Parms) {
	Parms.Projects = Build(Parms)
	if standalone || Parms.Checkpoints { Commit(*Parms) }
}

// workflowPublish optionally builds first, publishes projects and commits when required.
func workflowPublish(standalone bool, Parms *structures.Parms) {
	if Parms.All { workflowBuild(false, Parms) }

	Parms.Projects = Publish(Parms)
	if standalone || Parms.Checkpoints { Commit(*Parms) }
}

// workflowRelease optionally runs previous stages, releases and always commits.
func workflowRelease(standalone bool, Parms *structures.Parms) {
	if Parms.All { workflowPublish(false, Parms) }

	Parms.Projects = Release(Parms)
	Commit(*Parms)
}
