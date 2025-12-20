# Execution Flow: ExampleApiV1

> **Analysis Mode:** AST-lite (Conservative)
> **Scope:** Single Resource Targeted Extraction

## Comparison Summary

| HTTP Method + Path | Has Guards | Early Return | Outbound Calls |
| :--- | :---: | :---: | :---: |
| `GET /v1/example` | No | No | No |

> ℹ️ **Note:** No outbound REST calls detected in any analyzed handlers for this resource.

## Summary
Extracted 1 flow(s) for resource `ExampleApiV1`.

## Flow: GET /v1/example

### Entry

1. **ENTRY**: Enter: getSimple
   - **Evidence:** `testdata/flows/simple/input/ExampleApiV1.java (start)` [confidence: high]

### Early Exit / Return

2. **RETURN**: Return: Response.ok("Hello").build()
   - **Evidence:** `testdata/flows/simple/input/ExampleApiV1.java:12` [confidence: high]

_No outbound REST calls detected in this handler._

> ✅ **End Note:** Flow completed with a detected return statement.

## Observations

### Gating & Guardrails
- No explicit gating conditions detected.

### Early Exits
- No early exits detected.

## Limitations (AST-lite)
- Logic is extracted via line-based lexical analysis.
- Data propagation across variables or loops is not tracked.
- Complex boolean expressions may be truncated.
- Only same-file internal methods are expanded.

