# jz — Design Contract

## Purpose

`jz` is a CLI-based static analysis tool written in Go that helps engineers
understand large enterprise Java backend systems built using:

- OSGi
- JAX-RS
- WebSphere Liberty
- Ant and/or Maven

The tool extracts **deterministic, explainable facts** from source code and
configuration to build a **human-understandable architectural model**.

The primary use case is **onboarding and system understanding**, not build
analysis or runtime simulation.

---

## Core Principles

### 1. Deterministic First
- All extracted information must come from static analysis
- No guessing, inference, heuristics, or probabilistic logic
- If something cannot be determined statically, it is ignored

### 2. Runtime-Relevant Only
- Prefer runtime configuration over source conventions
- Ignore code that does not participate in runtime execution
- Configuration files (MANIFEST.MF, DS XML, server.xml) are authoritative

### 3. Explicit Boundaries
- **Service**: a deployable runtime unit (OSGi bundle / Liberty application)
- **Component**: an OSGi Declarative Service defined via XML
- **Dependency**: an explicit reference from one component/service to another

### 4. Explainability
- Every relationship must be traceable to a concrete source
- The tool must be able to explain *why* a dependency exists
- No “black box” results

### 5. No Magic
- No reflection
- No bytecode analysis
- No runtime simulation
- No dynamic resolution
- No execution of build tools

---

## What jz Analyzes

### Services
- OSGi bundles with runtime relevance
- Deployed via WebSphere Liberty

### Entry Points
- JAX-RS REST endpoints
- Derived from `@Path` and HTTP method annotations

### Components
- OSGi Declarative Services defined via Service-Component XML

### Dependencies
- Internal: DS `provide` → `reference`
- System-level: service-to-service via shared interfaces

### Runtime Context
- Liberty `server.xml`
- Enabled features
- Application deployment grouping

---

## What jz Does NOT Do

- Execute Ant or Maven builds
- Resolve Maven dependency graphs
- Simulate OSGi runtime behavior
- Infer business logic
- Infer message flows
- Analyze performance or scalability
- Modify or generate source code
- Provide refactoring suggestions

---

## Technology Constraints

- Language: Go
- Prefer Go standard library
- Avoid heavy frameworks
- Prefer simple string scanning over complex AST parsing
- Parsing must be readable and debuggable

---

## CLI Design

- The CLI must remain minimal
- Commands map directly to analysis stages
- Prefer Go standard library for argument handling
- CLI frameworks (e.g., Cobra) may be introduced later **only if justified**
- The CLI is a thin shell; all logic lives outside command handlers

---

## Output Formats

- JSON: Intermediate Representation (source of truth)
- Markdown: human-readable reports
- Mermaid: architecture diagrams

---

## Role of AI

AI may be used ONLY to:
- Generate boilerplate code
- Implement clearly specified parsing logic
- Translate structured IR into human-readable explanations (optional phase)

AI must NEVER:
- Redefine the architecture
- Invent relationships or behaviors
- Expand scope beyond the requested task
- Introduce new concepts or abstractions

This document is the **single source of truth** for the project.
