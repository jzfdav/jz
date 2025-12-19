package main

import (
    "fmt"
    "jz/app"
    "jz/report"

    "github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
    Use:   "scan <root-path>",
    Short: "Scan a directory and generate a Markdown report",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        rootDir := args[0]
        services, sysGraph, diagnostic := app.Analyze(rootDir)
        fmt.Println(report.GenerateMarkdown(services, sysGraph, diagnostic))
    },
}

func init() {
    rootCmd.AddCommand(scanCmd)
}
