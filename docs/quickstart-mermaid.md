# jz Quickstart (Mermaid-only)

This quickstart introduces **jz** using GitHub-renderable Markdown and Mermaid diagrams only.
No screenshots. No runtime execution. Pure static analysis.

ℹ️ This is a diagram-only companion. For full narrative explanations, see the [narrative Quickstart](./quickstart.md).

---

## What jz Does

```mermaid
flowchart TD
    A[Java Source Code] --> B[jz Static Analysis]
    B --> C[Services & Resources]
    B --> D[REST Interactions]
    B --> E[Execution Flows]
    B --> F[Flow Diffs]
```

jz analyzes Java source code **without running it**, extracting:
- REST entry points
- Cross-service calls
- Handler execution flows
- Structural diffs between versions

---

## Installation

```bash
go install ./cmd/jz
```

Requires **Go 1.21+**.

---

## 1. Scan a Codebase

```bash
jz scan .
```

```mermaid
graph TD
    Repo[Repository] --> Service[Detected Service]
    Service --> Resources[REST Resources]
```

Purpose:
- Detect services (OSGi bundles or Liberty WAR)
- List REST resources and endpoints
- Surface diagnostics

---

## 2. Full Static Report

```bash
jz report markdown .
```

```mermaid
graph LR
    example-service --> ExampleApiV1
    example-service --> ExampleApiV2
    ExampleApiV1 -->|GET| ExampleApiV2
```

Generates:
- Service summaries
- REST resource listings
- Confidence and resolution metadata

---

## 3. REST Interaction Graph

```bash
jz report mermaid . --calls
```

```mermaid
graph TD
    A[example-service-a/ExampleApiV1] -->|GET [same]| B[example-service-a/ExampleApiV2]
    A ==> |POST [cross]| C[example-service-b/ExampleApiV3]
    A -.-> |unresolved| D[UNKNOWN]
```

Legend:
- `-->` same-service
- `==>` cross-service
- `-.->` unresolved

---

## 4. Targeted Execution Flow

```bash
jz flow extract . --resource ExampleApiV1
```

```mermaid
graph TD
    Entry((ENTRY))
    Entry --> Guard{{Check condition}}
    Guard -->|ok| Call[[Outbound Call]]
    Call ==> Next[example-service-b/ExampleApiV3]
    Next --> End((RETURN))
```

Shows:
- Guards
- Internal calls
- Outbound REST calls
- Explicit termination

---

## 5. Compact Flow View

```bash
jz flow extract . --resource ExampleApiV1 --format mermaid --compact
```

```mermaid
graph TD
    Entry --> Guards{{GUARDS: a && b && c}}
    Guards --> Core[Core Logic]
    Core --> End((RETURN))
```

Use `--compact` to collapse guard chains for readability.

---

## 6. Flow Diff Between Versions

```bash
jz flow diff ./v1 ./v2 --resource ExampleApiV1
```

```mermaid
flowchart TD
    A[Old Flow] -->|Removed Guard| B[Diff]
    B -->|Added Call| C[New Flow]
```

Diff guarantees:
- Ordered comparison
- No reordering tolerance
- Structural-only changes

---

## How to Read Results

```mermaid
flowchart LR
    Guard -->|fails| EarlyReturn
    Guard -->|passes| CoreLogic
    CoreLogic --> End
```

- Guards gate logic
- Early returns are explicit exits
- Unresolved calls are expected and safe

---

## Safety Guarantees

- No code execution
- No runtime assumptions
- No speculative linking
- False negatives preferred over false positives

---

## When to Use Which Command

```mermaid
flowchart TD
    Q[What do you want?]
    Q -->|Architecture| Scan[jz scan]
    Q -->|Dependencies| Report[jz report]
    Q -->|Logic| Flow[jz flow extract]
    Q -->|Regression| Diff[jz flow diff]
```

---

## Next Steps

- Add to CI for architectural drift detection
- Review flows during refactors
- Use diffs for design reviews

---

**End of Quickstart**
