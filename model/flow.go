package model

// ExecutionFlow represents the extracted execution flow for a specific REST resource.
type ExecutionFlow struct {
	ResourceName string
	EntryPoint   string // HTTP method + path
	Steps        []FlowStep
}

// FlowStepKind defines the type of execution step.
type FlowStepKind string

const (
	FlowStepEntry      FlowStepKind = "entry"
	FlowStepCondition  FlowStepKind = "condition"
	FlowStepCall       FlowStepKind = "call"
	FlowStepOutbound   FlowStepKind = "outbound"
	FlowStepReturn     FlowStepKind = "return"
	FlowStepUnexpanded FlowStepKind = "unexpanded"
)

// FlowStep represents a single step in an execution flow.
type FlowStep struct {
	Index           int
	Kind            FlowStepKind
	Description     string
	FromMethod      string
	ToMethod        string
	Confidence      string
	Evidence        string
	ResolutionScope string
}
