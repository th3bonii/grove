# GROVE

> Spec-Driven Development framework for AI agents

GROVE (Genesis Reimagined Over Vast Ecosystems) is an intelligent development framework that bridges human intent with code implementation through Spec-Driven Development (SDD).

## Features

- **Spec-Driven Development**: Define what to build, let AI handle the how
- **Structured Workflow**: Proposal → Spec → Design → Tasks → Apply → Verify → Archive
- **Persistent Memory**: Cross-session context with engram integration
- **Multi-Project Support**: Works with any language/framework

## Quick Start

```bash
# Initialize SDD in your project
make sdd-init

# Create a new change
make sdd-new feature-name

# Fast-forward through planning
make sdd-ff

# Implement tasks
make sdd-apply

# Verify implementation
make sdd-verify
```

## Architecture

```
proposal → specs → tasks → apply → verify → archive
              ↑
            design
```

## Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the application |
| `make test` | Run all tests |
| `make clean` | Remove build artifacts |
| `make install` | Install dependencies |
| `make help` | Show available targets |

## License

Apache License 2.0 - See [LICENSE](LICENSE) file
