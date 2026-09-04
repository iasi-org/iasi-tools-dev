package cli

type OutputOptions struct {
	Verbose int
	Debug   bool
	LogFile string
}

func ParseOutputOptions(args []string) (OutputOptions, []string) {
	options := OutputOptions{Verbose: 1}
	remaining := make([]string, 0, len(args))
	silent := false

	for _, arg := range args {
		switch arg {
		case "-s":
			silent = true
		case "-v":
			if options.Verbose < 3 {
				options.Verbose = 3
			}
		case "-V":
			options.Verbose = 7
		case "-d":
			options.Debug = true
		default:
			remaining = append(remaining, arg)
		}
	}

	if silent {
		options.Verbose = 0
	}

	return options, remaining
}
