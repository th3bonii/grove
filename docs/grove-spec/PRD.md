# GROVE Spec — Product Requirements

## TL;DR

GROVE Spec transforms raw, unstructured project input into structured specification packages ready for AI-driven development.

## Goals

### Business Goals
- Zero ambiguity at handoff
- Reduce time-to-readiness from days to minutes
- Guarantee ecosystem conformance

### User Goals
- From chaos to clarity
- No penalties for incomplete knowledge
- Guaranteed OpenCode compatibility

## User Stories

**As a developer**, I want to drop unordered design references and get deeply broken down requirements.

**As an OpenCode agent**, I want every AGENTS.md and SKILL.md to be complete and actionable.

## Functional Requirements

### 1. Deep Input Processing
- Accept markdown, text, images, wireframes
- Extract components, states, behaviors, edge cases
- Granular decomposition of each element

### 2. Self-Questioning Loop
- Why is this needed?
- How else might it be accomplished?
- Is the proposed approach optimal?

### 3. Iteration Loop
- Quality scoring: 7 dimensions
- Loop until: All dims ≥8 AND composite ≥85 AND loop ≥2
- Resume from checkpoint

### 4. AGENTS.md Generation
- Root and scoped AGENTS.md
- Skill discovery and registration
- Intelligent merge with existing

## Quality Dimensions

| Dimension | Min Score |
|-----------|-----------|
| Flow Coverage | 8/10 |
| Component Decomposition | 8/10 |
| Logical Consistency | 8/10 |
| Inter-Component Connectivity | 8/10 |
| Edge Case Coverage | 8/10 |
| Decision Justification | 8/10 |
| Agent Consumability | 8/10 |

## CLI Commands

```bash
grove-spec --input ./ideas
grove-spec --update
grove-spec --reverse
grove-spec --loop-max 10
grove-spec --resume
```

## Output

```
/spec/
├── SPEC.md
├── DESIGN.md
├── TASKS.md
├── GROVE-SPEC-LOOP-LOG.md
├── GROVE-SPEC-LOOP-STATE.json
└── GROVE-SPEC-COMPLETE-REPORT.md
```
