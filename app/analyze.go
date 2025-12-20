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
	"sort"
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

		// Phase F4: Detect Outbound Calls
		// Deduplicate outbound REST calls within a single service
		callMap := make(map[string]bool)
		for _, res := range svc.RESTResources {
			for _, ep := range res.EntryPoints {
				parts := strings.Split(ep.Handler, ".")
				if len(parts) > 1 {
					methodName := parts[1]
					calls := scanOutboundCalls(ep.SourceFile, methodName, svc.Name, res.Name)
					for _, call := range calls {
						key := restCallKey(methodName, call)
						if !callMap[key] {
							svc.RESTCalls = append(svc.RESTCalls, call)
							callMap[key] = true
						}
					}
				}
			}
		}

		// Phase F4: Boundary Detection (simplistic package-based)
		pkgMap := make(map[string]bool)
		for _, res := range svc.RESTResources {
			parts := strings.Split(res.Name, ".")
			if len(parts) > 1 {
				pkg := strings.Join(parts[:len(parts)-1], ".")
				pkgMap[pkg] = true
			}
		}
		for pkg := range pkgMap {
			svc.Boundaries = append(svc.Boundaries, model.ServiceBoundary{
				ServiceName:  svc.Name,
				BoundaryType: "package",
				Identifier:   pkg,
				Evidence:     "Resource found in this package",
			})
		}
		sort.Slice(svc.Boundaries, func(i, j int) bool {
			return svc.Boundaries[i].Identifier < svc.Boundaries[j].Identifier
		})

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

			// Phase F4: Detect Outbound Calls
			// Deduplicate outbound REST calls within a single service
			callMap := make(map[string]bool)
			for _, res := range svc.RESTResources {
				for _, ep := range res.EntryPoints {
					parts := strings.Split(ep.Handler, ".")
					if len(parts) > 1 {
						methodName := parts[1]
						calls := scanOutboundCalls(ep.SourceFile, methodName, svc.Name, res.Name)
						for _, call := range calls {
							key := restCallKey(methodName, call)
							if !callMap[key] {
								svc.RESTCalls = append(svc.RESTCalls, call)
								callMap[key] = true
							}
						}
					}
				}
			}

			// Phase F4: Boundary Detection
			svc.Boundaries = append(svc.Boundaries, model.ServiceBoundary{
				ServiceName:  svc.Name,
				BoundaryType: "resource-group",
				Identifier:   "rest-api",
				Evidence:     "Liberty WAR modeled as a single REST resource group",
			})

			services = append(services, svc)
		}
	}

	// 5. Link Calls and Deterministic Sorting
	linkCallsToResources(services)

	// 6. Build System Graph
	sysGraph := graph.BuildSystemGraph(services)

	return services, sysGraph, diag
}

