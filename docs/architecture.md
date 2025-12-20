# Architecture Overview (Contributor Guide)

> This document is intended for contributors to understand *how jz is built*, not how to use it.

## Design Philosophy
jz is built around a few non-negotiable principles:

- **Static-only analysis**: Never execute user code.
- **AST-lite approach**: Prefer fast, line-based scanning over full compiler ASTs.
- **Determinism**: The same input must always produce the same output.
- **Conservatism**: Prefer false negatives over false positives.
- **Explainability**: Every reported fact must have clear source evidence.

## High-Level Architecture

```
cmd/        → CLI commands and flag wiring (cobra)
app/        → Analysis orchestration and flow/diff engines
model/      → Shared domain models (services, resources, flows, diffs)
report/     → Markdown and Mermaid renderers
docs/       → User and contributor documentation
```

## Core Pipelines

### 1. Scan / Analyze
- Entry point: `app.Analyze`
- Discovers services, REST resources, and metadata
- Performs AST-lite scanning without symbol resolution

### 2. Structural REST Analysis (F4/F5)
- Detects outbound REST calls via lexical patterns
- Resolves same-service and cross-service calls conservatively
- Annotates calls with confidence and resolution scope

### 3. Execution Flow Extraction (F6.0 / F6.1)
- Targets a single REST resource at a time
- Extracts guards, internal calls, outbound calls, and returns
- Uses bounded recursion with cycle protection
- Clearly marks unexpanded or scope-limited paths

### 4. Flow Diffing (F6.2)
- Compares two extracted flows step-by-step
- No reordering or semantic inference
- Reports only structural changes (added/removed/modified)

## What Not to Add
Contributors should **not** introduce:
- Runtime execution or reflection
- Speculative inference or heuristics without proof
- Deep Java parsing requiring full AST or bytecode
- Business-logic interpretation

If in doubt, leave it out.

## Adding New Features Safely
When proposing changes:
1. Clearly state assumptions and limitations
2. Show worst-case false-positive behavior
3. Keep heuristics opt-in where possible
4. Ensure output remains explainable

## Testing Philosophy
- Prefer small, focused fixtures
- Validate determinism
- Test negative and ambiguous cases explicitly

---
This document exists to preserve architectural integrity as the project evolves.
