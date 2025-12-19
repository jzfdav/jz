package model

// RESTResource groups JAX-RS entry points by their implementation class.
type RESTResource struct {
	Name       string // Resource class name (e.g. TenantApiV1)
	SourceFile string // Java source file path
	BasePath   string // Class-level @Path if known

	// Phase F3.2 additions
	Methods     []RESTMethod   // Flattened REST operations
	HTTPMethods map[string]int // Summary: GET -> 12, POST -> 4

	// Existing (backward compatibility)
	EntryPoints []EntryPoint
}

// RESTMethod represents a single REST operation mapped to a handler method.
type RESTMethod struct {
	HTTPMethod string // GET, POST, PUT, DELETE, etc.
	SubPath    string // Method-level @Path
	FullPath   string // BasePath + SubPath
	Handler    string // e.g. TenantApiV1.getTenant
	SourceFile string
}