func groupRESTResources(eps []model.EntryPoint) []model.RESTResource {
	groups := make(map[string][]model.EntryPoint)
	for _, ep := range eps {
		groups[ep.Resource] = append(groups[ep.Resource], ep)
	}

	var names []string
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)

	var resources []model.RESTResource
	for _, name := range names {
		groupEps := groups[name]
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
		sort.Strings(res.PathParams)

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
//
// Limitations (AST-lite):
// - Line-based scanning: Does not understand multi-line statements or complex expressions.
// - No variable resolution: Only literal values are extracted.
// - No control-flow analysis: Cannot determine if code is reachable.
// - No constant evaluation: Constants (e.g., MediaType.APPLICATION_JSON) are not resolved.
// - False negatives preferred: Items are skipped if parsing is ambiguous (favors safety over completeness).
// - Media-type parsing (@Consumes, @Produces) only supports literal string values.
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
				authMap[pref] = true
			}
		}

		// 3. Detect Media Types
		if strings.HasPrefix(line, "@Consumes") {
			if mt := extractAnnotationString(line); mt != "" {
				consumesMap[strings.ToLower(mt)] = true
			}
		}
		if strings.HasPrefix(line, "@Produces") {
			if mt := extractAnnotationString(line); mt != "" {
				producesMap[strings.ToLower(mt)] = true
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
	sort.Strings(meta.auth)

	for k := range consumesMap {
		meta.consumes = append(meta.consumes, k)
	}
	sort.Strings(meta.consumes)

	for k := range producesMap {
		meta.produces = append(meta.produces, k)
	}
	sort.Strings(meta.produces)

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

// scanOutboundCalls performs an AST-lite scan of a Java method to find outbound REST calls.
//
// Limitations (AST-lite):
// - Line-based scanning: Scans within detected method boundaries.
// - No variable resolution: Target URLs must be string literals.
// - No control-flow analysis: All detected calls are recorded regardless of execution path.
// - No constant evaluation: Parameterized or constant-based URLs are skipped (Confidence: Low).
// - False negatives preferred: Ambiguous or complex call patterns are intentionally ignored.
func scanOutboundCalls(sourceFile, methodName, fromService, fromResource string) []model.RESTCall {
	f, err := os.Open(sourceFile)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var calls []model.RESTCall
	var inMethod bool
	var braceCount int

	methodPatterns := []string{"get(", "post(", "put(", "delete(", "RESTClient", "WebTarget", "HttpURLConnection"}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect start of method
		if !inMethod && strings.Contains(line, methodName) && strings.Contains(line, "(") && (strings.Contains(line, "public") || strings.Contains(line, "private") || strings.Contains(line, "protected")) {
			inMethod = true
			braceCount = strings.Count(line, "{") - strings.Count(line, "}")
			continue
		}

		if inMethod {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount <= 0 && strings.Contains(line, "}") {
				inMethod = false
				continue
			}

			// Scan for call patterns
			for _, p := range methodPatterns {
				if strings.Contains(trimmed, p) {
					call := model.RESTCall{
						FromService:   fromService,
						FromResource:  fromResource,
						FromHandler:   methodName,
						SourceFile:    sourceFile,
						DetectionType: model.DetectionUnknown,
						Confidence:    model.ConfidenceLow,
					}

					// Try to find HTTP method
					upper := strings.ToUpper(trimmed)
					for _, m := range []string{"GET", "POST", "PUT", "DELETE"} {
						if strings.Contains(upper, m) {
							call.HTTPMethod = m
							break
						}
					}

					// Try to find target path (very simplistic)
					if strings.Contains(trimmed, "\"") {
						start := strings.Index(trimmed, "\"")
						end := strings.LastIndex(trimmed, "\"")
						if end > start {
							path := trimmed[start+1 : end]
							if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "http") {
								call.TargetPath = path
								call.DetectionType = model.DetectionLiteral
								call.Confidence = model.ConfidenceHigh
							}
						}
					}

					if call.TargetPath == "" {
						call.DetectionType = model.DetectionUnknown
						call.Confidence = model.ConfidenceLow
					}

					calls = append(calls, call)
					break // Only one call per line for now
				}
			}
		}
	}
	return calls
}

type targetResource struct {
	serviceName  string
	resourceName string
}

