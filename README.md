# jz — Understand Large Java Backends Without Running Them

> ⚠️ **Note**: This project is **vibe coded**. It was rapidly prototyped with AI assistance. While it works, standard engineering rigor may vary.


`jz` is a **static analysis CLI tool** designed to help engineers understand **large, legacy, multi-service Java systems**—especially those built with **OSGi**, **JAX-RS**, **Ant/Maven**, and deployed on **WebSphere Liberty**.

It extracts architecture, dependencies, and workflows **without executing the system**, making it safe, fast, and suitable for unfamiliar or production-critical codebases.

---

## 🚩 The Problem

Modern enterprise Java systems often suffer from:

- 30+ microservices or bundles
- Mixed build systems (Ant + Maven)
- OSGi Declarative Services wiring spread across XML
- REST endpoints scattered across large codebases
- Poor or outdated documentation
- High onboarding cost for new engineers

When joining such a team, questions like these are hard to answer:

- What services exist?
- How do they depend on each other?
- Which service provides which interface?
- What breaks if I change this component?
- Where are the REST entry points?

`jz` exists to answer those questions **from code alone**.

---

## 🎯 What jz Does

`jz` performs **static analysis** and builds an **Intermediate Representation (IR)** of the system, from which it generates:

### Service-level insights
- OSGi bundle discovery
- Declarative Service (DS) components
- Provided and referenced interfaces
- Internal component dependency graphs

### API insights
- JAX-RS REST entry points
- HTTP method, path, handler mapping

### Runtime context (static)
- WebSphere Liberty `server.xml`
- Enabled features
- Deployed applications (without guessing runtime resolution)

### Architecture visualizations
- System-level service dependency graphs
- Component-level dependency graphs (per service)
- Mermaid diagrams for quick visualization

All of this is done **without running the application**.

---

## 🧠 Design Philosophy

`jz` is intentionally designed to be:

### Deterministic
- No inference
- No guessing
- No heuristics beyond what is explicitly encoded

### IR-first
- All analysis produces a structured IR
- Reports are generated *from* the IR

### Cleanly layered

CLI (Cobra)
↓
Application Layer (Analyze)
↓
IR + Graphs
↓
Reports (Markdown / Mermaid)

### Safe for legacy systems
- No code execution
- No classloading
- No runtime dependencies
- Read-only filesystem access

---

## 🛠️ Installation

### Option 1: Install to GOBIN (Recommended)

This compiles `jz` and places it in your `$GOPATH/bin` (usually `~/go/bin`). Ensure this directory is in your `$PATH`.

```bash
go install ./cmd/jz
```

### Option 2: Build Locally

```bash
go build -o jz ./cmd/jz
```

---

## 🚀 Usage

```
jz scan /path/to/codebase
jz report markdown /path/to/codebase
jz report mermaid /path/to/codebase
```

---

## 🎯 High-Value Flags

### --service <name>

```
jz report markdown /path/to/codebase --service com.example.orders
```

### --output <file>

```
jz report mermaid /path/to/codebase --output architecture.mmd
```

---

## 📄 License

MIT
