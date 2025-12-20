# Contributing to jz

Thank you for your interest in contributing to **jz** 🎉  
Contributions are welcome and appreciated.

This document outlines the contribution process and the standards we follow to keep the project safe, clear, and publishable.

---

## Architecture & Design

Contributors should familiarize themselves with the internal design principles described in `docs/architecture.md` before making architectural changes.

## Guiding Principles

jz is a **static analysis tool** focused on structural accuracy, determinism, and safety.

When contributing, please keep these principles in mind:

- **Prefer false negatives over false positives**
- **Never speculate** beyond what can be proven from source code
- **Deterministic output only** (no randomness, no environment dependence)
- **Readable UX over cleverness**

---

## Fictional Examples Policy (Important)

⚠️ **All examples in this repository must be fictional.**

This includes:
- REST resource names
- Class names
- Method names
- Paths, URLs, and identifiers
- Diagrams and screenshots
- Example outputs in documentation

### ✅ Allowed Examples
- `ExampleApiV1`
- `SampleService`
- `DemoResource`
- `/api/v1/example`

### ❌ Disallowed Examples
- Names derived from real enterprise systems
- Customer-specific terminology
- Internal project names
- Anything that could imply a real production system

If in doubt, **rename it to something generic**.

This rule exists to ensure the project can be safely shared and used publicly.

---

## Code Style & Expectations

### Go Code
- Follow standard Go formatting (`gofmt`)
- Prefer clarity over abstraction
- Keep functions small and single-purpose
- Avoid reflection and runtime evaluation unless explicitly required

### Analysis Logic
- Do not introduce runtime execution or tracing
- Do not add speculative inference (e.g. guessing targets)
- All detected relationships must be backed by literal evidence

---

## Documentation Standards

- Documentation must match actual behavior
- Examples must remain fictional
- Avoid screenshots when Mermaid diagrams are sufficient
- Explain *limitations* as clearly as features

---

## Submitting Changes

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and basic CLI checks
5. Open a Pull Request with:
   - A clear description of **what** changed
   - Why the change is safe and deterministic
   - Any known limitations

---

## Reporting Issues

When filing an issue, please include:
- What command you ran
- Expected behavior
- Actual behavior
- Whether the behavior is acceptable given AST-lite limitations

---

## Code of Conduct

Be respectful and constructive.  
This project values thoughtful engineering discussions and calm technical debate.

---

Thank you for contributing to **jz** 🚀
