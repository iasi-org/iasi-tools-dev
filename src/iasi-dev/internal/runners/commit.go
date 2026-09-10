package runners

import (
	"fmt"
	"path/filepath"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/cli"
	"iasi-dev/internal/commands"
	"iasi-dev/internal/structures"
)

// Commit commits the selected repositories and returns the targets that remain active.
func Commit(Parms structures.Parms) []string {
	if Parms.Debug { fmt.Printf("Commit: repos=%v blackList=%v\n", Parms.Repos, Parms.BlackList) }
	targets := []string{}

	for _, repository := range Parms.Repos {
		if isBlackListed(Parms, repository) { continue }

		cli.Info(Parms, "Commit %s", filepath.Base(repository))

		rc := commitRepository(repository, Parms)
		switch rc {
		case RC.OK:    targets = append(targets, repository)
		case RC.Skip:
		case RC.Error: cli.Error(RC.Commit, Parms, "Error en commit de %s.", filepath.Base(repository))
		default:       cli.Error(RC.Commit, Parms, "Resultado inesperado en commit de %s: rc=%d.", filepath.Base(repository), rc)
		}
	}

	return targets
}

func commitRepository(repository string, Parms structures.Parms) int {
	if Parms.Debug { fmt.Printf("commitRepository: repository=%s tolerant=%t local=%t\n", repository, Parms.Tolerant, Parms.Local) }
	rc := changesPending(repository, Parms)
	if RC.IsErroneous(rc) { return checkTolerant(Parms, RC.Commit) }
	if rc == RC.NothingToDo { return RC.OK }

	rc = addChanges(repository, Parms)
	if RC.IsErroneous(rc) { return checkTolerant(Parms, RC.Commit) }

	rc = commitChanges(repository, Parms)
	if RC.IsErroneous(rc) { return checkTolerant(Parms, RC.Commit) }

	if !Parms.Local {
		rc = pushChanges(repository, Parms)
		if RC.IsErroneous(rc) { return checkTolerant(Parms, RC.Commit) }
	}

	return RC.OK
}

func changesPending(repository string, Parms structures.Parms) int {
	if Parms.Debug { fmt.Printf("changesPending: repository=%s\n", repository) }
	result := commands.RunFriendly(repository, Parms.LogFile, "git", "status")

	return result.RC
}

func addChanges(repository string, Parms structures.Parms) int {
	if Parms.Debug { fmt.Printf("addChanges: repository=%s\n", repository) }
	result := commands.RunLogged(repository, Parms.LogFile, "git", "add", "-A", ".")

	if result.RC != RC.OK { return RC.Error }

	return RC.OK
}

func commitChanges(repository string, Parms structures.Parms) int {
	if Parms.Debug { fmt.Printf("commitChanges: repository=%s message=%q\n", repository, Parms.Message) }
	result := commands.RunLogged(repository, Parms.LogFile, "git", "commit", "-m", Parms.Message)

	if result.RC != RC.OK { return RC.Error }

	return RC.OK
}

func pushChanges(repository string, Parms structures.Parms) int {
	if Parms.Debug { fmt.Printf("pushChanges: repository=%s\n", repository) }
	result := commands.RunLogged(repository, Parms.LogFile, "git", "push")

	if result.RC != RC.OK { return RC.Error }

	return RC.OK
}
