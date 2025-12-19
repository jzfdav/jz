package main

import (
	"fmt"
	"jz/app"
	"jz/report"

	"github.com/spf13/cobra"
)

var reportMermaidCmd = &cobra.Command{
	Use:   "mermaid <root-path>",
	Short: "Generate Mermaid diagrams",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir := args[0]
		services, sysGraph := app.Analyze(rootDir)

		// System Level
		fmt.Println(report.GenerateSystemMermaid(sysGraph))

		// Component Level (per service)
		for _, svc := range services {
			fmt.Println(report.GenerateComponentMermaid(svc.InternalGraph))
		}
	},
}

func init() {
	reportCmd.AddCommand(reportMermaidCmd)
}
