# GROVE Opti Prompt — Product Requirements

## TL;DR

GROVE Opti Prompt converts natural language requests into precise, context-rich prompts for OpenCode agents while teaching better prompting techniques.

## Goals

### Business Goals
- Eliminate generic prompts
- Reduce rework rate to <10%
- Keep tokens <2000 per optimization

### User Goals
- Transform plain language into precise prompts
- Understand why each addition was made
- Retain full editorial control

## Functional Requirements

### 1. Intent Classification
- Detect: feature-addition, bug-fix, refactor, documentation, config-change
- Extract domain and keywords
- Calculate confidence score

### 2. Context Collection (4-Layer)
```
Layer 1: AGENTS.md explicit references (highest)
Layer 2: Recent git commits + keywords
Layer 3: Path keyword matching
Layer 4: SPEC.md component references (lowest)
```

### 3. Prompt Optimization
- Add @file references
- Add skill() calls
- Add success criteria
- Add scope boundaries
- Add risk warnings

### 4. Teaching Layer
- Adaptive explanations based on times_seen
- Bidirectional learning from edits
- Template saving

## CLI Commands

```bash
grove-opti "add login button"
grove-opti --clipboard
grove-opti --batch prompts.txt
grove-opti --explain-all
grove-opti --max-tokens 3000
```

## Output

```
GROVE-OPTI-LOG.md
GROVE-OPTI-BATCH-<timestamp>.md
```
