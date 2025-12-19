package main

import (
	"fmt"
	"jz/model"
	"os"
)

// filterData filters services and system graph based on the service name.
func filterData(services []model.Service, sysGraph model.SystemGraph, serviceName string) ([]model.Service, model.SystemGraph, error) {
	if serviceName == "" {
		return services, sysGraph, nil
	}

	// 1. Check if service exists
	var found bool
	for _, s := range services {
		if s.Name == serviceName {
			found = true
			break
		}
	}
	if !found {
		return nil, model.SystemGraph{}, fmt.Errorf("service '%s' not found", serviceName)
	}

	// 2. Filter Services
	var newServices []model.Service
	for _, s := range services {
		if s.Name == serviceName {
			newServices = append(newServices, s)
		}
	}

	// 3. Filter System Graph Dependencies
	var newDeps []model.ServiceDependency
	for _, dep := range sysGraph.Dependencies {
		if dep.FromService == serviceName || dep.ToService == serviceName {
			newDeps = append(newDeps, dep)
		}
	}

	// 4. Update System Graph
	newGraph := model.SystemGraph{
		Services:     []string{serviceName},
		Dependencies: newDeps,
	}

	return newServices, newGraph, nil
}

// writeOutput writes content to stdout or a file.
func writeOutput(content string, outputPath string) error {
	if outputPath == "" {
		fmt.Println(content)
		return nil
	}

	return os.WriteFile(outputPath, []byte(content), 0644)
}
