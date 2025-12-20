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
	flowCompact  bool
)

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Execution flow extraction and comparison",
	Long:  `jz flow provides tools to extract and compare REST execution flows.`,
}

var flowExtractCmd = &cobra.Command{
	Use:   "extract <path>",
	Short: "Extract execution flow for a specific REST resource",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runExtract(args[0])
	},
}

var flowDiffCmd = &cobra.Command{
	Use:   "diff <pathA> <pathB>",
	Short: "Compare execution flows between two code versions",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if flowResource == "" {
			fmt.Fprintln(os.Stderr, "Error: --resource is required")
			os.Exit(1)
		}

		pathA := args[0]
		pathB := args[1]

		servicesA, _, _ := app.Analyze(pathA)
		flowsA, err := app.ExtractFlow(servicesA, flowResource, flowMethod, flowPath, flowMaxDepth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing pathA: %v\n", err)
			os.Exit(1)
		}

		servicesB, _, _ := app.Analyze(pathB)
		flowsB, err := app.ExtractFlow(servicesB, flowResource, flowMethod, flowPath, flowMaxDepth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error analyzing pathB: %v\n", err)
			os.Exit(1)
		}

		diffs := app.DiffFlows(flowsA, flowsB)
		output := report.GenerateFlowDiffMarkdown(diffs, flowResource)

		if flowOutput != "" {
			err := os.WriteFile(flowOutput, []byte(output), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Diff output written to %s\n", flowOutput)
		} else {
			fmt.Println(output)
		}
	},
}

func runExtract(rootDir string) {
	if flowResource == "" {
		fmt.Fprintln(os.Stderr, "Error: --resource is required")
		os.Exit(1)
	}

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
		output = report.GenerateFlowMermaid(flows, flowResource, flowCompact)
	case "all":
		md := report.GenerateFlowMarkdown(flows, flowResource, flowPath)
		mmd := report.GenerateFlowMermaid(flows, flowResource, flowCompact)
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
}

func init() {
	// Root flags for flow
	flowCmd.PersistentFlags().StringVar(&flowResource, "resource", "", "REST resource class name (required)")
	flowCmd.PersistentFlags().StringVar(&flowMethod, "method", "", "Filter to a single HTTP method")
	flowCmd.PersistentFlags().StringVar(&flowPath, "path", "", "Filter to a single REST path")
	flowCmd.PersistentFlags().IntVar(&flowMaxDepth, "max-depth", 3, "Limit call expansion depth")
	flowCmd.PersistentFlags().StringVar(&flowFormat, "format", "markdown", "Output format: markdown|mermaid|all")
	flowCmd.PersistentFlags().StringVar(&flowOutput, "output", "", "Output file path")
	flowCmd.PersistentFlags().BoolVar(&flowCompact, "compact", false, "Enable visual-only guard chain compaction in Mermaid diagrams")

	// Allow "jz flow <path>" for backward compatibility if no subcommand is provided
	// We do this by setting Run on flowCmd to treat the first arg as path if it's not "extract" or "diff"
	flowCmd.Args = cobra.ArbitraryArgs
	flowCmd.Run = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		// If first arg is a subcommand name, cobra handled it.
		// If we are here, it means no subcommand matched.
		if len(args) == 1 {
			runExtract(args[0])
		} else {
			cmd.Help()
		}
	}

	flowCmd.AddCommand(flowExtractCmd)
	flowCmd.AddCommand(flowDiffCmd)

	rootCmd.AddCommand(flowCmd)
}
