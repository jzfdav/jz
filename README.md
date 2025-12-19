# jz — Understand Large Java Backends Without Running Them

> ⚠️ **Note**: This project is **vibe coded**. It was rapidly prototyped with AI assistance. While it works, standard engineering rigor may vary.


`jz` is a **static analysis CLI tool** designed to help engineers understand **large, legacy, multi-service Java systems**—especially those built with **OSGi**, **JAX-RS**, **Ant/Maven**, and deployed on **WebSphere Liberty**.

It extracts architecture, dependencies, and workflows **without executing the system**, making it safe, fast, and suitable for unfamiliar or production-critical codebases.

---

## 🚀 Quick Start (Liberty WAR example)

If you are working on a repository that follows the WebSphere Liberty WAR model (non-OSGi), `jz` treats the entire repository as a single service and groups endpoints by resource class.

```bash
# Scan and view a summary
jz scan .
```

### Example Output
```markdown
# System Overview
- Total number of services: 1
- Total number of system-level dependencies: 0

## Diagnostics
- Liberty WAR service detected.
- OSGi bundles not found; modeled as a single Liberty service.

# Services
## my-web-app
- Root Path: /Users/dev/repo
- REST Entry Points: 16
- Liberty Server: defaultServer

### REST Resources
#### TenantApiV1
Base path: /api/v1/tenants
Auth: @RolesAllowed
Consumes: application/json
Produces: application/json
Path Params: tenantId

- GET     /api/v1/tenants
- POST    /api/v1/tenants
- GET     /api/v1/tenants/{tenantId}

Methods summary:
- GET: 12
- POST: 4
```

---

## 🔍 Runtime Detection Logic

`jz` uses the following flags to determine how to model your codebase:

- **HasOSGi**: Set if `META-INF/MANIFEST.MF` files are found with OSGi headers. `jz` will model each bundle as a separate service.
- **HasLiberty**: Set if a `server.xml` is detected anywhere in the tree.
- **HasLibertyWAR**: Set when **no OSGi bundles** are found, but a Liberty configuration exists with a `webApplication` entry or a `WEB-INF/web.xml` file.

**For Liberty WAR repos:**
- The repository is modeled as **one logical service**.
- REST resources are grouped by their implementation class.
- System-level dependencies are usually empty (unless multiple services are detected).

---

## 🧠 AST-lite Scanning

`jz` avoids the overhead of a full Java parser or the fragility of raw regex by using a line-based "AST-lite" approach:

- **Line-by-line scanning**: Reads Java files to find JAX-RS annotations and class declarations.
- **No constant resolution**: Does not resolve constants (e.g., `@Path(Constants.BASE)` is not resolved).
- **No array-based detection**: `@Consumes({"a", "b"})` is not currently parsed.
- **No inheritance**: Class-level metadata is extracted from the immediate source file only.
- **Deterministic**: Results are stable and never inferred or guessed.
- **Safe**: Does not load your classes or execute bytecode.

---

## 📊 Visualizations

Generated via `jz report mermaid .`.

### System Graph
For Liberty WAR services, these appear as a single node labeled with `(WAR)`.

```mermaid
graph TD
    my_app[my-app (WAR)]
```

### Component Graph
For OSGi services, `jz` visualizes internal Declarative Service (DS) wiring.

---

## 🛑 What jz Does NOT do (yet)

- **No inter-resource call graphs**: Cannot yet show how one REST resource calls another.
- **No auth propagation**: Does not track if a user’s role is checked across service boundaries.
- **No schema modeling**: Does not parse POJOs for request/response bodies.
- **No dependency inference**: Does not guess service dependencies based on imports or library usage.

---

## 🛠️ Installation

### Option 1: Install to GOBIN (Recommended)

```bash
go install ./cmd/jz
```

### Option 2: Build Locally

```bash
mkdir -p bin
go build -o bin/jz ./cmd/jz
```

---

## 🚀 Usage

```bash
jz scan .                          # Quick Markdown summary
jz report markdown .               # Full Markdown report
jz report mermaid .                # Generate Mermaid diagrams
```

Both `report` commands support `--service` to filter by name and `--output` to save to a file.

---

## 🔮 Phase F4 – Planned: Cross-Resource and Cross-Service Analysis

*Status: Not implemented yet.*

Upcoming goals include:
- **Call graph inference**: Visualizing interactions between different REST resources.
- **Internal Boundary Detection**: Identifying hidden service boundaries inside giant WAR files.
- **Auth Propagation**: Tracking security context across service calls for risk visibility.
- **Multi-service Liberty support**: Better modeling of EAR files and sidecar deployments.

Analysis will remain incremental and opt-in, preserving the tool's performance and safety.

---

## 📄 License

MIT
