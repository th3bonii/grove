---
name: grove-opti-prompt
description: >
  Prompt optimization. Converts natural language requests into precise,
  context-rich prompts for OpenCode agents.
  Trigger: When user wants to optimize a prompt before sending to AI agent.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
---

# GROVE Opti Prompt — Prompt Optimization

## When to Use

Load this skill whenever you need to:
- Optimize a prompt before sending to OpenCode
- Add project context to requests
- Learn prompt engineering patterns

## Overview

GROVE Opti Prompt enhances prompts:
1. **Capture** — Get user's natural language
2. **Context** — Collect relevant project files
3. **Optimize** — Add @file refs, skill() calls
4. **Explain** — Teach why each addition matters

## Quick Start

```bash
grove-opti "add login button to header"
grove-opti --clipboard
grove-opti --batch prompts.txt
```

## Output

```
GROVE-OPTI-LOG.md
```
