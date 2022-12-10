package args

import "strings"

type Args struct {
	DebugMode bool
}

func Parse(args []string) Args {
	flags := args[1:]
	res := Args{}
	for _, flag := range flags {
		switch strings.ToLower(flag) {
		case "debug":
			res.DebugMode = true
		}
	}
	return res
}
