# GROVE Ralph Loop — Quick Start

## Prerequisites

GROVE Ralph Loop requires a project prepared by GROVE Spec:
```
my-project/
├── spec/
│   ├── SPEC.md
│   ├── DESIGN.md
│   └── TASKS.md
├── AGENTS.md
└── .opencode/skills/
```

## Basic Usage

### 1. Navigate to Project

```bash
cd my-project
```

### 2. Start the Loop

```bash
grove-loop
```

Output:
```
GROVE Ralph Loop v1.0
────────────────────────────
🔍 Pre-flight Validation
  ✓ ROOT AGENTS.md found
  ✓ SPEC.md valid
  ✓ TASKS.md valid (24 tasks)

🔄 Starting Build Loop
  [task-1] Auth setup        [████████░░] PASS
  [task-2] Login form        [██████████] PASS
  ...

✅ Production Readiness Check
  PRODUCTION READY
  Tasks: 24/24 complete
```

### 3. Monitor Progress

```bash
grove-loop --status
```

### 4. Pause/Resume

```bash
grove-loop --pause-after task-15
# ... review progress ...
grove-loop  # Resume
```

## Commands

| Command | Description |
|---------|-------------|
| `grove-loop` | Start/resume |
| `grove-loop --status` | Show state |
| `grove-loop --report` | Final report |
| `grove-loop --pause-after <id>` | Pause |

## Output Files

```
GROVE-LOOP-STATE.json      # Checkpoint
GROVE-LOOP-LOG.md           # Audit log
GROVE-READY-REPORT.md       # Final report
```
