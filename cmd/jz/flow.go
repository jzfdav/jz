package main

import (
	"fmt"
	"jz/app"
	"jz/report"
	"os"

	"github.com/spf13/cobra"
)

var (
	flowResource string
	flowMethod   string
	flowPath     string
	flowMaxDepth int
	flowFormat   string
	flowOutput   string
)

var flowCmd = &cobra.Command{
	Use:   "flow [path]",
	Short: "Extract execution flow for a specific REST resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if flowResource == "" {
			fmt.Fprintln(os.Stderr, "Error: --resource is required")
			os.Exit(1)
		}

		rootDir := args[0]
		services, _, _ := app.Analyze(rootDir)

		flows, err := app.ExtractFlow(services, flowResource, flowMethod, flowPath, flowMaxDepth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		var output string
		switch flowFormat {
		case "markdown":
			output = report.GenerateFlowMarkdown(flows, flowResource, flowPath)
		case "mermaid":
			output = report.GenerateFlowMermaid(flows, flowResource)
		case "all":
			md := report.GenerateFlowMarkdown(flows, flowResource, flowPath)
			mmd := report.GenerateFlowMermaid(flows, flowResource)
			output = md + "\n\n---\n\n" + mmd
		default:
			fmt.Fprintf(os.Stderr, "Error: invalid format '%s'\n", flowFormat)
			os.Exit(1)
		}

		if flowOutput != "" {
			err := os.WriteFile(flowOutput, []byte(output), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Output written to %s\n", flowOutput)
		} else {
			fmt.Println(output)
		}
	},
}

func init() {
	flowCmd.Flags().StringVar(&flowResource, "resource", "", "REST resource class name (required)")
	flowCmd.Flags().StringVar(&flowMethod, "method", "", "Filter to a single HTTP method")
	flowCmd.Flags().StringVar(&flowPath, "path", "", "Filter to a single REST path")
	flowCmd.Flags().IntVar(&flowMaxDepth, "max-depth", 3, "Limit call expansion depth")
	flowCmd.Flags().StringVar(&flowFormat, "format", "markdown", "Output format: markdown|mermaid|all")
	flowCmd.Flags().StringVar(&flowOutput, "output", "", "Output file path")

	rootCmd.AddCommand(flowCmd)
}
