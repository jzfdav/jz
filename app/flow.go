package app

import (
	"bufio"
	"fmt"
	"jz/model"
	"os"
	"strings"
)

// ExtractFlow coordinates the extraction of execution flows for a specific resource.
func ExtractFlow(services []model.Service, resourceName string, methodFilter string, pathFilter string, maxDepth int) ([]model.ExecutionFlow, error) {
	var targetRes *model.RESTResource
	var targetSvc *model.Service

	// 1. Locate the target resource
	for i := range services {
		for j := range services[i].RESTResources {
			if services[i].RESTResources[j].Name == resourceName {
				targetRes = &services[i].RESTResources[j]
				targetSvc = &services[i]
				break
			}
		}
		if targetRes != nil {
			break
		}
	}

	if targetRes == nil {
		return nil, fmt.Errorf("resource '%s' not found", resourceName)
	}

	var flows []model.ExecutionFlow

	// 2. Process each entry point (Method)
	for _, m := range targetRes.Methods {
		// Apply filters
		if methodFilter != "" && !strings.EqualFold(m.HTTPMethod, methodFilter) {
			continue
		}
		if pathFilter != "" && pathFilter != "*" && !strings.Contains(m.FullPath, pathFilter) {
			continue
		}

		flow := model.ExecutionFlow{
			ResourceName: resourceName,
			EntryPoint:   fmt.Sprintf("%s %s", m.HTTPMethod, m.FullPath),
		}

		parts := strings.Split(m.Handler, ".")
		if len(parts) < 2 {
			continue
		}
		handlerMethod := parts[1]

		visited := make(map[string]bool)
		flow.Steps = scanMethodFlow(m.SourceFile, targetRes.Name, handlerMethod, targetSvc, 0, maxDepth, visited)

		// Re-index steps
		for i := range flow.Steps {
			flow.Steps[i].Index = i + 1
		}

		flows = append(flows, flow)
	}

	return flows, nil
}

