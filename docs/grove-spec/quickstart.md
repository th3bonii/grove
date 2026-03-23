# GROVE Spec — Quick Start

## Installation

```bash
cd grove
make build-all
make install
```

## Basic Usage

### 1. Prepare Your Ideas

```bash
mkdir my-project/ideas
cat > my-project/ideas/notes.md << 'EOF'
# My App

## Features
- User authentication
- Dashboard
- Settings

## Tech
- React + TypeScript
EOF
```

### 2. Generate Specifications

```bash
grove-spec --input ./my-project/ideas
```

Output:
```
GROVE Spec v1.0
────────────────────────────
📥 Phase 1: Ingestion & Analysis
  ✓ Components extracted: 3

🔄 Phase 3: Iteration Loop
  Iteration 1/10: Quality score: 72/100
  Iteration 2/10: Quality score: 89/100 ✓

📄 Phase 4: Specification Generation
  ✓ SPEC.md generated
  ✓ DESIGN.md generated
  ✓ TASKS.md generated

✅ Spec generation complete!
```

### 3. Review Output

```bash
ls my-project/spec/
cat my-project/spec/GROVE-SPEC-COMPLETE-REPORT.md
```

## Modes

| Mode | Command | Use Case |
|------|---------|----------|
| Standard | `grove-spec --input ./ideas` | New project |
| Update | `grove-spec --update` | Add to existing |
| Reverse | `grove-spec --reverse` | From code |

## Next Steps

After Spec completes, run Ralph Loop:
```bash
cd my-project
grove-loop
```
