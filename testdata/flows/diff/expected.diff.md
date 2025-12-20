# Flow Diff: ExampleApiV1

> **Analysis Mode:** Structural Execution-Flow Diff
> **Comparison:** Ordered Step-by-Step

## Flow: GET /v1/example
Status: **MODIFIED**

### Guards
~ Modified condition:
  - PREV: Return: Response.ok("OK").build()
  - NEXT: Check: id.length() < 5

### Outbound Calls
+ Added outbound: Call: GET http://audit-service/v1/log

### Termination
~ Modified condition:
  - PREV: Return: Response.ok("OK").build()
  - NEXT: Check: id.length() < 5
+ Added return: Return: Response.status(422).build()
+ Added return: Return: Response.ok("OK").build()


