package model

// FlowDiff represents the difference between two execution flows.
type FlowDiff struct {
	EntryPoint string
	Status     string // UNCHANGED, MODIFIED, ADDED, REMOVED
	StepDiffs  []StepDiff
}

// StepDiffKind defines the type of change in a step.
type StepDiffKind string

const (
	StepAdded     StepDiffKind = "added"
	StepRemoved   StepDiffKind = "removed"
	StepModified  StepDiffKind = "modified"
	StepUnchanged StepDiffKind = "unchanged"
)

// StepDiff represents the difference between two flow steps.
type StepDiff struct {
	Kind   StepDiffKind
	Before *FlowStep
	After  *FlowStep
}
