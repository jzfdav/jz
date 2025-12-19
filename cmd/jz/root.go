package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jz",
	Short: "jz is a static analysis tool for legacy Java systems",
	Long:  `jz performs static analysis on Java codebases, focusing on OSGi and JAX-RS constructs.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
