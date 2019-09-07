package main

import (
	"github.com/suiqirui1987/kstrace/pkg/cmd"
	"os"

	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func main() {
	flags := pflag.NewFlagSet("kstrace", pflag.ExitOnError)
	pflag.CommandLine = flags

	root := cmd.NewKStraceCommand(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
