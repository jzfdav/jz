package model

// DetectionType describes how a call was discovered.
const (
	DetectionLiteral  = "literal"
	DetectionConstant = "constant"
	DetectionUnknown  = "unknown"
)

// Confidence level for the detected call.
const (
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
	ConfidenceLow    = "low"
)

// RESTCall represents an outbound HTTP call detected in the source code.
// It captures facts about the call's origin and potential target.
type RESTCall struct {
	FromService    string
	FromResource   string
	FromHandler    string
	HTTPMethod     string
	TargetPath     string // Literal or resolved string
	TargetService  string // Only populated if unambiguous (e.g., same service)
	TargetResource string // Only populated if unambiguous
	SourceFile     string
	DetectionType  string // literal, constant, unknown
	Confidence     string // high, medium, low
}

// ServiceBoundary represents an architectural boundary detected within a service.
type ServiceBoundary struct {
	ServiceName  string
	BoundaryType string // package, resource-group, layer
	Identifier   string // e.g., package name or prefix
	Evidence     string // Short string explaining why this boundary exists
}