// linkCallsToResources attempts to link detected calls to known resources.
// It prioritizes same-service links, then attempts cross-service resolution
// if a unique match exists globally (AST-lite conservative matching).
func linkCallsToResources(services []model.Service) {
	// 1. Build a global registry of all REST entry points: (Method, Path) -> []targetResource
	registry := make(map[string][]targetResource)
	for _, svc := range services {
		for _, res := range svc.RESTResources {
			for _, m := range res.Methods {
				key := fmt.Sprintf("%s|%s", m.HTTPMethod, m.FullPath)
				registry[key] = append(registry[key], targetResource{
					serviceName:  svc.Name,
					resourceName: res.Name,
				})
			}
		}
	}

	for i := range services {
		// 2. Build map of paths for the same service for fast priority lookup
		sameServiceMap := make(map[string][]string) // path -> []resourceName
		for _, res := range services[i].RESTResources {
			for _, m := range res.Methods {
				key := fmt.Sprintf("%s|%s", m.HTTPMethod, m.FullPath)
				sameServiceMap[key] = append(sameServiceMap[key], res.Name)
			}
		}

		for j := range services[i].RESTCalls {
			call := &services[i].RESTCalls[j]
			call.ResolutionScope = model.ResolutionUnresolved

			// Key for lookup (Method + Path)
			regKey := fmt.Sprintf("%s|%s", call.HTTPMethod, call.TargetPath)

			// 2a. Attempt same-service resolution (Priority 1)
			if targets, ok := sameServiceMap[regKey]; ok {
				// Special Case: If same-service has multiple matches for the same path/method,
				// we still link if the resolution is unambiguous within the service.
				// However, Phase F5 usually assumes unique paths.
				if len(targets) == 1 {
					call.TargetService = services[i].Name
					call.TargetResource = targets[0]
					call.ResolutionScope = model.ResolutionSameService
					call.ResolutionEvidence = "exact path+method match (internal)"
				}
			}

			// 2b. Attempt cross-service resolution (Priority 2)
			// Only if unresolved, and confidence is High/Medium
			if call.ResolutionScope == model.ResolutionUnresolved &&
				(call.Confidence == model.ConfidenceHigh || call.Confidence == model.ConfidenceMedium) {

				if globalTargets, ok := registry[regKey]; ok {
					if len(globalTargets) == 1 {
						// Unique global match found
						call.TargetService = globalTargets[0].serviceName
						call.TargetResource = globalTargets[0].resourceName
						call.ResolutionScope = model.ResolutionCrossService
						call.ResolutionEvidence = "exact path+method match (global)"
					}
				}
			}

			// 3. Populate InboundCalls on the TargetResource (wherever it is)
			if call.TargetService != "" && call.TargetResource != "" {
				for sIdx := range services {
					if services[sIdx].Name == call.TargetService {
						for rIdx := range services[sIdx].RESTResources {
							if services[sIdx].RESTResources[rIdx].Name == call.TargetResource {
								services[sIdx].RESTResources[rIdx].InboundCalls = append(services[sIdx].RESTResources[rIdx].InboundCalls, *call)
							}
						}
					}
				}
			}

			// 4. Populate OutboundCalls on the Originating Resource
			for k := range services[i].RESTResources {
				if services[i].RESTResources[k].Name == call.FromResource {
					services[i].RESTResources[k].OutboundCalls = append(services[i].RESTResources[k].OutboundCalls, *call)
				}
			}
		}

		// Sort calls for determinism
		sort.Slice(services[i].RESTCalls, func(a, b int) bool {
			if services[i].RESTCalls[a].FromHandler != services[i].RESTCalls[b].FromHandler {
				return services[i].RESTCalls[a].FromHandler < services[i].RESTCalls[b].FromHandler
			}
			return services[i].RESTCalls[a].TargetPath < services[i].RESTCalls[b].TargetPath
		})

		for k := range services[i].RESTResources {
			sort.Slice(services[i].RESTResources[k].OutboundCalls, func(a, b int) bool {
				if services[i].RESTResources[k].OutboundCalls[a].FromHandler != services[i].RESTResources[k].OutboundCalls[b].FromHandler {
					return services[i].RESTResources[k].OutboundCalls[a].FromHandler < services[i].RESTResources[k].OutboundCalls[b].FromHandler
				}
				return services[i].RESTResources[k].OutboundCalls[a].TargetPath < services[i].RESTResources[k].OutboundCalls[b].TargetPath
			})
			sort.Slice(services[i].RESTResources[k].InboundCalls, func(a, b int) bool {
				if services[i].RESTResources[k].InboundCalls[a].FromHandler != services[i].RESTResources[k].InboundCalls[b].FromHandler {
					return services[i].RESTResources[k].InboundCalls[a].FromHandler < services[i].RESTResources[k].InboundCalls[b].FromHandler
				}
				return services[i].RESTResources[k].InboundCalls[a].TargetPath < services[i].RESTResources[k].InboundCalls[b].TargetPath
			})
		}
	}
}

// restCallKey generates a unique key for deduplicating outbound calls within a service.
func restCallKey(methodName string, call model.RESTCall) string {
	return fmt.Sprintf("%s|%s|%s", methodName, call.HTTPMethod, call.TargetPath)
}
