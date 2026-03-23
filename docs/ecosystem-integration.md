# GROVE Ecosystem — Integration Guide

## Philosophy

GROVE is a **complement** to the gentle-ai ecosystem, not a replacement. Every tool respects the configuration that gentle-ai already established and works within that environment — never on top of it.

```
┌─────────────────────────────────────────────────────────────┐
│                    gentle-ai                               │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────┐  │
│  │   Engram    │  │ SDD Skills  │  │       GGA        │  │
│  │   Memory    │  │ sdd-apply   │  │  Code Review     │  │
│  │   Cross-ses │  │ sdd-verify  │  │  Pre-commit      │  │
│  └─────────────┘  └─────────────┘  └──────────────────┘  │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────┐  │
│  │     MCP     │  │   Persona   │  │   Permissions    │  │
│  │  Context7   │  │  Gentleman  │  │  Security-first  │  │
│  │  Notion/Jira│  │  Neutral    │  │  Guardrails      │  │
│  └─────────────┘  └─────────────┘  └──────────────────┘  │
│                                                             │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    GROVE (Complement)                      │
│                                                             │
│  ┌─────────────┐    ┌─────────────┐    ┌──────────────┐  │
│  │ GROVE Spec  │───▶│GROVE Ralph │───▶│GROVE Opti   │  │
│  │             │    │   Loop      │    │   Prompt     │  │
│  │ Ideas →     │    │ Specs →     │    │ Prompt →     │  │
│  │ Specs       │    │ Code Ready  │    │ Better       │  │
│  └─────────────┘    └─────────────┘    └──────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

## What GROVE Reads from gentle-ai

| Component | What GROVE Reads | Purpose |
|-----------|-----------------|---------|
| `~/.config/opencode/skills/` | Skill names and descriptions | Include in discovery |
| `~/.config/opencode/AGENTS.md` | Global constraints | Respect in optimization |
| Active preset | Available components | Feature availability |
| Engram config | Connection settings | Cross-session memory |
| GGA status | Whether initialized | Initialize if needed |
| Persona active | Behavior style | Maintain consistency |

## What GROVE NEVER Modifies

| Component | Action |
|-----------|--------|
| Global gentle-ai config | ❌ Never |
| Global skills directory | ❌ Never |
| Global AGENTS.md | ❌ Never |
| Existing Engram entries | ❌ Never (only append) |
| Security permissions | ❌ Never |
| Persona configuration | ❌ Never |
| Preset configuration | ❌ Never |

## What GROVE Creates (Project Only)

```
/project/
├── spec/                          # GROVE Spec output
│   ├── SPEC.md
│   ├── DESIGN.md
│   ├── TASKS.md
│   └── GROVE-SPEC-*.{json,md}
│
├── AGENTS.md                      # Created or merged
├── [module]/AGENTS.md            # Scoped configs
├── .opencode/skills/             # Project-specific skills
│
├── GROVE-LOOP-STATE.json        # Loop checkpoint
├── GROVE-LOOP-LOG.md            # Audit trail
├── GROVE-READY-REPORT.md        # Final report
│
├── GROVE-OPTI-LOG.md            # Optimization log
│
└── [application code]            # Implemented code
```

## Pipeline Flow

### Phase 1: GROVE Spec (Idea → Specs)

```
User Input                    GROVE Spec Output
─────────────                ─────────────────
Wireframes ──┐               ┌─────────────────┐
Notes       ├──▶ GROVE Spec ──▶│ SPEC.md         │
References  │               │ │ DESIGN.md       │
Prompts     │               │ │ TASKS.md        │
Images      ──┘               │ │ AGENTS.md       │
                              │ │ SKILLS.md       │
                              │ │ Quality Report  │
                              └─────────────────┘
                                      │
                                      ▼
                              Ready for Loop
```

### Phase 2: GROVE Ralph Loop (Specs → Code)

```
Specs Input                   GROVE Loop Output
─────────────                ─────────────────
SPEC.md      ──┐             ┌─────────────────┐
DESIGN.md    ├──▶ Ralph Loop ──▶│ Production Code │
TASKS.md     │             │ │ State Checkpoint│
AGENTS.md    ──┘             │ │ Audit Log       │
                             │ │ Ready Report    │
                             └─────────────────┘
                                     │
                                     ▼
                             If docs insufficient:
                             ┌─────────────────┐
                             │ Call GROVE Spec │
                             │ with feedback   │
                             └─────────────────┘
```

### Phase 3: GROVE Opti Prompt (Daily Driver)

```
User Prompt                   GROVE Opti Output
─────────────                ─────────────────
"add login   ──┐             ┌─────────────────┐
 button"      ├──▶ Opti    ──▶│ Optimized Prompt│
Natural      │  Prompt     │ │ @file refs      │
Language     ──┘             │ │ skill() calls   │
                             │ │ boundaries      │
                             │ │ explanations    │
                             └─────────────────┘
                                     │
                                     ▼
                             [User Review]
                             ┌─────────────────┐
                             │ [Accept] [Edit] │
                             │   [Cancel]      │
                             └─────────────────┘
                                     │
                                     ▼
                             [Send to OpenCode]
```

## GGA Integration

GROVE Spec can initialize GGA if available:

```bash
# Check if GGA is available
gga --version

# If not initialized
gga init
gga install

# GGA now validates commits automatically
```

GROVE respects GGA as AI provider — never tries to change it.

## Engram Integration

GROVE uses Engram for cross-session memory:

```
Before spec loop:
- Read: grove-spec/{stack}/decisions
- Read: grove-spec/{stack}/gap-patterns

After spec loop:
- Write: grove-spec/{stack}/decisions
- Write: grove-spec/{stack}/gap-patterns

During loop:
- Write: grove-ralph/{project}/state
- Write: grove-ralph/benchmarks/{stack}
```

If Engram unavailable → graceful degradation (warning, continue).

## Token Efficiency

GROVE Opti Prompt is designed for minimal token usage:

| Component | Budget |
|-----------|--------|
| Core request | ~100 tokens |
| Context files | ~500 tokens |
| Quality requirements | ~200 tokens |
| References | ~300 tokens |
| **Total** | **~1250 tokens** |

Sub-agent approach: Context collector runs independently with bounded context.

## Conflict Resolution

If GROVE detects potential conflict:

1. **Skill name collision** → Prefix with `grove-`
2. **AGENTS.md conflict** → Intelligent merge, never overwrite
3. **Engram key collision** → Use `grove-` prefix namespace

All conflicts documented in completion reports.

---

## Summary

GROVE tools are **complements** that enhance the gentle-ai ecosystem:

- ✅ Read from gentle-ai configuration
- ✅ Respect existing setup
- ✅ Use SDD skills
- ✅ Integrate with Engram
- ✅ Work with GGA
- ❌ Never modify global config
- ❌ Never overwrite existing files
- ❌ Never break existing workflows
