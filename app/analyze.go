package app

import (
	"fmt"
	"jz/graph"
	"jz/model"
	"jz/scan"
	"os"
	"path/filepath"
	"strings"
)

// Diagnostic holds information about what was detected during analysis.
type Diagnostic struct {
	HasOSGi          bool
	HasLiberty       bool
	AnyManifestFound bool
	HasLibertyWAR    bool
}

// Analyze performs static analysis on the given root directory.
func Analyze(rootDir string) ([]model.Service, model.SystemGraph, Diagnostic) {
	diag := Diagnostic{}

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
	if len(bundles) > 0 {
		diag.HasOSGi = true
	}

	// Check for any MANIFEST.MF (even if not a valid OSGi bundle)
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.ToUpper(filepath.Base(path)) == "MANIFEST.MF" {
			diag.AnyManifestFound = true
			return filepath.SkipDir // Found at least one
		}
		return nil
	})

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
	filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // ignore walk errors
		}
		if !info.IsDir() && info.Name() == "server.xml" {
			srv, err := scan.ScanLiberty(path)
			if err == nil {
				libertyServer = srv
				hasLiberty = true
				diag.HasLiberty = true
				return filepath.SkipDir // Stop after first server.xml found
			}
		}
		return nil
	})

	// 4. Assemble Services
	var services []model.Service

	// 4a. OSGi Services
	for _, bundle := range bundles {
		// Determine Service Root (parent of META-INF)
		metaInfDir := filepath.Dir(bundle.ManifestPath)
		serviceRoot := filepath.Dir(metaInfDir)

		svc := model.Service{
			Name:     bundle.SymbolicName,
			RootPath: serviceRoot,
		}

		// Attach Entry Points
		for _, ep := range entryPoints {
			if strings.HasPrefix(ep.SourceFile, serviceRoot) {
				svc.EntryPoints = append(svc.EntryPoints, ep)
			}
		}

		// Parse DS Components
		var compPaths []string
		for _, sc := range bundle.ServiceComponents {
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
				// No complex resolution logic as per constraints
			}
		}

		services = append(services, svc)
	}

	// 4b. Phase F2: Liberty WAR Service Support
	// Detect when no OSGi bundles are present and Liberty is used.
	if len(services) == 0 && hasLiberty {
		var hasWebApp bool
		var libertyApp model.LibertyApp

		// Check for webApplication in server.xml
		for _, app := range libertyServer.DeployedApps {
			if app.Type == "webApplication" {
				hasWebApp = true
				libertyApp = app
				break
			}
		}

		// Check for WEB-INF/web.xml if not already found via server.xml
		if !hasWebApp {
			filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() && info.Name() == "web.xml" && filepath.Base(filepath.Dir(path)) == "WEB-INF" {
					hasWebApp = true
					return filepath.SkipDir
				}
				return nil
			})
		}

		if hasWebApp {
			diag.HasLibertyWAR = true

			name := libertyApp.ID
			if name == "" {
				name = filepath.Base(rootDir)
			}

			svc := model.Service{
				Name:        name,
				RootPath:    rootDir,
				EntryPoints: entryPoints, // All entry points in repo
				ServerName:  libertyServer.Name,
				Features:    libertyServer.EnabledFeatures,
				Application: libertyApp,
			}
			services = append(services, svc)
		}
	}

	// 5. Build System Graph
	sysGraph := graph.BuildSystemGraph(services)

	return services, sysGraph, diag
}
