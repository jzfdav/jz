# AI Task Constraints — Code Modification

This document applies to all tasks that MODIFY existing code.

---

## General Rules for Revisions

When revising existing files, the AI MUST:

1. Treat the task as corrective, not creative
2. Change ONLY what is explicitly requested
3. Preserve all existing structure unless told otherwise
4. Avoid refactoring, reformatting, or cleanup
5. Avoid renaming identifiers unless explicitly instructed
6. Avoid introducing helper functions or abstractions
7. Keep changes minimal and localized
8. Maintain existing function signatures
9. Maintain existing file boundaries

---

## Prohibited Behaviors During Revisions

The AI MUST NOT:

- Add new features
- Add new abstractions
- Add new files
- Add comments explaining changes
- Add TODOs or future ideas
- “Improve” code quality beyond the request
- Reorganize imports or code blocks
- Introduce defensive logic not requested

---

## If Uncertain

If the requested change is ambiguous:

- STOP
- Ask a clarification question
- Do NOT guess

This document is mandatory for all revision tasks.
