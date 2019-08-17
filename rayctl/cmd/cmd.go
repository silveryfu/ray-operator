package main

import (
	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"os"
)

func newRayctlCommand() *cobra.Command {
	return &cobra.Command{
		Use:              "rayctl COMMAND",
		Short:            "Cli for managing ray applications",
		SilenceUsage:     false,
		SilenceErrors:    true,
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				os.Exit(1)
			}
		},
	}
}

func newSubmitCommand() *cobra.Command {
	return &cobra.Command{
		Use:              "submit COMMAND",
		Short:            "Submit a ray application",
		SilenceUsage:     false,
		SilenceErrors:    true,
		TraverseChildren: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO: to submit")

			if err := cmd.Help(); err != nil {
				os.Exit(1)
			}
		},
	}
}

func main() {
	// set logging
	klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(klogFlags)

	// set flags
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	flag.Parse()

	rootCmd := newRayctlCommand()
	rootCmd.AddCommand(newSubmitCommand())

	if err := rootCmd.Execute(); err != nil {
		klog.Error(err)
		os.Exit(1)
	}
}
