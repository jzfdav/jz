package main

import (
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate reports from the analysis",
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
