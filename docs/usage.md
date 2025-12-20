# jz Usage Guide

📘 **New to jz?** Start with the [Quick Start](./quickstart.md). This document is a technical reference for flags and commands.

This guide provides detailed information on running `jz` and interpreting its results.

## Global Flags

Most `jz` commands support the following flags:

- `--resource <Name>`: (Required for `flow` commands) The class name of the JAX-RS resource to analyze.
- `--service <Name>`: Filter output to a specific service.
- `--output <path>`: Write the report to a file instead of stdout.
- `--format <markdown|mermaid|all>`: Select the output format.

---

## Commands

### `jz scan <path>`
Performs a high-level scan of the directory tree.
- Detects OSGi bundles, Liberty configurations, and JAX-RS resources.
- Provides a summary of entry points and system-level diagnostics.

### `jz report markdown <path>`
Generates a comprehensive Markdown report.
- Breaks down every service and its associated REST resources.
- Lists outbound calls detected within handlers.
- Surfaces "Inbound Calls" for resources that are targets of other services.

### `jz report mermaid <path> [--calls]`
Visualizes the system architecture.
- Default: Shows service and component-level dependencies.
- `--calls`: Shows the **Resource Interaction Graph**, tracing how APIs call each other.

### `jz flow extract <path>`
Extracts the logic of a specific resource.
- Use `--method` and `--path` to narrow down to a single endpoint.
- Use `--max-depth <N>` (default 3) to control how deep internal method calls are expanded.
- Use `--compact` with `--format mermaid` to merge sequential guard conditions.

### `jz flow diff <pathA> <pathB>`
Compares two versions of a codebase.
- Focuses on structural changes in the execution flow.
- Identifies added/removed guards, modified outbound call targets, and changes in termination logic.

---

## Understanding Analysis Results

### Confidence Levels
`jz` assigns confidence to every detected outbound REST call:
- **High**: String literals for URLs (e.g., `"http://example-service/v1/example"`).
- **Medium**: Partial literals or well-known client patterns.
- **Low**: URLs built from variables or complex expressions that AST-lite cannot resolve.

### Resolution Scopes
When `jz` finds an outbound call, it tries to link it to a known resource:
- **same-service**: The target is within the same OSGi bundle or Liberty app.
- **cross-service**: The target is a unique match found in another service in the same scan.
- **unresolved**: No unique match was found. This happens if the URL is dynamic, use Constants, or points to an external system not included in the scan.

---

## Troubleshooting

### No services detected
- Ensure you are scanning the root of the project.
- `jz` looks for `META-INF/MANIFEST.MF` for OSGi or `server.xml` for Liberty.
- If your project uses a different runtime, `jz` may not model it automatically.

### High number of unresolved calls
- Check if your code uses constants for URLs. AST-lite does not resolve constants across files.
- Ensure all target services are included in the directory path provided to `jz`.
