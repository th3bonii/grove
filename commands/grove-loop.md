---
description: Execute autonomous build loop from specifications
agent: gentleman
---

Execute the GROVE Ralph Loop to build from specifications.

$ARGUMENTS

Follow the Ralph Loop workflow:
1. Validate documentation (SPEC.md, DESIGN.md, TASKS.md)
2. Run quality gate - score documentation
3. Execute tasks from TASKS.md via SDD sub-agents
4. Verify each implementation
5. Generate GROVE-READY-REPORT when complete
