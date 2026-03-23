# GROVE Examples

Examples demonstrating GROVE's main features.

## Available Examples

### 01-grove-spec-basic
Basic example of transforming ideas into structured specifications.

**Input**: Raw idea (e.g., "add JWT authentication")
**Output**: SPEC.md, DESIGN.md, TASKS.md

```bash
cd examples/01-grove-spec-basic
grove-spec --input ./input/idea.md --output ./output
```

### 02-ralph-loop
Autonomous documentation-to-code loop with reverse docs, smart defaults, and incremental updates.

**Features**:
- Reverse documentation parsing
- Smart seed data generation
- Continuous mode
- Jira/GitHub integration

```bash
cd examples/02-ralph-loop
grove-loop ./docs
```

### 03-opti-prompt
Prompt optimization for OpenCode agents.

**Input**: Natural language prompts
**Output**: Optimized, context-rich prompts

```bash
cd examples/03-opti-prompt
grove-opti --batch prompts.txt
```

## Running Examples

Each example contains:
- `README.md` - Example documentation
- `input/` - Input files
- `expected/` - Expected output (for reference)

```bash
# Navigate to an example
cd examples/<example-name>

# Run the example
<command> --input ./input/<file> --output ./output
```

## Learning Path

1. **Start with 01-grove-spec-basic**: Understand how ideas become specs
2. **Continue with 03-opti-prompt**: Learn prompt optimization
3. **Finish with 02-ralph-loop**: See the full autonomous loop in action
