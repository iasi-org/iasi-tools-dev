package cli

import (
	"fmt"
	"io"
	"os"
	"time"
)

type messageLevel int
type messageVisibility int

const (
	levelDebug messageLevel = iota
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
	colorGray   = "\033[90m"
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
)

var outputOptions OutputOptions

func ConfigureOutput(options OutputOptions) {
	outputOptions = options
}

func Direct(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLog(message)

	if outputOptions.Verbose == 0 {
		return
	}

	fmt.Fprint(os.Stdout, message)
}

func Debug(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	writeLog(fmt.Sprintf("%s - %s\n", time.Now().Format("15:04:05"), message))

	if !outputOptions.Debug {
		return
	}

	writeConsoleMessage(os.Stdout, levelDebug, false, message)
}

func Verbose(format string, args ...any) {
	writeMessage(os.Stdout, visibilityVerbose, levelInfo, false, format, args...)
}

func VeryVerbose(format string, args ...any) {
	writeMessage(os.Stdout, visibilityVeryVerbose, levelSuccess, false, format, args...)
}

func Info(format string, args ...any) {
	writeMessage(os.Stdout, visibilityNormal, levelInfo, true, format, args...)
}

func Success(format string, args ...any) {
	writeMessage(os.Stdout, visibilityNormal, levelSuccess, false, format, args...)
}

func Warning(format string, args ...any) {
	writeMessage(os.Stderr, visibilityNormal, levelWarning, false, format, args...)
}

func Error(format string, args ...any) {
	writeMessage(os.Stderr, visibilityNormal, levelError, false, format, args...)
}

func writeMessage(writer io.Writer, visibility messageVisibility, level messageLevel, bold bool, format string, args ...any) {
	now := time.Now()
	message := fmt.Sprintf(format, args...)
	writeLog(fmt.Sprintf("%s - %s\n", now.Format("15:04:05"), message))

	if !messageVisible(visibility) {
		return
	}

	style := messageColor(level)
	if bold {
		style = colorBold + style
	}

	fmt.Fprintf(writer, "%s - %s%s%s\n", now.Format("15:04:05"), style, message, colorReset)
}

func writeConsoleMessage(writer io.Writer, level messageLevel, bold bool, message string) {
	style := messageColor(level)
	if bold {
		style = colorBold + style
	}

	fmt.Fprintf(writer, "%s - %s%s%s\n", time.Now().Format("15:04:05"), style, message, colorReset)
}

func writeLog(message string) {
	if outputOptions.LogFile == "" {
		return
	}

	file, err := os.OpenFile(outputOptions.LogFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	fmt.Fprint(file, message)
}

func messageVisible(visibility messageVisibility) bool {
	return outputOptions.Verbose&int(visibility) != 0
}

func messageColor(level messageLevel) string {
	switch level {
	case levelDebug:
		return colorGray
	case levelInfo:
		return colorBlue
	case levelSuccess:
		return colorGreen
	case levelWarning:
		return colorYellow
	case levelError:
		return colorRed
	default:
		return colorReset
	}
}
