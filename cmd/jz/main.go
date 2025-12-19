package main

import (
	"fmt"
	"jz/graph"
	"jz/model"
	"jz/report"
	"jz/scan"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 3 || os.Args[1] != "scan" {
		fmt.Fprintf(os.Stderr, "Usage: jz scan <root-path>\n")
		os.Exit(1)
	}

	rootDir := os.Args[2]
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory '%s' does not exist\n", rootDir)
		os.Exit(1)
	}

	// 1. Discover OSGi Bundles
	bundles, err := scan.ScanOSGi(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning OSGi bundles: %v\n", err)
		os.Exit(1)
	}

	// 2. Extract REST Entry Points (global)
	entryPoints, err := scan.Scan(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning JAX-RS entry points: %v\n", err)
		os.Exit(1)
	}

	// 3. Find and Parse Liberty Server (if exists)
	var libertyServer model.LibertyServer
	var hasLiberty bool

	// Simple search for server.xml
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore walk errors
		}
		if !info.IsDir() && info.Name() == "server.xml" {
			srv, err := scan.ScanLiberty(path)
			if err == nil {
				libertyServer = srv
				hasLiberty = true
				return filepath.SkipDir // Stop after first server.xml found (simplification)
			}
		}
		return nil
	})
	// ignore walk errors

	// 4. Assemble Services
	var services []model.Service

	for _, bundle := range bundles {
		// Determine Service Root (parent of META-INF)
		// bundle.ManifestPath ends with META-INF/MANIFEST.MF
		metaInfDir := filepath.Dir(bundle.ManifestPath)
		serviceRoot := filepath.Dir(metaInfDir)

		svc := model.Service{
			Name:     bundle.SymbolicName,
			RootPath: serviceRoot,
		}

		// Attach Entry Points
		for _, ep := range entryPoints {
			// Check if ep.SourceFile is inside serviceRoot
			if strings.HasPrefix(ep.SourceFile, serviceRoot) {
				svc.EntryPoints = append(svc.EntryPoints, ep)
			}
		}

		// Parse DS Components
		var compPaths []string
		for _, sc := range bundle.ServiceComponents {
			// Service-Component entries are relative to bundle root (or contain OSGI-INF/...)
			// If wildcard, we need to glob.
			// The header is usually: OSGI-INF/component.xml, OSGI-INF/*.xml

			// We handle simple wildcards manually if needed, or assume they are paths.
			// scan.ScanDSComponents takes paths. We must resolve them.

			fullPattern := filepath.Join(serviceRoot, sc)
			matches, err := filepath.Glob(fullPattern)
			if err == nil {
				compPaths = append(compPaths, matches...)
			}
		}

		if len(compPaths) > 0 {
			comps, err := scan.ScanDSComponents(compPaths)
			if err == nil {
				svc.Components = comps
			}
		}

		// Build Internal Graph
		svc.InternalGraph = graph.BuildInternalGraph(svc.Components)

		// Attach Liberty Context
		if hasLiberty {
			svc.ServerName = libertyServer.Name
			svc.Features = libertyServer.EnabledFeatures
			// Try to match application
			for range libertyServer.DeployedApps {
				// Location might be relative or absolute.
				// Simple heuristic: check if location filename matches service folder name
				// or if resolved location points to service root.
				// For now, we leave Application struct empty unless exact match found?
				// Constraint says "Do NOT resolve application locations".
				// So maybe we just check basic string containment/match if explicitly clear.
				// Let's NOT guess too much.
			}
		}

		services = append(services, svc)
	}

	// 5. Build System Graph
	sysGraph := graph.BuildSystemGraph(services)

	// 6. Generate Report
	reportMD := report.GenerateMarkdown(services, sysGraph)

	// 7. Output
	fmt.Println(reportMD)
}
