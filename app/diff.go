package app

import (
	"jz/model"
)

// DiffFlows compares two sets of execution flows for the same resource.
func DiffFlows(flowsA, flowsB []model.ExecutionFlow) []model.FlowDiff {
	diffs := []model.FlowDiff{}

	// Map flows by EntryPoint for lookup
	mapA := make(map[string]model.ExecutionFlow)
	for _, f := range flowsA {
		mapA[f.EntryPoint] = f
	}

	mapB := make(map[string]model.ExecutionFlow)
	for _, f := range flowsB {
		mapB[f.EntryPoint] = f
	}

	// Collected all unique EntryPoints to ensure deterministic order
	allEntries := []string{}
	seen := make(map[string]bool)

	// Use order from A then B
	for _, f := range flowsA {
		if !seen[f.EntryPoint] {
			allEntries = append(allEntries, f.EntryPoint)
			seen[f.EntryPoint] = true
		}
	}
	for _, f := range flowsB {
		if !seen[f.EntryPoint] {
			allEntries = append(allEntries, f.EntryPoint)
			seen[f.EntryPoint] = true
		}
	}

	for _, entry := range allEntries {
		flowA, okA := mapA[entry]
		flowB, okB := mapB[entry]

		var diff model.FlowDiff
		diff.EntryPoint = entry

		if !okA {
			diff.Status = "ADDED"
			for _, step := range flowB.Steps {
				s := step // copy
				diff.StepDiffs = append(diff.StepDiffs, model.StepDiff{
					Kind:  model.StepAdded,
					After: &s,
				})
			}
		} else if !okB {
			diff.Status = "REMOVED"
			for _, step := range flowA.Steps {
				s := step // copy
				diff.StepDiffs = append(diff.StepDiffs, model.StepDiff{
					Kind:   model.StepRemoved,
					Before: &s,
				})
			}
		} else {
			// Both exist, diff steps
			diff.StepDiffs = diffSteps(flowA.Steps, flowB.Steps)
			diff.Status = "UNCHANGED"
			for _, sd := range diff.StepDiffs {
				if sd.Kind != model.StepUnchanged {
					diff.Status = "MODIFIED"
					break
				}
			}
		}
		diffs = append(diffs, diff)
	}

	return diffs
}

// diffSteps performs a sequence-based diff of flow steps.
// Since the requirement says "NO reordering" and "ordered comparison",
// we compare them index by index.
func diffSteps(stepsA, stepsB []model.FlowStep) []model.StepDiff {
	diffs := []model.StepDiff{}
	maxLen := len(stepsA)
	if len(stepsB) > maxLen {
		maxLen = len(stepsB)
	}

	for i := 0; i < maxLen; i++ {
		if i < len(stepsA) && i < len(stepsB) {
			sA := stepsA[i]
			sB := stepsB[i]
			if stepsEqual(sA, sB) {
				diffs = append(diffs, model.StepDiff{
					Kind:   model.StepUnchanged,
					Before: &sA,
					After:  &sB,
				})
			} else {
				diffs = append(diffs, model.StepDiff{
					Kind:   model.StepModified,
					Before: &sA,
					After:  &sB,
				})
			}
		} else if i < len(stepsA) {
			sA := stepsA[i]
			diffs = append(diffs, model.StepDiff{
				Kind:   model.StepRemoved,
				Before: &sA,
			})
		} else if i < len(stepsB) {
			sB := stepsB[i]
			diffs = append(diffs, model.StepDiff{
				Kind:  model.StepAdded,
				After: &sB,
			})
		}
	}
	return diffs
}

func stepsEqual(a, b model.FlowStep) bool {
	return a.Kind == b.Kind &&
		a.Description == b.Description &&
		a.ToMethod == b.ToMethod &&
		a.ResolutionScope == b.ResolutionScope
}
