package report

import (
	"fmt"
	"jz/app"
	"jz/model"
	"sort"
	"strings"
)

// GenerateMarkdown creates a human-readable Markdown report from the analysis results.
func GenerateMarkdown(services []model.Service, sysGraph model.SystemGraph, diag app.Diagnostic) string {
	var sb strings.Builder

	// 1. System Overview
	sb.WriteString("# System Overview\n\n")
	sb.WriteString(fmt.Sprintf("- Total number of services: %d\n", len(services)))
	sb.WriteString(fmt.Sprintf("- Total number of system-level dependencies: %d\n", len(sysGraph.Dependencies)))
	sb.WriteString("\n")

	// Diagnostics section
	if !diag.HasOSGi {
		sb.WriteString("## Diagnostics\n\n")
		if diag.HasLibertyWAR {
			sb.WriteString("- Liberty WAR service detected.\n")
			sb.WriteString("- OSGi bundles not found; modeled as a single Liberty service.\n")
		} else if diag.HasLiberty {
			if !diag.AnyManifestFound {
				sb.WriteString("- No MANIFEST.MF files found.\n")
			}
			sb.WriteString("- OSGi-based analysis skipped.\n")
			sb.WriteString("- server.xml was detected.\n")
			sb.WriteString("- Liberty-only service support is planned in a future release.\n")
		} else {
			sb.WriteString("- No supported runtime model detected.\n")
			sb.WriteString("- Supported models: OSGi bundles (via META-INF/MANIFEST.MF).\n")
			sb.WriteString("- Analysis skipped.\n")
		}
		sb.WriteString("\n")
	}

	// 2. Services
	sb.WriteString("# Services\n\n")
	for _, svc := range services {
		sb.WriteString(fmt.Sprintf("## %s\n\n", svc.Name))
		sb.WriteString(fmt.Sprintf("- Root Path: %s\n", svc.RootPath))
		sb.WriteString(fmt.Sprintf("- REST Entry Points: %d\n", len(svc.EntryPoints)))
		sb.WriteString(fmt.Sprintf("- DS Components: %d\n", len(svc.Components)))

		if svc.ServerName != "" {
			sb.WriteString(fmt.Sprintf("- Liberty Server: %s\n", svc.ServerName))
		}

		if len(svc.Features) > 0 {
			sb.WriteString("- Enabled Features:\n")
			for _, f := range svc.Features {
				sb.WriteString(fmt.Sprintf("  - %s\n", f))
			}
		}

		if len(svc.RESTResources) > 0 {
			sb.WriteString("### REST Resources\n\n")
			for _, res := range svc.RESTResources {
				sb.WriteString(fmt.Sprintf("#### %s\n", res.Name))
				if res.BasePath != "" {
					sb.WriteString(fmt.Sprintf("Base path: %s\n", res.BasePath))
				}
				if len(res.AuthAnnotations) > 0 {
					sb.WriteString(fmt.Sprintf("Auth: %s\n", strings.Join(res.AuthAnnotations, ", ")))
				}
				if len(res.Consumes) > 0 {
					sb.WriteString(fmt.Sprintf("Consumes: %s\n", strings.Join(res.Consumes, ", ")))
				}
				if len(res.Produces) > 0 {
					sb.WriteString(fmt.Sprintf("Produces: %s\n", strings.Join(res.Produces, ", ")))
				}
				if len(res.PathParams) > 0 {
					sb.WriteString(fmt.Sprintf("Path Params: %s\n", strings.Join(res.PathParams, ", ")))
				}
				sb.WriteString("\n")

				for _, m := range res.Methods {
					sb.WriteString(fmt.Sprintf("- %-7s %s\n", m.HTTPMethod, m.FullPath))
				}

				if len(res.HTTPMethods) > 0 {
					sb.WriteString("\nMethods summary:\n")
					var methods []string
					for m := range res.HTTPMethods {
						methods = append(methods, m)
					}
					sort.Strings(methods)
					for _, m := range methods {
						sb.WriteString(fmt.Sprintf("- %s: %d\n", m, res.HTTPMethods[m]))
					}
				}
				sb.WriteString("\n")
			}
		}

		sb.WriteString("\n")
	}

	// 3. REST Entry Points
	sb.WriteString("# REST Entry Points\n\n")
	for _, svc := range services {
		if len(svc.EntryPoints) > 0 {
			sb.WriteString(fmt.Sprintf("## %s\n\n", svc.Name))
			for _, ep := range svc.EntryPoints {
				sb.WriteString(fmt.Sprintf("- %s %s (%s)\n", ep.Method, ep.Path, ep.Handler))
			}
			sb.WriteString("\n")
		}
	}

	// 4. Internal Component Dependencies
	sb.WriteString("# Internal Component Dependencies\n\n")
	for _, svc := range services {
		sb.WriteString(fmt.Sprintf("## %s\n\n", svc.Name))
		if len(svc.InternalGraph.Edges) == 0 {
			sb.WriteString("No internal component dependencies.\n\n")
		} else {
			for _, edge := range svc.InternalGraph.Edges {
				sb.WriteString(fmt.Sprintf("- %s -> %s (%s)\n", edge.FromComponent, edge.ToComponent, edge.Interface))
			}
			sb.WriteString("\n")
		}
	}

	// 5. System-Level Dependencies
	sb.WriteString("# System-Level Dependencies\n\n")
	if len(sysGraph.Dependencies) == 0 {
		sb.WriteString("No system-level dependencies.\n")
	} else {
		for _, dep := range sysGraph.Dependencies {
			sb.WriteString(fmt.Sprintf("- %s -> %s (%s)\n", dep.FromService, dep.ToService, dep.Interface))
		}
	}

	return sb.String()
}
