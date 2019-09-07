package main

import (
	"os"

	"github.com/spf13/pflag"
	"github.com/suiqirui1987/kstrace/docker/strace"
)

func main() {
	flags := pflag.NewFlagSet("strace-exec", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := strace.NewStraceExecCommand()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
