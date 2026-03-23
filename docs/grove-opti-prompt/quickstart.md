# GROVE Opti Prompt — Quick Start

## Installation

```bash
cd grove
make build-all
make install
```

## Basic Usage

### 1. Optimize a Prompt

```bash
grove-opti "add login button to header"
```

Output:
```
Classified as: feature-addition (auth)
Context collected: 3 files / 847 tokens

┌─────────────────────────────────────────────────────────────┐
│ OPTIMIZED PROMPT                                            │
├─────────────────────────────────────────────────────────────┤
│ Add a login button to Header component                      │
│ @src/components/Header.tsx                                  │
│                                                             │
│ Requirements:                                               │
│ - Use Button from @ui/components                            │
│ - Include loading state                                     │
│ - Add unit tests                                            │
│                                                             │
│ Do NOT modify:                                              │
│ - @src/components/Navigation.tsx                           │
└─────────────────────────────────────────────────────────────┘

WHY: File reference ensures correct component edited.

[Accept] [Edit] [Cancel] [Save Template]
```

### 2. Batch Mode

```bash
cat > prompts.txt << 'EOF'
add dark mode toggle
fix login validation
update user profile
EOF

grove-opti --batch prompts.txt
```

### 3. From Clipboard

```bash
grove-opti --clipboard
```

## Commands

| Command | Description |
|---------|-------------|
| `grove-opti "prompt"` | Optimize prompt |
| `grove-opti --clipboard` | From clipboard |
| `grove-opti --batch <file>` | Batch mode |
| `grove-opti --explain-all` | Full explanations |
| `grove-opti --templates` | List templates |

## Learning System

GROVE Opti Prompt learns your patterns:
- 1-3 times: Full explanation
- 4-10 times: Short reminder
- 11+ times: Label only

Use `--explain-all` to force full explanations.
