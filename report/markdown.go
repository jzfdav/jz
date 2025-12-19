package report

import (
	"fmt"
	"jz/model"
	"strings"
)

// GenerateMarkdown creates a human-readable Markdown report from the analysis results.
func GenerateMarkdown(services []model.Service, sysGraph model.SystemGraph) string {
	var sb strings.Builder

	// 1. System Overview
	sb.WriteString("# System Overview\n\n")
	sb.WriteString(fmt.Sprintf("- Total number of services: %d\n", len(services)))
	sb.WriteString(fmt.Sprintf("- Total number of system-level dependencies: %d\n", len(sysGraph.Dependencies)))
	sb.WriteString("\n")

	// 2. Services
	sb.WriteString("# Services\n\n")
	for _, svc := range services {
		sb.WriteString(fmt.Sprintf("## %s\n\n", svc.Name))
		sb.WriteString(fmt.Sprintf("- Root Path: %s\n", svc.RootPath))
		sb.WriteString(fmt.Sprintf("- REST Entry Points: %d\n", len(svc.EntryPoints)))
		sb.WriteString(fmt.Sprintf("- DS Components: %d\n", len(svc.Components)))

		if svc.ServerName != "" {
			sb.WriteString(fmt.Sprintf("- Liberty Server: %s\n", svc.ServerName))
		}

		if len(svc.Features) > 0 {
			sb.WriteString("- Enabled Features:\n")
			for _, f := range svc.Features {
				sb.WriteString(fmt.Sprintf("  - %s\n", f))
			}
		}
		sb.WriteString("\n")
	}

	// 3. REST Entry Points
	sb.WriteString("# REST Entry Points\n\n")
	for _, svc := range services {
		if len(svc.EntryPoints) > 0 {
			sb.WriteString(fmt.Sprintf("## %s\n\n", svc.Name))
			for _, ep := range svc.EntryPoints {
				sb.WriteString(fmt.Sprintf("- %s %s (%s)\n", ep.Method, ep.Path, ep.Handler))
			}
			sb.WriteString("\n")
		}
	}

	// 4. Internal Component Dependencies
	sb.WriteString("# Internal Component Dependencies\n\n")
	for _, svc := range services {
		sb.WriteString(fmt.Sprintf("## %s\n\n", svc.Name))
		if len(svc.InternalGraph.Edges) == 0 {
			sb.WriteString("No internal component dependencies.\n\n")
		} else {
			for _, edge := range svc.InternalGraph.Edges {
				sb.WriteString(fmt.Sprintf("- %s -> %s (%s)\n", edge.FromComponent, edge.ToComponent, edge.Interface))
			}
			sb.WriteString("\n")
		}
	}

	// 5. System-Level Dependencies
	sb.WriteString("# System-Level Dependencies\n\n")
	if len(sysGraph.Dependencies) == 0 {
		sb.WriteString("No system-level dependencies.\n")
	} else {
		for _, dep := range sysGraph.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s -> %s (%s)\n", dep.FromService, dep.ToService, dep.Interface))
		}
	}

	return sb.String()
}
