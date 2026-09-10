// Package RC contains the process return codes used by iasi-dev.
package RC

const (
	OK               =  0
	InvalidArguments =  2
	Build            = 10
	Publish          = 20
	Commit           = 30
	Release          = 40
	Sync             = 60

	NothingToDo = 2
	Warning     = 4

	Error  = 16
	Severe = 32
	Fatal  = 64

	Skip = 256	
)

func IsErroneous(rc int) bool {
	return rc&0xF0 != 0
}