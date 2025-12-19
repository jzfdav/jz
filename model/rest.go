package model

// RESTResource groups JAX-RS entry points by their implementation class.
type RESTResource struct {
	Name        string // Resource class name (e.g. TenantApiV1)
	SourceFile  string // Java source file path
	BasePath    string // Class-level @Path if available
	EntryPoints []EntryPoint
}
