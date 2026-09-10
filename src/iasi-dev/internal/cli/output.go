package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"iasi-dev/internal/consts/RC"
	"iasi-dev/internal/structures"
)

type messageLevel int
type messageVisibility int

const (
	levelVerbose messageLevel = iota
	levelInfo
	levelSuccess
	levelWarning
	levelError
)

const (
	visibilityNormal messageVisibility = 1 << iota
	visibilityVerbose
	visibilityVeryVerbose
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorBlue   = "\033[34m"
	colorWhite  = "\033[37m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

func Direct(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format, args...)
}

func Verbose(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stdout, visibilityVerbose, levelVerbose, false, format, args...)
}

func VeryVerbose(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stdout, visibilityVeryVerbose, levelSuccess, false, format, args...)
}

func Info(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stdout, visibilityNormal, levelInfo, true, format, args...)
}

func Success(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stdout, visibilityNormal, levelSuccess, false, format, args...)
}

func Warning(Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stderr, visibilityNormal, levelWarning, false, format, args...)
}

func Error(rc int, Parms structures.Parms, format string, args ...any) {
	writeMessage(Parms, os.Stderr, visibilityNormal, levelError, false, format, args...)
	if rc == RC.OK { return }
	if Parms.LogFile != nil { _ = Parms.LogFile.Close() }
	os.Exit(rc)
}

func writeMessage(Parms structures.Parms, writer io.Writer, visibility messageVisibility, level messageLevel, bold bool, format string, args ...any) {
	now := time.Now()
	message := fmt.Sprintf(format, args...)
	writeLog(Parms, fmt.Sprintf("%s - %s\n", now.Format("15:04:05"), message))

	if !messageVisible(Parms, visibility) { return }

	style := messageColor(level)
	if bold { style = colorBold + style }

	fmt.Fprintf(writer, "%s - %s%s%s\n", now.Format("15:04:05"), style, message, colorReset)
}

func writeLog(Parms structures.Parms, message string) {
	if Parms.LogFile == nil { return }
	fmt.Fprint(Parms.LogFile, message)
}

func messageVisible(Parms structures.Parms, visibility messageVisibility) bool {
	return Parms.Verbose&int(visibility) != 0
}

func messageColor(level messageLevel) string {
	switch level {
	case levelVerbose: return colorBlue
	case levelInfo:    return colorWhite
	case levelSuccess: return colorGreen
	case levelWarning: return colorYellow
	case levelError:   return colorRed
	default:           return colorReset
	}
}
