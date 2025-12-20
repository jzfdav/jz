package model

// RESTResource groups JAX-RS entry points by their implementation class.
type RESTResource struct {
	Name       string // Resource class name (e.g. ExampleApiV1)
	SourceFile string // Java source file path
	BasePath   string // Class-level @Path if known

	// Phase F3.2 additions
	Methods     []RESTMethod   // Flattened REST operations
	HTTPMethods map[string]int // Summary: GET -> 12, POST -> 4

	// Existing (backward compatibility)
	EntryPoints []EntryPoint

	// Phase F3.4 additions
	AuthAnnotations []string
	Consumes        []string
	Produces        []string
	PathParams      []string

	// Phase F4 additions
	OutboundCalls []RESTCall // Calls originating from this resource
	InboundCalls  []RESTCall // Calls targeting this resource (if known)
}

// RESTMethod represents a single REST operation mapped to a handler method.
type RESTMethod struct {
	HTTPMethod string // GET, POST, PUT, DELETE, etc.
	SubPath    string // Method-level @Path
	FullPath   string // BasePath + SubPath
	Handler    string // e.g. ExampleApiV1.handleExample
	SourceFile string
}
