---
name: grove-spec
description: >
  Idea-to-specification workflow. Converts raw ideas, wireframes, visual references,
  and unstructured documentation into complete specification documents.
  Trigger: When user wants to convert ideas into structured specifications.
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
---

# GROVE Spec — Idea to Specification

## When to Use

Load this skill whenever you need to:
- Transform raw ideas into complete specifications
- Process wireframes or visual references
- Create documentation ready for AI development
- Prepare a project for GROVE Ralph Loop

## Overview

GROVE Spec ingests raw input and produces structured specifications through:
1. **Ingestion** — Parse ideas, wireframes, docs
2. **Questioning** — Self-analyze for gaps
3. **Generation** — Create SPEC.md, DESIGN.md, TASKS.md
4. **Iteration** — Loop until quality threshold met

## Quick Start

```bash
grove-spec --input ./my-ideas
grove-spec --update
grove-spec --reverse
```

## Output

```
/spec/
├── SPEC.md
├── DESIGN.md
├── TASKS.md
├── GROVE-SPEC-LOOP-LOG.md
└── GROVE-SPEC-COMPLETE-REPORT.md
```
