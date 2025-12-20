# Execution Flow: ExampleApiV1

> **Analysis Mode:** AST-lite (Conservative)
> **Scope:** Single Resource Targeted Extraction

## Comparison Summary

| HTTP Method + Path | Has Guards | Early Return | Outbound Calls |
| :--- | :---: | :---: | :---: |
| `POST /v1/example` | Yes | Yes | No |

> ℹ️ **Note:** No outbound REST calls detected in any analyzed handlers for this resource.

## Summary
Extracted 1 flow(s) for resource `ExampleApiV1`.

## Flow: POST /v1/example

### Entry

1. **ENTRY**: Enter: handleGuards
   - **Evidence:** `testdata/flows/guards/input/ExampleApiV1.java (start)` [confidence: high]

### Guard Conditions

2. **CONDITION**: **Guard:** Check: input == null
   - **Evidence:** `testdata/flows/guards/input/ExampleApiV1.java:12` [confidence: medium]

### Early Exit / Return

3. **RETURN**: Return: Response.status(400).build()
   - **Evidence:** `testdata/flows/guards/input/ExampleApiV1.java:13` [confidence: high]

### Guard Conditions

4. **CONDITION**: **Guard:** Check: input.isEmpty()
   - **Evidence:** `testdata/flows/guards/input/ExampleApiV1.java:16` [confidence: medium]

### Early Exit / Return

5. **RETURN**: Return: Response.status(422).build()
   - **Evidence:** `testdata/flows/guards/input/ExampleApiV1.java:17` [confidence: high]

6. **RETURN**: Return: Response.ok("Valid").build()
   - **Evidence:** `testdata/flows/guards/input/ExampleApiV1.java:20` [confidence: high]

_No outbound REST calls detected in this handler._

> ✅ **End Note:** Flow completed with a detected return statement.

## Observations

### Gating & Guardrails
- Flow `POST /v1/example` is gated by: `Check: input == null`
- Flow `POST /v1/example` is gated by: `Check: input.isEmpty()`

### Early Exits
- Flow `POST /v1/example` has an early exit: `Return: Response.status(400).build()`
- Flow `POST /v1/example` has an early exit: `Return: Response.status(422).build()`

## Limitations (AST-lite)
- Logic is extracted via line-based lexical analysis.
- Data propagation across variables or loops is not tracked.
- Complex boolean expressions may be truncated.
- Only same-file internal methods are expanded.

