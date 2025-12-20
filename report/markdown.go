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

				// Phase F4: Inbound Calls
				if len(res.InboundCalls) > 0 {
					sb.WriteString("\n#### Inbound Calls\n\n")
					calls := res.InboundCalls
					sortRESTCalls(calls)
					for _, call := range calls {
						sb.WriteString(fmt.Sprintf("- FROM %s/%s.%s\n", call.FromService, call.FromResource, call.FromHandler))
						sb.WriteString(fmt.Sprintf("  TO %s/%s\n", svc.Name, res.Name))
						sb.WriteString(fmt.Sprintf("  %s %s\n", call.HTTPMethod, call.TargetPath))
						sb.WriteString(fmt.Sprintf("  Confidence: %s | Detection: %s\n", call.Confidence, call.DetectionType))
						sb.WriteString(fmt.Sprintf("  File: %s\n", call.SourceFile))
					}
				}

				// Phase F4: Outbound Calls
				if len(res.OutboundCalls) > 0 {
					sb.WriteString("\n#### Outbound Calls\n\n")
					calls := res.OutboundCalls
					sortRESTCalls(calls)
					for _, call := range calls {
						target := "UNRESOLVED"
						if call.TargetService != "" {
							target = fmt.Sprintf("%s/%s", call.TargetService, call.TargetResource)
						}
						sb.WriteString(fmt.Sprintf("- FROM %s/%s.%s\n", call.FromService, call.FromResource, call.FromHandler))
						sb.WriteString(fmt.Sprintf("  TO %s\n", target))
						sb.WriteString(fmt.Sprintf("  %s %s\n", call.HTTPMethod, call.TargetPath))
						sb.WriteString(fmt.Sprintf("  Confidence: %s | Detection: %s\n", call.Confidence, call.DetectionType))
						sb.WriteString(fmt.Sprintf("  File: %s\n", call.SourceFile))
					}
				}

				sb.WriteString("\n")
			}
		}

		// Phase F4: Service Boundaries
		if len(svc.Boundaries) > 0 {
			sb.WriteString("### Detected Service Boundaries\n\n")
			boundaries := svc.Boundaries
			sort.Slice(boundaries, func(i, j int) bool {
				if boundaries[i].BoundaryType != boundaries[j].BoundaryType {
					return boundaries[i].BoundaryType < boundaries[j].BoundaryType
				}
				return boundaries[i].Identifier < boundaries[j].Identifier
			})
			for _, b := range boundaries {
				sb.WriteString(fmt.Sprintf("- **%s**: %s\n", b.BoundaryType, b.Identifier))
				sb.WriteString(fmt.Sprintf("  - Evidence: %s\n", b.Evidence))
			}
			sb.WriteString("\n")
		}

		// Phase F4: REST Call Summary
		if len(svc.RESTCalls) > 0 {
			sb.WriteString("### REST Call Summary\n\n")
			resolved := 0
			confCounts := make(map[string]int)
			detCounts := make(map[string]int)
			paths := make(map[string]bool)

			for _, call := range svc.RESTCalls {
				if call.TargetService != "" {
					resolved++
				}
				confCounts[call.Confidence]++
				detCounts[call.DetectionType]++
				paths[call.TargetPath] = true
			}

			sb.WriteString(fmt.Sprintf("- Total outbound calls: %d\n", len(svc.RESTCalls)))
			sb.WriteString(fmt.Sprintf("- Resolved within service: %d\n", resolved))
			sb.WriteString(fmt.Sprintf("- Unresolved calls: %d\n", len(svc.RESTCalls)-resolved))
			sb.WriteString(fmt.Sprintf("- Distinct target paths: %d\n", len(paths)))

			sb.WriteString("- Confidence breakdown:\n")
			for _, c := range []string{model.ConfidenceHigh, model.ConfidenceMedium, model.ConfidenceLow} {
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", c, confCounts[c]))
			}

			sb.WriteString("- Detection breakdown:\n")
			for _, d := range []string{model.DetectionLiteral, model.DetectionConstant, model.DetectionUnknown} {
				sb.WriteString(fmt.Sprintf("  - %s: %d\n", d, detCounts[d]))
			}

			if resolved < len(svc.RESTCalls) {
				sb.WriteString("\n*Note: This tool performs AST-lite static analysis. Calls may remain unresolved due to dynamic URL construction, variables, or cross-service boundaries not yet modeled.*\n")
			}
			sb.WriteString("\n")
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

func sortRESTCalls(calls []model.RESTCall) {
	sort.Slice(calls, func(i, j int) bool {
		vi := confidenceRank(calls[i].Confidence)
		vj := confidenceRank(calls[j].Confidence)
		if vi != vj {
			return vi > vj
		}
		if calls[i].HTTPMethod != calls[j].HTTPMethod {
			return calls[i].HTTPMethod < calls[j].HTTPMethod
		}
		return calls[i].TargetPath < calls[j].TargetPath
	})
}

// confidenceRank returns a numeric rank for a confidence string to simplify sorting.
func confidenceRank(c string) int {
	switch c {
	case model.ConfidenceHigh:
		return 3
	case model.ConfidenceMedium:
		return 2
	case model.ConfidenceLow:
		return 1
	default:
		return 0
	}
}
