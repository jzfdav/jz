package report

import (
	"fmt"
	"jz/model"
	"strings"
)

// GenerateFlowDiffMarkdown produces a markdown report for flow diffs.
func GenerateFlowDiffMarkdown(diffs []model.FlowDiff, resourceName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Flow Diff: %s\n\n", resourceName))
	sb.WriteString("> **Analysis Mode:** Structural Execution-Flow Diff\n")
	sb.WriteString("> **Comparison:** Ordered Step-by-Step\n\n")

	for _, d := range diffs {
		sb.WriteString(fmt.Sprintf("## Flow: %s\n", d.EntryPoint))
		sb.WriteString(fmt.Sprintf("Status: **%s**\n\n", d.Status))

		if d.Status == "UNCHANGED" {
			sb.WriteString("No structural differences detected between flows.\n\n")
			continue
		}

		// 2. Guards
		renderDiffSection(&sb, "Guards", d.StepDiffs, model.FlowStepCondition)

		// 3. Outbound Calls
		renderDiffSection(&sb, "Outbound Calls", d.StepDiffs, model.FlowStepOutbound)

		// 4. Termination
		renderDiffSection(&sb, "Termination", d.StepDiffs, model.FlowStepReturn, model.FlowStepUnexpanded)

		// 5. Unchanged Summary (if not already mostly covered)
		// We could add a full trace or just summarize changes.
		// The prompt asks for these sections in this order.
	}

	return sb.String()
}

func renderDiffSection(sb *strings.Builder, title string, diffs []model.StepDiff, kinds ...model.FlowStepKind) {
	relevant := []model.StepDiff{}
	for _, d := range diffs {
		isKind := false
		for _, k := range kinds {
			if (d.Before != nil && d.Before.Kind == k) || (d.After != nil && d.After.Kind == k) {
				isKind = true
				break
			}
		}
		if isKind && d.Kind != model.StepUnchanged {
			relevant = append(relevant, d)
		}
	}

	if len(relevant) == 0 {
		return
	}

	sb.WriteString(fmt.Sprintf("### %s\n", title))
	for _, rd := range relevant {
		switch rd.Kind {
		case model.StepAdded:
			sb.WriteString(fmt.Sprintf("+ Added %s: %s\n", rd.After.Kind, rd.After.Description))
		case model.StepRemoved:
			sb.WriteString(fmt.Sprintf("- Removed %s: %s\n", rd.Before.Kind, rd.Before.Description))
		case model.StepModified:
			sb.WriteString(fmt.Sprintf("~ Modified %s:\n", rd.After.Kind))
			sb.WriteString(fmt.Sprintf("  - PREV: %s\n", rd.Before.Description))
			sb.WriteString(fmt.Sprintf("  - NEXT: %s\n", rd.After.Description))
			if rd.Before.ToMethod != rd.After.ToMethod {
				sb.WriteString(fmt.Sprintf("  - Target changed: `%s` -> `%s`\n", rd.Before.ToMethod, rd.After.ToMethod))
			}
			if rd.Before.ResolutionScope != rd.After.ResolutionScope {
				sb.WriteString(fmt.Sprintf("  - Resolution changed: %s -> %s\n", rd.Before.ResolutionScope, rd.After.ResolutionScope))
			}
		}
	}
	sb.WriteString("\n")
}
