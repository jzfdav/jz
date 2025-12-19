package main

import (
	"fmt"
	"jz/app"
	"jz/report"
	"os"

	"github.com/spf13/cobra"
)

var (
	mdService string
	mdOutput  string
)

var reportMarkdownCmd = &cobra.Command{
	Use:   "markdown <root-path>",
	Short: "Generate a Markdown report",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir := args[0]
		services, sysGraph := app.Analyze(rootDir)

		// Filter
		services, sysGraph, err := filterData(services, sysGraph, mdService)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Generate
		content := report.GenerateMarkdown(services, sysGraph)

		// Output
		if err := writeOutput(content, mdOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	reportMarkdownCmd.Flags().StringVar(&mdService, "service", "", "Filter by service name")
	reportMarkdownCmd.Flags().StringVar(&mdOutput, "output", "", "Write output to file")
	reportCmd.AddCommand(reportMarkdownCmd)
}
