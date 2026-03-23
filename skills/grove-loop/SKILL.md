---
name: grove-loop
description: >
  Autonomous development loop. Takes specifications and implements the project
  through iterative loops until production-ready.
  Trigger: When specifications are complete and ready for implementation.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
---

# GROVE Ralph Loop — Autonomous Build

## When to Use

Load this skill whenever you need to:
- Implement a project from complete specifications
- Run autonomous development loops
- Resume from interruptions
- Verify production readiness

## Overview

GROVE Ralph Loop implements projects autonomously:
1. **Validation** — Pre-flight checks
2. **Build Loop** — Implement tasks via SDD
3. **Verification** — Verify each implementation
4. **Recovery** — Handle failures gracefully

## Quick Start

```bash
grove-loop
grove-loop --pause-after task-15
grove-loop --status
grove-loop --report
```

## Output

```
GROVE-LOOP-STATE.json
GROVE-LOOP-LOG.md
GROVE-READY-REPORT.md
```
