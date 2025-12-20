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
		// Intentional O(n²) scan: Service counts are expected to be small, and this
		// approach favors code clarity over premature optimization.
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

// GenerateCallMermaid creates a Mermaid graph for REST resource interactions and boundaries.
func GenerateCallMermaid(services []model.Service) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	emittedNodes := make(map[string]bool)

	// 1. Boundaries as subgraphs
	for _, svc := range services {
		for _, b := range svc.Boundaries {
			boundaryID := sanitize(svc.Name + "_" + b.Identifier)
			sb.WriteString(fmt.Sprintf("\tsubgraph %s [%s: %s]\n", boundaryID, b.BoundaryType, b.Identifier))

			// Group resources by boundary prefix/identifier
			for _, res := range svc.RESTResources {
				// Identifiers: packages for OSGi, 'rest-api' for WAR
				if (b.BoundaryType == "package" && strings.HasPrefix(res.Name, b.Identifier)) ||
					(b.BoundaryType == "resource-group" && b.Identifier == "rest-api") {
					resID := sanitize(svc.Name + "_" + res.Name)
					sb.WriteString(fmt.Sprintf("\t\t%s\n", resID))
					emittedNodes[resID] = true
				}
			}
			sb.WriteString("\tend\n")
		}
	}

	// 2. Resource nodes (ensuring labels are set for all, including those not in subgraphs)
	for _, svc := range services {
		for _, res := range svc.RESTResources {
			resID := sanitize(svc.Name + "_" + res.Name)
			if !emittedNodes[resID] {
				sb.WriteString(fmt.Sprintf("\t%s[%s]\n", resID, res.Name))
				emittedNodes[resID] = true
			} else {
				// Even if already emitted in a subgraph, we apply the label here to be safe
				// Mermaid uses the first definition for the label.
				sb.WriteString(fmt.Sprintf("\t%s[%s]\n", resID, res.Name))
			}
		}
	}

	// 3. Edges for calls
	hasUnknown := false
	for _, svc := range services {
		// Outbound calls are already deduplicated per service in Analyze
		for _, res := range svc.RESTResources {
			fromID := sanitize(svc.Name + "_" + res.Name)
			for _, call := range res.OutboundCalls {
				// Intentional filtering: only show calls with high/medium confidence.
				// Low-confidence calls (e.g. partial AST matches or unresolved variables)
				// are omitted to prevent misleading "noise" in the visualization.
				if call.Confidence == model.ConfidenceLow || call.HTTPMethod == "" {
					continue
				}

				toID := "UNKNOWN"
				arrow := "-.->"
				if call.TargetService != "" && call.TargetResource != "" {
					toID = sanitize(call.TargetService + "_" + call.TargetResource)
					arrow = "-->"
				}

				if toID == "UNKNOWN" {
					hasUnknown = true
				}

				sb.WriteString(fmt.Sprintf("\t%s %s|%s| %s\n", fromID, arrow, call.HTTPMethod, toID))
			}
		}
	}

	if hasUnknown {
		sb.WriteString("\tUNKNOWN[UNKNOWN]\n")
	}

	return sb.String()
}

// sanitize creates a valid Mermaid identifier.
func sanitize(name string) string {
	// Replace invalid chars with underscore
	s := strings.ReplaceAll(name, ".", "_")
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	s = strings.ReplaceAll(s, "$", "")
	return s
}
