---
name: git-commit
description: Enforces Conventional Commits format with type prefixes, scopes, and Signed-off-by for kubectl-mtv. Use when committing changes, creating commits, or writing commit messages.
---

# Git Commit Convention

All commits must follow the Conventional Commits format with a DCO sign-off.

## Format

```
<type>(<scope>): <short description>

<body - explain WHY, not WHAT>

Signed-off-by: Name <email>
```

## Commit Command

Always sign off with `-s` and include a body:

```bash
git commit -s -m "$(cat <<'EOF'
fix(mcp): handle nil response from inventory API

Prevents panic when provider inventory returns empty body.
EOF
)"
```

## Type Prefixes

| Type | When to use |
|------|-------------|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `refactor` | Code restructuring, no behavior change |
| `docs` | Documentation only |
| `test` | Adding or updating tests |
| `ci` | CI/CD changes (workflows, Makefile targets) |
| `build` | Build system, dependencies, container images |
| `chore` | Maintenance (deps update, cleanup) |
| `perf` | Performance improvement |
| `style` | Formatting, linting (no logic change) |

## Scope (Optional)

Area of the codebase in parentheses after the type:

| Scope | Area |
|-------|------|
| `cli` | CLI commands (`cmd/`, `pkg/cmd/`) |
| `mcp` | MCP server and tools (`pkg/mcp/`, `cmd/mcpserver/`) |
| `inventory` | Inventory API and provider client |
| `guide` | Technical guide (`guide/`) |
| `deploy` | Deployment manifests (`deploy/`) |
| `tui` | Terminal UI (`pkg/util/tui/`) |
| `e2e` | End-to-end tests (`e2e/`) |

Omit scope for cross-cutting changes.

## Title Rules

- Imperative mood: "add", "fix", "update" (not "added", "fixes", "updates")
- Lowercase after the colon
- Max ~72 characters
- No trailing period

## Body

Required. Explain **why** the change was made, not what changed (the diff shows that). Separate from the title with a blank line. Even for small changes, provide a brief rationale.

## Signed-off-by

Every commit must include `Signed-off-by: Name <email>` (DCO sign-off). Use `git commit -s` to add it automatically.

## AI Co-authorship

When an AI agent contributes to a commit, **ask the user** whether to add a co-authorship trailer before committing. There are two levels:

1. **Full co-authorship** -- the AI agent helped write production code (features, fixes, refactoring, etc.):

```
Co-authored-by: AI Agent <noreply@github.com>
```

2. **Unit-test co-authorship** -- the AI agent only generated or helped write unit tests:

```
Co-authored-by: AI Agent (unit tests) <noreply@github.com>
```

Place the `Co-authored-by` line after the `Signed-off-by` line. If the user declines, omit it entirely.

## Examples

```
feat(cli): add describe hook command

Allow users to inspect hook details including playbook
content and execution phase.

Signed-off-by: Yaacov Zamir <yzamir@redhat.com>
```

```
fix(mcp): handle nil response from inventory API

Prevents panic when provider inventory returns empty body.

Signed-off-by: Yaacov Zamir <yzamir@redhat.com>
Co-authored-by: AI Agent <noreply@github.com>
```

```
docs(guide): add chapter on warm migration workflow

Covers warm migration lifecycle, cutover timing, and
snapshot scheduling for vSphere providers.

Signed-off-by: Yaacov Zamir <yzamir@redhat.com>
```

```
refactor(tui): extract table rendering into reusable component

Reduces duplication across get commands and makes it
easier to add new table-based views.

Co-authored-by: AI Agent <noreply@github.com>

Signed-off-by: Yaacov Zamir <yzamir@redhat.com>
```

```
test(cli): add unit tests for describe hook command

Covers edge cases for missing hooks, empty playbooks,
and invalid hook phases.

Co-authored-by: AI Agent (unit tests) <noreply@github.com>

Signed-off-by: Yaacov Zamir <yzamir@redhat.com>
```

```
build: update forklift dependency to latest v1beta1

Picks up new DynamicProvider CRD types needed for
custom provider support.

Signed-off-by: Yaacov Zamir <yzamir@redhat.com>
```

## Breaking Changes

Add `!` after type/scope and include `BREAKING CHANGE:` in the body:

```
feat(mcp)!: rename mtv_query tool to mtv_read

BREAKING CHANGE: mtv_query has been renamed to mtv_read.
Update all MCP client configurations.

Signed-off-by: Yaacov Zamir <yzamir@redhat.com>
```
