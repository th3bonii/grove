# GROVE

> Spec-Driven Development framework for AI agents

GROVE (Genesis Reimagined Over Vast Ecosystems) es un framework de desarrollo inteligente que conecta la intención humana con la implementación de código a través de Spec-Driven Development (SDD).

## Features

- **Spec-Driven Development**: Define what to build, let AI handle the how
- **Structured Workflow**: Proposal → Spec → Design → Tasks → Apply → Verify → Archive
- **Persistent Memory**: Cross-session context with engram integration
- **Multi-Project Support**: Works with any language/framework
- **Prompt Optimization**: Transform natural language into precise prompts
- **Autonomous Loop**: Build projects from specifications automatically

## Installation

```bash
# Clone the repository
git clone https://github.com/Gentleman-Programming/grove.git
cd grove

# Build from source
go build -o bin/ ./cmd/...

# Or use pre-built binaries
unzip grove-windows.zip   # Windows
unzip grove-darwin-amd64.zip  # macOS
unzip grove-linux-amd64.zip  # Linux
```

## Commands

### grove-spec
Convierte ideas en bruto en especificaciones estructuradas.

```bash
# Basic usage
grove-spec --input ./my-ideas

# With output directory
grove-spec --input ./idea.md --output ./specs

# With options
grove-spec --input ./idea.md --model gpt-4 --quality-gate
```

### grove-opti
Optimiza prompts de lenguaje natural para OpenCode.

```bash
# Single prompt
grove-opti "add login button to header"

# Batch mode
grove-opti --batch prompts.txt

# With context
grove-opti "fix auth bug" --context ./src/auth
```

### grove-loop
Construye proyectos autonomía desde especificaciones.

```bash
# Run with spec file
grove-loop --spec ./SPEC.md

# With checkpoint recovery
grove-loop --spec ./SPEC.md --resume

# Dry run
grove-loop --spec ./SPEC.md --dry-run
```

## SDD Workflow

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│ explore │ -> │ propose │ -> │  spec   │ -> │ design  │
└─────────┘    └─────────┘    └─────────┘    └─────────┘
                                                    │
                                                    v
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│ archive │ <- │ verify  │ <- │  apply  │ <- │  tasks  │
└─────────┘    └─────────┘    └─────────┘    └─────────┘
```

| Phase | Description |
|-------|-------------|
| explore | Investigate ideas and clarify requirements |
| propose | Create change proposal with intent and scope |
| spec | Write specifications with requirements and scenarios |
| design | Create technical design document |
| tasks | Break down into implementation checklist |
| apply | Implement tasks following specs and design |
| verify | Validate implementation matches specs |
| archive | Sync delta specs and archive completed change |

## Examples

See [examples/README.md](examples/README.md) for usage examples.

| Example | Description |
|---------|-------------|
| [01-grove-spec-basic](examples/01-grove-spec-basic) | Transform ideas into specs |
| [02-ralph-loop](examples/02-ralph-loop) | Autonomous documentation-to-code |
| [03-opti-prompt](examples/03-opti-prompt) | Optimize prompts |

## Architecture

```
proposal -> specs -> tasks -> apply -> verify -> archive
              ^
            design
```

## Development

```bash
# Install dependencies
make install

# Build
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## License

Apache License 2.0 - See [LICENSE](LICENSE) file
