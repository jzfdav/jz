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

	// 5. Entry-Point Comparison Summary Table
	sb.WriteString("## Comparison Summary\n\n")
	sb.WriteString("| HTTP Method + Path | Has Guards | Early Return | Outbound Calls |\n")
	sb.WriteString("| :--- | :---: | :---: | :---: |\n")
	for _, f := range flows {
		hasGuards := "No"
		hasEarlyReturn := "No"
		hasOutbound := "No"
		for i, s := range f.Steps {
			if s.Kind == model.FlowStepCondition {
				hasGuards = "Yes"
			}
			if s.Kind == model.FlowStepReturn && i < len(f.Steps)-1 {
				hasEarlyReturn = "Yes"
			}
			if s.Kind == model.FlowStepOutbound {
				hasOutbound = "Yes"
			}
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s |\n", f.EntryPoint, hasGuards, hasEarlyReturn, hasOutbound))
	}
	sb.WriteString("\n")

	// 4. Explicit "No Outbound Calls" Signal (Global)
	totalOutbound := 0
	for _, f := range flows {
		for _, s := range f.Steps {
			if s.Kind == model.FlowStepOutbound {
				totalOutbound++
			}
		}
	}
	if totalOutbound == 0 {
		sb.WriteString("> ℹ️ **Note:** No outbound REST calls detected in any analyzed handlers for this resource.\n\n")
	}

	sb.WriteString("## Summary\n")
	sb.WriteString(fmt.Sprintf("Extracted %d flow(s) for resource `%s`.\n\n", len(flows), resourceName))

	// Per-entry flow section
	for _, f := range flows {
		sb.WriteString(fmt.Sprintf("## Flow: %s\n\n", f.EntryPoint))

		if len(f.Steps) == 0 {
			sb.WriteString("_No steps detected (empty handler or failed to parse)._\n\n")
			continue
		}

		// 4. Explicit "No Outbound Calls" Signal (Per-flow)
		flowOutbound := false
		for _, s := range f.Steps {
			if s.Kind == model.FlowStepOutbound {
				flowOutbound = true
				break
			}
		}

		// 2. Logical Grouping of Steps (Visual Headers)
		currentGroup := ""

		for _, s := range f.Steps {
			newGroup := ""
			switch s.Kind {
			case model.FlowStepEntry:
				newGroup = "Entry"
			case model.FlowStepCondition:
				newGroup = "Guard Conditions"
			case model.FlowStepOutbound:
				newGroup = "Outbound Calls"
			case model.FlowStepCall:
				newGroup = "Method Execution"
			case model.FlowStepReturn:
				newGroup = "Early Exit / Return"
			case model.FlowStepUnexpanded:
				newGroup = "Scope Limits"
			}

			if newGroup != "" && newGroup != currentGroup {
				sb.WriteString(fmt.Sprintf("### %s\n\n", newGroup))
				currentGroup = newGroup
			}

			kindLabel := strings.ToUpper(string(s.Kind))
			description := s.Description
			// 3. Gating vs Core Path Distinction
			if s.Kind == model.FlowStepCondition {
				description = "**Guard:** " + description
			}

			sb.WriteString(fmt.Sprintf("%d. **%s**: %s\n", s.Index, kindLabel, description))

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
			sb.WriteString("\n")
		}

		if !flowOutbound {
			sb.WriteString("_No outbound REST calls detected in this handler._\n\n")
		}

		// 1. Flow Completion Signaling
		lastStep := f.Steps[len(f.Steps)-1]
		if lastStep.Kind != model.FlowStepReturn {
			sb.WriteString("> 💡 **End Note:** Flow may continue into helper services, injected components, or external layers not expanded in this view (reached end of analyzed handler without explicit return).\n\n")
		} else {
			sb.WriteString("> ✅ **End Note:** Flow completed with a detected return statement.\n\n")
		}
	}

	// Observations Section (Refined)
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
