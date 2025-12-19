package report

import (
	"fmt"
	"jz/model"
	"strings"
)

// GenerateSystemMermaid creates a Mermaid graph for system-level dependencies.
func GenerateSystemMermaid(sysGraph model.SystemGraph) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Dependencies
	for _, dep := range sysGraph.Dependencies {
		from := sanitize(dep.FromService)
		to := sanitize(dep.ToService)
		// A -->|Label| B
		sb.WriteString(fmt.Sprintf("\t%s -->|%s| %s\n", from, dep.Interface, to))
	}

	// If no dependencies, lift isolated nodes just to show them?
	// Requirements say: "If there are no system-level dependencies: Generate a diagram with services only (no edges)"
	// So we should list all services as nodes to ensure they appear even if isolated.
	// But Mermaid handles nodes implicitly if they are in edges.
	// We need to explicitly list isolated nodes if they are not in edges.
	// Or better: list ALL nodes to be safe and ensure consistent naming.

	// Create a set of nodes involved in edges
	involved := make(map[string]bool)
	for _, dep := range sysGraph.Dependencies {
		involved[dep.FromService] = true
		involved[dep.ToService] = true
	}

	// List nodes that are NOT involved or just all nodes with labels?
	// Simpler: Just list all nodes to define them.
	for _, svc := range sysGraph.Services {
		id := sanitize(svc)
		sb.WriteString(fmt.Sprintf("\t%s[%s]\n", id, svc))
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
