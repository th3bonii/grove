# GROVE Opti Prompt — Documentation

## Overview

GROVE Opti Prompt optimizes natural language prompts for OpenCode.

## Quick Start

```bash
grove-opti "add login button to header"
```

## Files

| File | Description |
|------|-------------|
| PRD.md | Product Requirements Document |
| README.md | This file |
| quickstart.md | Getting started guide |

## Output

```
GROVE-OPTI-LOG.md                    # Session log
GROVE-OPTI-BATCH-<timestamp>.md      # Batch output
```

## 4-Layer File Selection

1. AGENTS.md explicit references (highest)
2. Recent git commits + keywords
3. Path keyword matching
4. SPEC.md component references (lowest)
