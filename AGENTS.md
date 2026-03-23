# GROVE Agent Skills

## Skills Index

| Skill | Description |
|-------|-------------|
| `sdd-apply` | Implement tasks from changes, writing code following specs and design |
| `sdd-archive` | Sync delta specs to main specs and archive completed changes |
| `sdd-design` | Create technical design documents with architecture decisions |
| `sdd-explore` | Explore and investigate ideas before committing to a change |
| `sdd-init` | Initialize Spec-Driven Development context in any project |
| `sdd-propose` | Create change proposals with intent, scope, and approach |
| `sdd-spec` | Write specifications with requirements and scenarios |
| `sdd-tasks` | Break down changes into implementation task checklists |
| `sdd-verify` | Validate that implementation matches specs, design, and tasks |
| `skill-creator` | Create new AI agent skills following the Agent Skills spec |
| `go-testing` | Go testing patterns including Bubbletea TUI testing |
| `ralph-loop` | Autonomous documentation-to-code loop |

## Usage

Skills are loaded automatically based on context. When working on:
- **Go projects with tests**: `go-testing` skill is auto-loaded
- **Creating new skills**: `skill-creator` skill is auto-loaded
- **SDD workflow**: Corresponding phase skill is loaded per workflow step

## Commands

### SDD Workflow
- `/sdd-init` - Initialize SDD in project
- `/sdd-new <nombre>` - Create new change with full SDD pipeline
- `/sdd-continue` - Continue next missing artifact
- `/sdd-ff` - Fast-forward: propose → spec → design → tasks
- `/sdd-apply` - Implement tasks in batches
- `/sdd-verify` - Verify implementation against specs
- `/sdd-archive` - Archive completed change

### GROVE Commands
- `/grove-spec --input <path>` - Generate specifications from ideas/docs
- `/grove-spec --update --input <spec>` - Update existing specification
- `/grove-spec --reverse --input <code>` - Reverse engineer code to spec
- `/grove-opti <prompt>` - Optimize a natural language prompt
- `/grove-opti --batch <file>` - Batch optimize multiple prompts
- `/grove-loop --spec <spec>` - Execute autonomous build loop
- `/grove-loop --status` - Check loop status
- `/grove-loop --resume` - Resume paused loop
