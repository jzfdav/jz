package main

import (
	"fmt"
	"jz/app"
	"jz/report"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "scan":
		if len(os.Args) != 3 {
			printUsage()
			os.Exit(1)
		}
		rootDir := os.Args[2]
		services, sysGraph := app.Analyze(rootDir)
		fmt.Println(report.GenerateMarkdown(services, sysGraph))

	case "report":
		if len(os.Args) != 4 {
			printUsage()
			os.Exit(1)
		}
		subCmd := os.Args[2]
		rootDir := os.Args[3]
		services, sysGraph := app.Analyze(rootDir)

		switch subCmd {
		case "markdown":
			fmt.Println(report.GenerateMarkdown(services, sysGraph))
		case "mermaid":
			// System Level
			fmt.Println(report.GenerateSystemMermaid(sysGraph))
			// Component Level (per service)
			for _, svc := range services {
				// Only print if there are edges or nodes to show?
				// Constraints say: "If a service has no internal dependencies: Generate a diagram with components only (no edges)"
				// The GenerateComponentMermaid functions handles this by listing nodes.
				fmt.Println(report.GenerateComponentMermaid(svc.InternalGraph))
			}
		default:
			printUsage()
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  jz scan <root-path>\n")
	fmt.Fprintf(os.Stderr, "  jz report markdown <root-path>\n")
	fmt.Fprintf(os.Stderr, "  jz report mermaid <root-path>\n")
}
