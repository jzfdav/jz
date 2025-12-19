package main

import (
	"fmt"
	"jz/app"
	"jz/report"

	"github.com/spf13/cobra"
)

var reportMarkdownCmd = &cobra.Command{
	Use:   "markdown <root-path>",
	Short: "Generate a Markdown report",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir := args[0]
		services, sysGraph := app.Analyze(rootDir)
		fmt.Println(report.GenerateMarkdown(services, sysGraph))
	},
}

func init() {
	reportCmd.AddCommand(reportMarkdownCmd)
}
