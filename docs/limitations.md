# Limitations & Safety Guarantees

This document describes the intentional limitations of **jz** and the guarantees it provides as a static analysis tool. These constraints are deliberate design choices to ensure determinism, safety, and auditability.

---

## Core Design Philosophy

- **Static-only**: jz never executes code, loads classes, or inspects bytecode.
- **Deterministic**: The same input always produces the same output.
- **Conservative**: jz prefers false negatives over false positives.
- **Evidence-based**: Every reported item is backed by literal source evidence.

---

## What jz Does Not Do

### ❌ No Runtime Behavior
- No tracing, profiling, or runtime instrumentation
- No simulation of execution paths
- No environment-dependent behavior

### ❌ No Data Flow or State Tracking
- Variables are not resolved across assignments
- Object lifecycles are not tracked
- Conditional truth values are not evaluated

### ❌ No Authentication or Authorization Semantics
- Security annotations are detected but not interpreted
- Auth context is not propagated across calls
- Access control correctness is out of scope

### ❌ No Schema or Payload Modeling
- Request/response POJOs are not parsed
- JSON/XML schemas are not inferred
- Validation logic is not evaluated

### ❌ No Speculative Inference
- Calls are only linked when an unambiguous match exists
- Dynamic URLs, reflection, and factories are ignored
- Ambiguous matches remain unresolved by design

---

## Execution Flow Limitations (F6.x)

- Flow extraction is **lexical**, not semantic
- Loops are not unrolled
- Only same-file internal method expansion is supported
- Cross-service flow continuation is summarized, not expanded
- Reordering of steps is treated as a structural change

---

## Diffing Limitations (F6.2)

- Step-by-step ordered comparison only
- No semantic equivalence detection
- No tolerance for reordering or refactoring noise
- Focused on *what changed*, not *why it changed*

---

## Safety Guarantees

- ✅ Safe to run on production source code
- ✅ No network access
- ✅ No filesystem mutation
- ✅ No code execution
- ✅ Read-only analysis

---

## Why These Limits Exist

jz is designed for:
- Architecture discovery
- Risk assessment
- Codebase orientation
- Review and audit support

It is **not** designed to replace:
- Runtime observability tools
- Security scanners
- Full compilers or IDEs
- Human code review

---

## Summary

If jz reports something, it is because it can *prove it* from source text alone.

If jz does **not** report something, it means the evidence was ambiguous or outside its conservative analysis scope.

This trade-off is intentional and central to jz’s reliability.
