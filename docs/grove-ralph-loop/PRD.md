# GROVE Ralph Loop — Product Requirements

## TL;DR

GROVE Ralph Loop is an autonomous development execution engine. It consumes specifications and builds projects through iterative loops until production-ready.

## Goals

### Business Goals
- Deliver production-ready projects autonomously
- Zero user intervention post-initiation
- Decrease time-to-production from days to hours

### User Goals
- Run once, receive finished project
- Leave process unattended
- Never debug loop failures

## Functional Requirements

### 1. Pre-Loop Validation
- ROOT AGENTS.md validation
- SKILLS.md validation
- SPEC.md, DESIGN.md, TASKS.md existence
- Tech stack coherence
- Dependency resolution

### 2. Build Loop
```
For each task:
1. Load next unimplemented task
2. Spawn sub-agent with scoped context
3. Sub-agent implements task
4. Spawn verify sub-agent (must invoke sdd-verify)
5. If PASS: mark complete
6. If FAIL: retry once, then flag
7. Checkpoint state
```

### 3. Resilience
- State saved after every task
- LLM failure recovery with exponential backoff
- Pause/Resume support

### 4. Production Readiness
- All specs implemented
- All tests passing
- GGA validation passing
- No runtime errors

## CLI Commands

```bash
grove-loop
grove-loop --pause-after task-15
grove-loop --status
grove-loop --report
grove-loop --stop
```

## Output

```
GROVE-LOOP-STATE.json
GROVE-LOOP-LOG.md
GROVE-READY-REPORT.md
```
