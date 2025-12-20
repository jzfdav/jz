# Quick Start — jz

This guide helps you get value from `jz` in **10 minutes or less**, even if the codebase is unfamiliar.

📘 **Diagram-first?** Check the [Mermaid Quickstart](./quickstart-mermaid.md) for a visual-only companion designed for GitHub rendering.

---

## Prerequisites

- **Go 1.21+**
- Java source code (OSGi or Liberty WAR style)
- No build, no runtime, no database required

---

## Installation

From the repository root:

```bash
go install ./cmd/jz
```

Verify installation:

```bash
jz --help
```

---

## Step 1: Scan the Repository

Start with a high-level structural overview.

```bash
jz scan .
```

This tells you:
- How many services were detected
- Whether the repo is modeled as OSGi bundles or a single Liberty WAR
- Basic REST surface area

Use this step to confirm **jz understands your repo shape correctly**.

---

## Step 2: Generate a Full Markdown Report

```bash
jz report markdown .
```

You’ll get:
- Services and REST resources
- Entry points (HTTP method + path)
- Inbound and outbound REST calls
- Cross-service resolution (if provable)
- Deterministic summaries

**Tip:** Read this report once top-down, then keep it as a reference.

---

## Step 3: Visualize REST Interactions

Generate a global REST interaction graph:

```bash
jz report mermaid . --calls
```

This shows:
- Resource-to-resource calls
- Same-service vs cross-service interactions
- Unresolved calls (explicitly marked)

Arrow meanings:
- `-->` same-service  
- `==>` cross-service  
- `-.->` unresolved / conditional  

---

## Step 4: Extract a Targeted Execution Flow

Focus on **one API at a time**.

```bash
jz flow extract . --resource ExampleApiV1
```

This produces:
- A step-by-step execution narrative
- Guard conditions (early checks)
- Outbound REST calls
- Early returns vs normal termination

This is the fastest way to understand *what really happens* inside an API.

---

## Step 5: Visualize the Flow (Optional)

```bash
jz flow extract . --resource ExampleApiV1 --format mermaid --compact
```

Use this when:
- Explaining logic to others
- Reviewing guard-heavy endpoints
- Spotting early exits visually

`--compact` collapses guard chains into a single decision node.

---

## Step 6: Compare Flows Between Versions

When reviewing changes or refactors:

```bash
jz flow diff ./before ./after --resource ExampleApiV1
```

This highlights:
- Added or removed guards
- New or removed outbound calls
- Changes in resolution scope
- Structural logic changes only (no speculation)

Reordering is treated as a change **by design**.

---

## How to Interpret Results

- **Unresolved ≠ Broken**  
  It means the call could not be *proven* statically.

- **Confidence ≠ Resolution Scope**  
  Confidence = detection certainty  
  Scope = where the call was linked (if at all)

- **Guards are gates, not paths**  
  They explain *why* execution may stop early.

- **Flows are conservative**  
  Missing steps are intentional, not bugs.

---

## When jz Is Most Useful

- Onboarding to legacy systems
- Reviewing API behavior before changes
- Auditing cross-service dependencies
- Explaining logic to non-authors
- Preparing refactors safely

---

## Known Limitations (By Design)

- No runtime execution
- No variable or data propagation
- No speculative inference
- No auth or schema modeling

Every reported fact is backed by literal source evidence.

---

## Next Steps

- Read the main README for design philosophy
- Use `jz flow diff` during code reviews
- Embed Mermaid diagrams into docs or PRs
