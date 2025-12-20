package report

import (
	"fmt"
	"jz/model"
	"strings"
)

// GenerateFlowMermaid creates a Mermaid flow diagram for execution steps.
func GenerateFlowMermaid(flows []model.ExecutionFlow, resourceName string) string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	for i, f := range flows {
		flowID := fmt.Sprintf("Flow_%d", i)
		sb.WriteString(fmt.Sprintf("\tsubgraph %s [%s]\n", flowID, f.EntryPoint))

		var lastNode string
		for j, s := range f.Steps {
			nodeID := fmt.Sprintf("F%d_S%d", i, j)
			label := s.Description
			if s.Kind == model.FlowStepCondition {
				label = "{{" + s.Description + "}}"
			} else if s.Kind == model.FlowStepOutbound || s.Kind == model.FlowStepCall {
				label = "[" + s.Description + "]"
			} else if s.Kind == model.FlowStepReturn {
				label = "((" + s.Description + "))"
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
		sb.WriteString("\tend\n")
	}

	// Legend
	sb.WriteString("\n\t%% Legend\n")
	sb.WriteString("\tsubgraph Legend\n")
	sb.WriteString("\t\tl1[Same-service Call] --> l2[Next Step]\n")
	sb.WriteString("\t\tl3[Cross-service Call] ==> l4[Next Step]\n")
	sb.WriteString("\t\tl5[Conditional/Unresolved] -.-> l6[Next Step]\n")
	sb.WriteString("\tend\n")

	return sb.String()
}
