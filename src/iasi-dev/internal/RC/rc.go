// Package RC contains the process return codes used by iasi-dev.
package RC

const (
	OK               = 0
	Error            = 1
	InvalidArguments = 2
	Build            = 10
	Publish          = 20
	Commit           = 30
	Release          = 40
	Deploy           = 50
	Sync             = 60
)
