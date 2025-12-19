package graph

import (
	"jz/model"
)

// BuildInternalGraph constructs a dependency graph from a list of DS components.
func BuildInternalGraph(components []model.DSComponent) model.DependencyGraph {
	graph := model.DependencyGraph{
		Nodes: make([]model.ComponentNode, 0),
		Edges: make([]model.DependencyEdge, 0),
	}

	// Index providers: interface -> []componentName
	providers := make(map[string][]string)

	for _, comp := range components {
		// Create Node
		node := model.ComponentNode{
			Name:                comp.Name,
			ImplementationClass: comp.ImplementationClass,
			Immediate:           comp.Immediate,
		}
		graph.Nodes = append(graph.Nodes, node)

		// Index Provided Interfaces
		for _, iface := range comp.ProvidedInterfaces {
			providers[iface] = append(providers[iface], comp.Name)
		}
	}

	// Create Edges
	for _, comp := range components {
		for _, refIface := range comp.ReferencedInterfaces {
			// Find matching providers
			if providerNames, ok := providers[refIface]; ok {
				for _, providerName := range providerNames {
					// Ignore self-dependencies
					if providerName == comp.Name {
						continue
					}

					edge := model.DependencyEdge{
						FromComponent: comp.Name,
						ToComponent:   providerName,
						Interface:     refIface,
					}
					graph.Edges = append(graph.Edges, edge)
				}
			}
			// If not found, ignore (do not create dangling nodes)
		}
	}

	return graph
}
