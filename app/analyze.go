package app

import (
	"bufio"
	"fmt"
	"jz/graph"
	"jz/model"
	"jz/scan"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Diagnostic holds information about what was detected during analysis.
type Diagnostic struct {
	HasOSGi          bool // OSGi bundles detected
	HasLiberty       bool // Liberty runtime detected (e.g., server.xml)
	AnyManifestFound bool // Any MANIFEST.MF found (OSGi or otherwise)
	HasLibertyWAR    bool // Liberty WAR application modeled as a single service (Phase F2)
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
	for i := range entryPoints {
		parts := strings.Split(entryPoints[i].Handler, ".")
		if len(parts) > 0 {
			entryPoints[i].Resource = parts[0]
		}
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

		// Group REST Resources
		svc.RESTResources = groupRESTResources(svc.EntryPoints)

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
			svc.RESTResources = groupRESTResources(svc.EntryPoints)
			services = append(services, svc)
		}
	}

	// 5. Build System Graph
	sysGraph := graph.BuildSystemGraph(services)

	return services, sysGraph, diag
}

func groupRESTResources(eps []model.EntryPoint) []model.RESTResource {
	groups := make(map[string][]model.EntryPoint)
	for _, ep := range eps {
		groups[ep.Resource] = append(groups[ep.Resource], ep)
	}

	var resources []model.RESTResource
	for name, groupEps := range groups {
		sourceFile := groupEps[0].SourceFile
		meta := scanResourceMetadata(sourceFile, name)

		res := model.RESTResource{
			Name:            name,
			SourceFile:      sourceFile,
			BasePath:        meta.basePath,
			AuthAnnotations: meta.auth,
			Consumes:        meta.consumes,
			Produces:        meta.produces,
			Methods:         make([]model.RESTMethod, 0),
			HTTPMethods:     make(map[string]int),
			EntryPoints:     groupEps,
		}

		// Phase F3.3: Correct SubPath and FullPath computation
		paramMap := make(map[string]bool)
		for _, ep := range groupEps {
			subPath := ep.Path
			if meta.basePath != "" && strings.HasPrefix(ep.Path, meta.basePath) {
				subPath = strings.TrimPrefix(ep.Path, meta.basePath)
			}

			full := joinPaths(meta.basePath, subPath)
			method := model.RESTMethod{
				HTTPMethod: ep.Method,
				SubPath:    normalizePath(subPath),
				FullPath:   full,
				Handler:    ep.Handler,
				SourceFile: ep.SourceFile,
			}
			res.Methods = append(res.Methods, method)
			res.HTTPMethods[ep.Method]++

			// Extract path params
			for _, p := range extractPathParams(full) {
				paramMap[p] = true
			}
		}

		for p := range paramMap {
			res.PathParams = append(res.PathParams, p)
		}

		resources = append(resources, res)
	}
	return resources
}

type resourceMeta struct {
	basePath string
	auth     []string
	consumes []string
	produces []string
}

// scanResourceMetadata performs a lightweight scan of a Java file to find JAX-RS metadata.
// This is an AST-lite scan: it collects annotations before the class declaration
// and aggregates method-level annotations found throughout the file.
func scanResourceMetadata(javaFilePath string, className string) resourceMeta {
	f, err := os.Open(javaFilePath)
	if err != nil {
		return resourceMeta{}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var meta resourceMeta
	var latestPath string
	var classFound bool

	authMap := make(map[string]bool)
	consumesMap := make(map[string]bool)
	producesMap := make(map[string]bool)

	authPrefixes := []string{"@RolesAllowed", "@PermitAll", "@DenyAll", "@Authenticated", "@RequiresRole", "@Secured"}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 1. Detect Path
		if strings.HasPrefix(line, "@Path") {
			lp := extractAnnotationString(line)
			if !classFound {
				latestPath = lp
			}
		}

		// 2. Detect Auth
		for _, pref := range authPrefixes {
			if strings.HasPrefix(line, pref) {
				authMap[strings.TrimPrefix(pref, "@")] = true
			}
		}

		// 3. Detect Media Types
		if strings.HasPrefix(line, "@Consumes") {
			if mt := extractAnnotationString(line); mt != "" {
				consumesMap[mt] = true
			}
		}
		if strings.HasPrefix(line, "@Produces") {
			if mt := extractAnnotationString(line); mt != "" {
				producesMap[mt] = true
			}
		}

		// 4. Detect class declaration
		if !classFound && (strings.Contains(line, "class ") || strings.Contains(line, "interface ")) && strings.Contains(line, className) {
			parts := strings.Fields(line)
			for i, p := range parts {
				if (p == "class" || p == "interface") && i+1 < len(parts) {
					actualName := strings.Split(parts[i+1], "{")[0]
					if actualName == className {
						meta.basePath = normalizePath(latestPath)
						classFound = true
					}
				}
			}
		}
	}

	for k := range authMap {
		meta.auth = append(meta.auth, k)
	}
	for k := range consumesMap {
		meta.consumes = append(meta.consumes, k)
	}
	for k := range producesMap {
		meta.produces = append(meta.produces, k)
	}

	return meta
}

func extractAnnotationString(line string) string {
	start := strings.Index(line, "\"")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(line, "\"")
	if end > start {
		return line[start+1 : end]
	}
	return ""
}

var pathParamRegex = regexp.MustCompile(`\{([^}]+)\}`)

func extractPathParams(path string) []string {
	matches := pathParamRegex.FindAllStringSubmatch(path, -1)
	var params []string
	for _, m := range matches {
		if len(m) > 1 {
			params = append(params, m[1])
		}
	}
	return params
}

// normalizePath ensures a path starts with /, uses correct slashes, and has no duplicates.
// If the input is empty, it returns an empty string to preserve semantics.
func normalizePath(p string) string {
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	p = filepath.ToSlash(p)
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	if len(p) > 1 && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// joinPaths safely joins two path segments.
func joinPaths(base, sub string) string {
	if base == "" {
		return normalizePath(sub)
	}
	if sub == "" {
		return normalizePath(base)
	}
	return normalizePath(base + "/" + sub)
}
