package graph

import (
	"jz/model"
)

// BuildSystemGraph constructs a system-level dependency graph from a list of Services.
func BuildSystemGraph(services []model.Service) model.SystemGraph {
	graph := model.SystemGraph{
		Services:     make([]string, 0),
		Dependencies: make([]model.ServiceDependency, 0),
	}

	// Index providers: interface -> []serviceName
	providers := make(map[string][]string)

	for _, svc := range services {
		graph.Services = append(graph.Services, svc.Name)

		// A service provides an interface if ANY of its components provide it
		visitedInterfaces := make(map[string]bool)
		for _, comp := range svc.Components {
			for _, iface := range comp.ProvidedInterfaces {
				if !visitedInterfaces[iface] {
					providers[iface] = append(providers[iface], svc.Name)
					visitedInterfaces[iface] = true
				}
			}
		}
	}

	// Create Edges
	// We need to deduplicate edges: (From, To, Interface) tuple must be unique
	type edgeKey struct {
		From, To, Iface string
	}
	seenEdges := make(map[edgeKey]bool)

	for _, svc := range services {
		for _, comp := range svc.Components {
			for _, refIface := range comp.ReferencedInterfaces {
				// Find matching provider services
				if providerServices, ok := providers[refIface]; ok {
					for _, providerName := range providerServices {
						// Ignore self-dependencies
						if providerName == svc.Name {
							continue
						}

						key := edgeKey{
							From:  svc.Name,
							To:    providerName,
							Iface: refIface,
						}

						if !seenEdges[key] {
							dep := model.ServiceDependency{
								FromService: svc.Name,
								ToService:   providerName,
								Interface:   refIface,
							}
							graph.Dependencies = append(graph.Dependencies, dep)
							seenEdges[key] = true
						}
					}
				}
			}
		}
	}

	return graph
}
