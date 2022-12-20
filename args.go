package complete

import (
	"github.com/monk-io/complete/v2/internal/arg"
	"strconv"
)

func parseArgs() []arg.Arg {
	var (
		line  = getEnv("COMP_LINE")
		point = getEnv("COMP_POINT")
	)
	if line == "" {
		return nil
	}
	i, err := strconv.Atoi(point)
	if err != nil {
		panic("COMP_POINT env should be integer, got: " + point)
	}
	if i > len(line) {
		i = len(line)
	}

	// Parse the command line up to the completion point.
	args := arg.Parse(line[:i])

	// The first word is the current command name.
	args = args[1:]
	return args
}

func ParseArgs() []arg.Arg {
	return parseArgs()
}

func GetFlagValue(name string) string {
	flags := parseArgs()

	var found bool
	for _, flag := range flags {
		if found && flag.HasValue {
			return flag.Value
		}
		if flag.Flag == name {
			found = true
		}
	}

	return ""
}
