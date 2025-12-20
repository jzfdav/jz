# jz — Static Analysis for Large Java Backends

`jz` is a static analysis tool designed to help engineers understand large, legacy, multi-service Java systems—specifically those built with OSGi, JAX-RS, and deployed on WebSphere Liberty. It extracts architecture, dependencies, and execution flows directly from source code without requiring a running environment.

> All identifiers, examples, and diagrams in this project are fictional and provided solely for demonstration purposes.

📘 New here? Start with [Quick Start](docs/quickstart.md)

## Documentation Overview

- **README.md**: High-level overview, design philosophy, and worked examples.
- **docs/quickstart.md**: Getting started guide for narrative onboarding.
- **docs/usage.md**: Technical reference for all CLI commands and flags.
- **docs/quickstart-mermaid.md**: Visual, diagram-first introduction to `jz`.

---

## What jz Is
- **Static analysis tool**: Uses AST-lite techniques to parse source files.
- **REST-focused**: Specialized in JAX-RS entry points and their downstream interactions.
- **Execution flow engine**: Extracts step-by-step logic from handlers.
- **Deterministic & Conservative**: Only reports facts it can prove lexically; prioritizes safety and accuracy over guesses.

## What jz Is NOT
- **No runtime tracing**: Does not monitor or execute live code.
- **No data-flow resolution**: Does not track variable values or state propagation.
- **No auth or role inference**: Does not interpret security logic or permissions.
- **No speculative linking**: Only connects services if an unambiguous match is found.
- **No business-logic interpretation**: Does not "understand" what the code is trying to achieve (e.g., "complex validation logic").

---

## Core Concepts

- **Service**: A deployable unit, such as an OSGi bundle or a Liberty Web application (WAR).
- **REST Resource**: A Java class containing JAX-RS annotations representing a set of API endpoints.
- **Entry Point**: A specific HTTP method and path mapped to a single handler method.
- **Execution Flow**: A captured sequence of steps (guards, calls, returns) inside a handler.
- **Resolution Scope**: 
  - `same-service`: A call linked to a resource within the same deployment unit.
  - `cross-service`: A call linked to a resource in a different service.
  - `unresolved`: A call that could not be proved to target a known resource.
- **Confidence Levels**:
  - `high`: Exact literal matches (e.g., hardcoded URL strings).
  - `medium`: Likely matches with some abstraction.
  - `low`: Inferred or complex patterns that cannot be fully verified.
- **AST-lite Analysis**: A high-performance, line-based scanning technique that extracts structure without the overhead or fragility of a full Java compiler frontend.

---

## Command Overview

| Command | Purpose | Output |
| :--- | :--- | :--- |
| `jz scan <path>` | Quick system summary and diagnostics | Markdown |
| `jz report markdown <path>` | Full static analysis including all services and resources | Markdown |
| `jz report mermaid <path> --calls` | Global REST resource interaction graph | Mermaid |
| `jz flow extract <path>` | Detailed step-by-step execution flow for one resource | Markdown / Mermaid |
| `jz flow diff <pathA> <pathB>` | Structural difference between two versions of a flow | Markdown |

---

## Worked Examples

### Example 1: Targeted Flow Extraction
Extract the execution narrative for a specific API resource:

```bash
jz flow extract . --resource ExampleApiV1
```

- **Analyzed Scope**: Only the `ExampleApiV1` class and its immediate interactions are scanned.
- **Logic Capture**: `jz` identifies `ENTRY`, `GUARD`, `OUTBOUND`, and `RETURN` steps.
- **Safety**: Dynamic behaviors and variable-based routing are intentionally ignored to ensure the output is deterministic.

### Example 2: Mermaid Flow Visualization
Generate a visual diagram of the flow with guard chain compaction:

```bash
jz flow extract . --resource ExampleApiV1 --format mermaid --compact
```

- **Guard Compaction**: Sequential guard conditions are collapsed into a single decision node for readability.
- **Arrow Semantics**: `-->` denotes same-service calls, `==>` denotes cross-service, and `-.->` denotes conditional or unresolved paths.
- **Termination**: Explicit nodes signal `End (Return)` vs `End (Unexpanded)` (where analysis reached a scope limit).

### Example 3: Flow Diff Between Versions
Compare how a flow changed after a refactor or feature addition:

```bash
jz flow diff ./v1 ./v2 --resource ExampleApiV1
```

- **Output**: Clearly marks `ADDED`, `REMOVED`, or `MODIFIED` steps.
- **Strict Matching**: Diffs are performed step-by-step; reordering is reported as a structural change rather than a match.
- **Fact-Based**: The report surfaces structural changes only, leaving business interpretation to the reviewer.

---

## How to Read jz Output

- **Guard Chains vs. Core Path**: Guards are typically early conditional checks. `jz` displays them as gating the subsequent logic.
- **Early Returns vs. Normal Termination**: An early exit is a return statement occurring before the end of the method, often as a result of a guard check.
- **Unresolved Outbound Calls**: These are calls `jz` detected but could not uniquely link to a resource. This is normal for dynamic URLs or third-party APIs.
- **Confidence vs. Resolution Scope**: Confidence describes the *certainty* of the call detection; Scope describes *where* the call goes if it was resolved.
- **Unresolved ≠ Broken**: An unresolved call simply means it is beyond the scope of local static proof (e.g., depends on runtime configuration).

---

## Limitations & Safety Guarantees

- **Preference for False Negatives**: `jz` will skip reporting a dependency or flow step if it is ambiguous, ensuring that every reported item is backed by literal source evidence.
- **Static Only**: Dynamic logic (reflection, bytecode injection, runtime proxies) is intentionally ignored to maintain audit-level reliability.
- **Ordered Execution**: Flows are extracted in lexical order. While `jz` captures branches, it does not guarantee runtime execution ordering beyond what is written in the source.
- **Safety**: As a read-only static analyzer, `jz` is safe to run against production source code without side effects.

---

## Roadmap

### Completed (v1.x)
- **Phase F4/F5**: Structural analysis and Cross-Service call resolution.
- **Phase F6.0/F6.1**: Targeted execution flow extraction and UX refinement.
- **Phase F6.2**: Cross-version flow diffing.

### Possible Future
- **Auth Propagation**: Tracking security annotations across service boundaries.
- **Schema Modeling**: Basic mapping of Request/Response POJOs.
- **Extended AST-lite Heuristics**: Support for simple constant resolution and enhanced pattern matching.

---

## Installation

### From Source
Requires Go 1.21+

```bash
go install ./cmd/jz
```

## License
MIT
