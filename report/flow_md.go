package report

import (
	"fmt"
	"jz/model"
	"strings"
)

// GenerateFlowMarkdown produces a narrative execution flow report in Markdown.
func GenerateFlowMarkdown(flows []model.ExecutionFlow, resourceName string, pathFilter string) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("# Execution Flow: %s\n\n", resourceName))
	sb.WriteString("> **Analysis Mode:** AST-lite (Conservative)\n")
	sb.WriteString("> **Scope:** Single Resource Targeted Extraction\n")
	if pathFilter != "" {
		sb.WriteString(fmt.Sprintf("> **Filter:** Path contains '%s'\n", pathFilter))
	}
	sb.WriteString("\n")

	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("Extracted %d flow(s) for resource `%s`.\n\n", len(flows), resourceName))

	// Entry Points List
	sb.WriteString("### Entry Points\n")
	for _, f := range flows {
		sb.WriteString(fmt.Sprintf("- `%s`\n", f.EntryPoint))
	}
	sb.WriteString("\n")

	// Per-entry flow section
	for _, f := range flows {
		sb.WriteString(fmt.Sprintf("## Flow: %s\n\n", f.EntryPoint))

		if len(f.Steps) == 0 {
			sb.WriteString("_No steps detected (empty handler or failed to parse)._\n\n")
			continue
		}

		for _, s := range f.Steps {
			kindLabel := strings.ToUpper(string(s.Kind))
			sb.WriteString(fmt.Sprintf("%d. **%s**: %s\n", s.Index, kindLabel, s.Description))

			if s.ToMethod != "" {
				sb.WriteString(fmt.Sprintf("   - **Target:** `%s`", s.ToMethod))
				if s.ResolutionScope != "" {
					sb.WriteString(fmt.Sprintf(" (%s)", s.ResolutionScope))
				}
				sb.WriteString("\n")
			}

			sb.WriteString(fmt.Sprintf("   - **Evidence:** `%s` [confidence: %s]\n", s.Evidence, s.Confidence))

			if s.ResolutionScope == model.ResolutionUnresolved && s.Kind == model.FlowStepOutbound {
				sb.WriteString("   - ⚠️ *Note: This outbound call could not be resolved to a known resource.*\n")
			}
		}
		sb.WriteString("\n")
	}

	// Observations Section
	sb.WriteString("## Observations\n\n")

	// Gating conditions
	gatingFound := false
	sb.WriteString("### Gating & Guardrails\n")
	for _, f := range flows {
		for _, s := range f.Steps {
			if s.Kind == model.FlowStepCondition {
				sb.WriteString(fmt.Sprintf("- Flow `%s` is gated by: `%s`\n", f.EntryPoint, s.Description))
				gatingFound = true
			}
		}
	}
	if !gatingFound {
		sb.WriteString("- No explicit gating conditions detected.\n")
	}
	sb.WriteString("\n")

	// Early exits
	exitFound := false
	sb.WriteString("### Early Exits\n")
	for _, f := range flows {
		for i, s := range f.Steps {
			if s.Kind == model.FlowStepReturn && i < len(f.Steps)-1 {
				sb.WriteString(fmt.Sprintf("- Flow `%s` has an early exit: `%s`\n", f.EntryPoint, s.Description))
				exitFound = true
			}
		}
	}
	if !exitFound {
		sb.WriteString("- No early exits detected.\n")
	}
	sb.WriteString("\n")

	// Limitations Note
	sb.WriteString("## Limitations (AST-lite)\n")
	sb.WriteString("- Logic is extracted via line-based lexical analysis.\n")
	sb.WriteString("- Data propagation across variables or loops is not tracked.\n")
	sb.WriteString("- Complex boolean expressions may be truncated.\n")
	sb.WriteString("- Only same-file internal methods are expanded.\n")

	return sb.String()
}
