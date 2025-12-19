package model

// Service represents a deployable runtime unit (OSGi bundle / Liberty application).
type Service struct {
	Name        string
	RootPath    string
	EntryPoints []EntryPoint
	Components  []DSComponent

	// Internal component-level dependency graph
	InternalGraph DependencyGraph

	// Liberty runtime context
	ServerName  string
	Features    []string
	Application LibertyApp

	// REST Resources (grouped entry points)
	RESTResources []RESTResource

	// Phase F4 additions
	RESTCalls  []RESTCall
	Boundaries []ServiceBoundary
}

// EntryPoint represents a REST entry point.
type EntryPoint struct {
	Method     string
	Path       string
	Handler    string
	SourceFile string
	Resource   string // Resource class name (derived from handler)
}

// DSComponent represents an OSGi Declarative Service.
type DSComponent struct {
	Name                 string
	ImplementationClass  string
	Immediate            bool
	ProvidedInterfaces   []string
	ReferencedInterfaces []string
	SourceXML            string
}

// LibertyServer represents a Liberty server instance.
type LibertyServer struct {
	Name            string
	ServerXML       string
	EnabledFeatures []string
	DeployedApps    []LibertyApp
}

// LibertyApp represents an application deployed in Liberty.
type LibertyApp struct {
	ID          string
	Location    string
	Type        string
	ContextRoot string
}

// DependencyGraph represents internal component-level dependencies.
type DependencyGraph struct {
	Nodes []ComponentNode
	Edges []DependencyEdge
}

// ComponentNode represents a DS component in a graph.
type ComponentNode struct {
	Name                string
	ImplementationClass string
	Immediate           bool
}

// DependencyEdge represents a dependency between components.
type DependencyEdge struct {
	FromComponent string
	ToComponent   string
	Interface     string
}

// SystemGraph represents system-level service dependencies.
type SystemGraph struct {
	Services     []string
	Dependencies []ServiceDependency
}

// ServiceDependency represents a dependency from one service to another.
type ServiceDependency struct {
	FromService string
	ToService   string
	Interface   string
}
