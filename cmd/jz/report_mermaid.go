package main

import (
	"fmt"
	"jz/app"
	"jz/report"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	mermaidService string
	mermaidOutput  string
)

var reportMermaidCmd = &cobra.Command{
	Use:   "mermaid <root-path>",
	Short: "Generate Mermaid diagrams",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rootDir := args[0]
		services, sysGraph := app.Analyze(rootDir)

		// Filter
		services, sysGraph, err := filterData(services, sysGraph, mermaidService)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var sb strings.Builder

		// System Level
		sb.WriteString(report.GenerateSystemMermaid(sysGraph))
		sb.WriteString("\n")

		// Component Level (per service)
		for _, svc := range services {
			sb.WriteString(report.GenerateComponentMermaid(svc.InternalGraph))
			sb.WriteString("\n")
		}

		// Output
		if err := writeOutput(sb.String(), mermaidOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	reportMermaidCmd.Flags().StringVar(&mermaidService, "service", "", "Filter by service name")
	reportMermaidCmd.Flags().StringVar(&mermaidOutput, "output", "", "Write output to file")
	reportCmd.AddCommand(reportMermaidCmd)
}