func scanMethodFlow(sourceFile, className, methodName string, service *model.Service, depth, maxDepth int, visited map[string]bool) []model.FlowStep {
	fullHandler := fmt.Sprintf("%s.%s", className, methodName)
	visited[fullHandler] = true

	f, err := os.Open(sourceFile)
	if err != nil {
		return nil
	}
	defer f.Close()

	var steps []model.FlowStep
	scanner := bufio.NewScanner(f)

	inMethod := false
	braceCount := 0
	lineNum := 0

	// Entry Step
	steps = append(steps, model.FlowStep{
		Kind:        model.FlowStepEntry,
		Description: fmt.Sprintf("Enter: %s", methodName),
		FromMethod:  fullHandler,
		Confidence:  model.ConfidenceHigh,
		Evidence:    fmt.Sprintf("%s (start)", sourceFile),
	})

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Detect start of method
		if !inMethod && strings.Contains(line, methodName) && strings.Contains(line, "(") && (strings.Contains(line, "public") || strings.Contains(line, "private") || strings.Contains(line, "protected") || strings.Contains(line, "void ")) {
			// Basic check to ensure it's a method declaration, not a call
			if !strings.HasSuffix(trimmed, ";") {
				inMethod = true
				braceCount = strings.Count(line, "{") - strings.Count(line, "}")
				continue
			}
		}

		if inMethod {
			braceCount += strings.Count(line, "{") - strings.Count(line, "}")
			if braceCount <= 0 && strings.Contains(line, "}") {
				inMethod = false
				break
			}

			// 1. Detect Conditionals
			if strings.HasPrefix(trimmed, "if") || strings.HasPrefix(trimmed, "} else if") {
				cond := extractCondition(trimmed)
				steps = append(steps, model.FlowStep{
					Kind:        model.FlowStepCondition,
					Description: fmt.Sprintf("Check: %s", cond),
					FromMethod:  fullHandler,
					Confidence:  model.ConfidenceMedium,
					Evidence:    fmt.Sprintf("%s:%d", sourceFile, lineNum),
				})
			} else if strings.HasPrefix(trimmed, "else") || strings.HasPrefix(trimmed, "} else") {
				steps = append(steps, model.FlowStep{
					Kind:        model.FlowStepCondition,
					Description: "Otherwise",
					FromMethod:  fullHandler,
					Confidence:  model.ConfidenceMedium,
					Evidence:    fmt.Sprintf("%s:%d", sourceFile, lineNum),
				})
			}

			// 2. Detect Returns
			if strings.HasPrefix(trimmed, "return") {
				steps = append(steps, model.FlowStep{
					Kind:        model.FlowStepReturn,
					Description: fmt.Sprintf("Return: %s", strings.TrimSuffix(strings.TrimPrefix(trimmed, "return "), ";")),
					FromMethod:  fullHandler,
					Confidence:  model.ConfidenceHigh,
					Evidence:    fmt.Sprintf("%s:%d", sourceFile, lineNum),
				})
			}

			// 3. Detect Outbound REST Calls (reuse patterns from scanOutboundCalls)
			outboundPatterns := []string{"get(", "post(", "put(", "delete(", "RESTClient", "WebTarget", "HttpURLConnection"}
			isOutbound := false
			for _, p := range outboundPatterns {
				if strings.Contains(trimmed, p) {
					isOutbound = true
					call := model.RESTCall{
						FromHandler: methodName,
						SourceFile:  sourceFile,
						Confidence:  model.ConfidenceLow,
					}
					// Try to find HTTP method
					upper := strings.ToUpper(trimmed)
					for _, m := range []string{"GET", "POST", "PUT", "DELETE"} {
						if strings.Contains(upper, m) {
							call.HTTPMethod = m
							break
						}
					}
					// Try to find target path
					if strings.Contains(trimmed, "\"") {
						start := strings.Index(trimmed, "\"")
						end := strings.LastIndex(trimmed, "\"")
						if end > start {
							path := trimmed[start+1 : end]
							if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "http") {
								call.TargetPath = path
								call.Confidence = model.ConfidenceHigh
							}
						}
					}

					desc := "Outbound REST call"
					if call.HTTPMethod != "" && call.TargetPath != "" {
						desc = fmt.Sprintf("Call: %s %s", call.HTTPMethod, call.TargetPath)
					}

					// Check if this call was resolved in Phase F5
					resScope := model.ResolutionUnresolved
					resolvedTo := ""
					for _, svcCall := range service.RESTCalls {
						if svcCall.FromHandler == methodName && svcCall.HTTPMethod == call.HTTPMethod && svcCall.TargetPath == call.TargetPath {
							resScope = svcCall.ResolutionScope
							if svcCall.TargetResource != "" {
								resolvedTo = svcCall.TargetService + " -> " + svcCall.TargetResource
							}
							break
						}
					}

					steps = append(steps, model.FlowStep{
						Kind:            model.FlowStepOutbound,
						Description:     desc,
						FromMethod:      fullHandler,
						ToMethod:        resolvedTo,
						Confidence:      call.Confidence,
						Evidence:        fmt.Sprintf("%s:%d", sourceFile, lineNum),
						ResolutionScope: resScope,
					})
					break
				}
			}

			// 4. Detect Internal Method Calls (same class expansion)
			if !isOutbound && strings.Contains(trimmed, "(") && !strings.Contains(trimmed, "new ") && !strings.Contains(trimmed, "return ") && !strings.HasPrefix(trimmed, "if") && !strings.HasPrefix(trimmed, "for") && !strings.HasPrefix(trimmed, "while") {
				innerMethod := extractMethodName(trimmed)
				if innerMethod != "" && innerMethod != methodName {
					// Check if method exists in the same file (simplistic check)
					if methodExistsInFile(sourceFile, innerMethod) {
						targetHandler := fmt.Sprintf("%s.%s", className, innerMethod)
						if depth < maxDepth && !visited[targetHandler] {
							steps = append(steps, model.FlowStep{
								Kind:        model.FlowStepCall,
								Description: fmt.Sprintf("Call internal: %s", innerMethod),
								FromMethod:  fullHandler,
								ToMethod:    targetHandler,
								Confidence:  model.ConfidenceMedium,
								Evidence:    fmt.Sprintf("%s:%d", sourceFile, lineNum),
							})

							// Recurse
							innerSteps := scanMethodFlow(sourceFile, className, innerMethod, service, depth+1, maxDepth, visited)
							steps = append(steps, innerSteps...)
						} else {
							reason := "depth limit"
							if visited[targetHandler] {
								reason = "already visited / potential cycle"
							}
							steps = append(steps, model.FlowStep{
								Kind:        model.FlowStepUnexpanded,
								Description: fmt.Sprintf("Call internal: %s (unexpanded - %s)", innerMethod, reason),
								FromMethod:  fullHandler,
								ToMethod:    targetHandler,
								Confidence:  model.ConfidenceHigh,
								Evidence:    fmt.Sprintf("%s:%d", sourceFile, lineNum),
							})
						}
					}
				}
			}
		}
	}

	return steps
}

func extractCondition(line string) string {
	start := strings.Index(line, "(")
	end := strings.LastIndex(line, ")")
	if start != -1 && end != -1 && end > start {
		return line[start+1 : end]
	}
	return "unknown condition"
}

func extractMethodName(line string) string {
	// Very basic extraction: search for 'something('
	idx := strings.Index(line, "(")
	if idx <= 0 {
		return ""
	}

	// Scan backwards for the start of the identifier
	start := idx - 1
	for start >= 0 {
		c := line[start]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			break
		}
		start--
	}

	method := line[start+1 : idx]
	// Avoid keywords
	keywords := map[string]bool{"if": true, "for": true, "while": true, "switch": true, "catch": true, "synchronized": true, "super": true, "this": true}
	if keywords[method] {
		return ""
	}
	return method
}

func methodExistsInFile(filePath, methodName string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, methodName) && strings.Contains(line, "(") && (strings.Contains(line, "public") || strings.Contains(line, "private") || strings.Contains(line, "protected") || strings.Contains(line, "void ")) {
			if !strings.HasSuffix(strings.TrimSpace(line), ";") {
				return true
			}
		}
	}
	return false
}
