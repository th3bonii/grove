# GROVE Spec — Documentation

## Overview

GROVE Spec transforms raw ideas into structured specifications.

## Quick Start

```bash
grove-spec --input ./my-ideas
```

## Files

| File | Description |
|------|-------------|
| PRD.md | Product Requirements Document |
| README.md | This file |
| quickstart.md | Getting started guide |

## Output

```
/spec/
├── SPEC.md           # Product requirements
├── DESIGN.md         # Technical architecture
├── TASKS.md          # Implementation tasks
└── GROVE-SPEC-*.md   # Reports and logs
```

## Quality Scoring

7 dimensions, each scored 0-10:
- Flow Coverage
- Component Decomposition
- Logical Consistency
- Inter-Component Connectivity
- Edge Case Coverage
- Decision Justification
- Agent Consumability

**Exit:** All ≥8 AND composite ≥85 AND loop ≥2
