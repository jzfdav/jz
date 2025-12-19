package report

import (
	"fmt"
	"jz/model"
	"strings"
)

// GenerateSystemMermaid creates a Mermaid graph for system-level dependencies.
func GenerateSystemMermaid(services []model.Service, sysGraph model.SystemGraph) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Dependencies
	for _, dep := range sysGraph.Dependencies {
		from := sanitize(dep.FromService)
		to := sanitize(dep.ToService)
		// A -->|Label| B
		sb.WriteString(fmt.Sprintf("\t%s -->|%s| %s\n", from, dep.Interface, to))
	}

	// Create a set of nodes involved in edges
	involved := make(map[string]bool)
	for _, dep := range sysGraph.Dependencies {
		involved[dep.FromService] = true
		involved[dep.ToService] = true
	}

	// List all nodes with labels
	for _, svcName := range sysGraph.Services {
		label := svcName
		// Check if this is a Liberty WAR service
		for _, s := range services {
			if s.Name == svcName && s.Application.Type == "webApplication" {
				label += " (WAR)"
				break
			}
		}

		id := sanitize(svcName)
		sb.WriteString(fmt.Sprintf("\t%s[%s]\n", id, label))
	}

	return sb.String()
}

// GenerateComponentMermaid creates a Mermaid graph for internal component dependencies.
func GenerateComponentMermaid(graph model.DependencyGraph) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Edges
	for _, edge := range graph.Edges {
		from := sanitize(edge.FromComponent)
		to := sanitize(edge.ToComponent)
		sb.WriteString(fmt.Sprintf("\t%s -->|%s| %s\n", from, edge.Interface, to))
	}

	// Nodes definitions (to ensure isolated components show up)
	for _, node := range graph.Nodes {
		id := sanitize(node.Name)
		// Escape the label if needed, but keeping it simple as per requirements
		sb.WriteString(fmt.Sprintf("\t%s[%s]\n", id, node.Name))
	}

	return sb.String()
}

// sanitize creates a valid Mermaid identifier.
func sanitize(name string) string {
	// Replace invalid chars with underscore
	// internal.MyComponent -> internal_MyComponent
	s := strings.ReplaceAll(name, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	return s
}
