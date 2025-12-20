# Execution Flow: ExampleApiV1

> **Analysis Mode:** AST-lite (Conservative)
> **Scope:** Single Resource Targeted Extraction

## Comparison Summary

| HTTP Method + Path | Has Guards | Early Return | Outbound Calls |
| :--- | :---: | :---: | :---: |
| `GET /v1/example` | No | No | Yes |

## Summary
Extracted 1 flow(s) for resource `ExampleApiV1`.

## Flow: GET /v1/example

### Entry

1. **ENTRY**: Enter: handleOutbound
   - **Evidence:** `testdata/flows/outbound/input/ExampleApiV1.java (start)` [confidence: high]

### Outbound Calls

2. **OUTBOUND**: Call: GET http://external-service/v1/api
   - **Evidence:** `testdata/flows/outbound/input/ExampleApiV1.java:15` [confidence: high]
   - ⚠️ *Note: This outbound call could not be resolved to a known resource.*

### Early Exit / Return

3. **RETURN**: Return: Response.ok("Done").build()
   - **Evidence:** `testdata/flows/outbound/input/ExampleApiV1.java:16` [confidence: high]

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

