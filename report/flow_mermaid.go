package report

import (
	"fmt"
	"jz/model"
	"strings"
)

// GenerateFlowMermaid creates a Mermaid flow diagram for execution steps.
func GenerateFlowMermaid(flows []model.ExecutionFlow, resourceName string, compact bool) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	for i, f := range flows {
		flowID := fmt.Sprintf("Flow_%d", i)
		sb.WriteString(fmt.Sprintf("\tsubgraph %s [%s]\n", flowID, f.EntryPoint))

		var lastNode string
		var guardChain []string

		for j, s := range f.Steps {
			// Compact guards if enabled
			if compact && s.Kind == model.FlowStepCondition {
				guardChain = append(guardChain, s.Description)
				// If next step is not a guard, or this is the last step, flush the chain
				if j == len(f.Steps)-1 || f.Steps[j+1].Kind != model.FlowStepCondition {
					nodeID := fmt.Sprintf("F%d_G_%d", i, j)
					label := strings.Join(guardChain, " && ")
					if len(guardChain) > 1 {
						label = "GUARDS: " + label
					}
					sb.WriteString(fmt.Sprintf("\t\t%s{{\"%s\"}}\n", nodeID, label))
					if lastNode != "" {
						sb.WriteString(fmt.Sprintf("\t\t%s -.->|CONDITION| %s\n", lastNode, nodeID))
					}
					lastNode = nodeID
					guardChain = nil
				}
				continue
			}

			nodeID := fmt.Sprintf("F%d_S%d", i, j)
			label := s.Description

			// 6. Explicit flow termination nodes logic handled later
			switch s.Kind {
			case model.FlowStepCondition:
				label = "{{" + label + "}}"
			case model.FlowStepOutbound:
				label = "[[" + label + "]]" // Double bracket for outbound
			case model.FlowStepCall:
				label = "[" + label + "]"
			case model.FlowStepUnexpanded:
				label = "[/ " + label + " /]" // Parallelogram for unexpanded/external
			case model.FlowStepReturn:
				label = "((" + label + "))"
			}

			sb.WriteString(fmt.Sprintf("\t\t%s(\"%s\")\n", nodeID, label))

			if lastNode != "" {
				arrow := "-->"
				if s.Kind == model.FlowStepCondition {
					arrow = "-.->"
				}

				edgeLabel := fmt.Sprintf("%s [%s]", strings.ToUpper(string(s.Kind)), s.Confidence)

				// Special arrow for outbound calls based on resolution
				if s.Kind == model.FlowStepOutbound {
					switch s.ResolutionScope {
					case model.ResolutionSameService:
						arrow = "-->"
					case model.ResolutionCrossService:
						arrow = "==>"
					default:
						arrow = "-.->"
					}
				}

				sb.WriteString(fmt.Sprintf("\t\t%s %s|%s| %s\n", lastNode, arrow, edgeLabel, nodeID))
			}
			lastNode = nodeID
		}

		// 6. Explicit Flow Termination Nodes
		if len(f.Steps) > 0 {
			lastStep := f.Steps[len(f.Steps)-1]
			termID := fmt.Sprintf("F%d_TERM", i)
			if lastStep.Kind == model.FlowStepReturn {
				sb.WriteString(fmt.Sprintf("\t\t%s(X \"End (Return)\")\n", termID))
				sb.WriteString(fmt.Sprintf("\t\t%s --> %s\n", lastNode, termID))
			} else {
				sb.WriteString(fmt.Sprintf("\t\t%s[/ \"End (Unexpanded)\" /]\n", termID))
				sb.WriteString(fmt.Sprintf("\t\t%s -.->|SCOPE LIMIT| %s\n", lastNode, termID))
			}
		}

		sb.WriteString("\tend\n")
	}

	// Legend (Updated)
	sb.WriteString("\n\t%% Legend\n")
	sb.WriteString("\tsubgraph Legend\n")
	sb.WriteString("\t\tl1[Internal Call] --> l2[Next Step]\n")
	sb.WriteString("\t\tl3[[Outbound Call]] ==> l4[Cross-service Target]\n")
	sb.WriteString("\t\tl5{{Condition}} -.-> l6[Guarded Path]\n")
	sb.WriteString("\t\tl7[/ Unexpanded Call /] -.-> l8[End (Unexpanded)]\n")
	sb.WriteString("\tend\n")

	return sb.String()
}
