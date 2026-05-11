package main

import (
	"flag"
	"fmt"
	"os"
)

// overridden by -ldflags -X
var version = "unknown"

type cmdFn func(name string, args []string, usage func()) int

var cmds = map[string]cmdFn{
	"passwd": passwd,
}

func cmdUsage(f *flag.FlagSet, parentUsage func(), cmds map[string]cmdFn, help string) {
	f.Usage = func() {
		if parentUsage != nil {
			parentUsage()
			fmt.Fprintln(f.Output(), "")
		}
		if len(cmds) > 0 {
			fmt.Fprintf(f.Output(), "%s [flags] command [flags] %s\n", f.Name(), help)
		} else {
			fmt.Fprintf(f.Output(), "... %s [flags] %s\n", f.Name(), help)
		}
		fmt.Fprintln(f.Output(), "\nFlags:")
		f.PrintDefaults()
		if len(cmds) > 0 {
			fmt.Fprintln(f.Output(), "\nCommands:")
			for cmd := range cmds {
				fmt.Fprintf(f.Output(), "  %s\n", cmd)
			}
		}
	}
}

func flagUsageExit(f *flag.FlagSet, msg string, code int) {
	fmt.Fprintln(f.Output(), msg)
	f.Usage()
	os.Exit(code)
}

func main() {
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var (
		flVersion = f.Bool("version", false, "print version and exit")
	)
	cmdUsage(f, nil, cmds, "")
	f.Parse(os.Args[1:])

	if *flVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if len(f.Args()) < 1 {
		flagUsageExit(f, "missing command", 2)
	}

	cmd, ok := cmds[f.Args()[0]]
	if !ok {
		flagUsageExit(f, "invalid command", 2)
	}

	os.Exit(cmd(f.Args()[0], f.Args()[1:], f.Usage))
}
