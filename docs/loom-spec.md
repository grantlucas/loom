# Loom

**A read-only TUI for humans to monitor Beads projects.**

Beads are threaded onto a loom — Loom is where you see the full picture.

```
loom — the human side of bd
```

---

## Why "Loom"

Beads go on a loom. The name is short, memorable, and evokes seeing the interconnected structure of work. It also draws a clean conceptual boundary: `bd` is the agent's tool for managing work, `loom` is the human's tool for understanding it. The CLI invocation is just `loom`.

Alternative names considered: `strand`, `thread`, `glass`, `abacus`. Loom wins because it implies both *structure* and *visibility* — you don't just see individual beads, you see the weave.

---

## Design Philosophy

1. **Read-only.** Loom never writes to the Beads database. You manage issues through your agent; Loom is your window into what's happening.
2. **CLI-native data source.** Loom shells out to `bd` commands with `--json` flags rather than linking against Dolt or importing Beads internals. This keeps Loom decoupled from Beads' storage layer and immune to the kind of breakage that killed beads_viewer when Beads migrated from SQLite to Dolt.
3. **Focused scope.** No augmentation of Beads, no alternative workflows, no agent features. Just a clean, fast way for a human to answer: *What's the state of my project?*
4. **Keyboard-driven.** Full navigation via keyboard. Mouse support is a nice-to-have but not a design driver.

---

## Architecture

### Technology

| Component | Choice | Rationale |
|---|---|---|
| Language | Go | Matches Beads ecosystem, your active learning interest, single binary distribution |
| TUI framework | [Bubble Tea](https://github.com/charmbracelet/bubbletea) | Elm-architecture TUI, excellent composability |
| Styling | [Lip Gloss](https://github.com/charmbracelet/lipgloss) | Declarative terminal styling from the same ecosystem |
| Components | [Bubbles](https://github.com/charmbracelet/bubbles) | Pre-built table, viewport, text input, spinner, tabs |
| Data source | `bd` CLI (`--json` output) | Decoupled from Beads internals; survives storage backend changes |

### Data Layer

Loom communicates with Beads exclusively through the `bd` CLI's JSON interface:

```
bd list --json                       → all issues with metadata
bd show <id> --json                  → single issue detail with deps/dependents/comments
bd ready --json --explain            → ready queue with blocking explanations
bd dep tree <id>                     → dependency tree (text, parsed)
bd dep cycles                        → cycle detection
bd list --status open --json         → filtered views
bd list --priority 0 --json          → priority filtering
```

The data layer is a Go package (`internal/datasource`) that wraps these CLI calls, parses JSON output, and exposes typed structs. This package is the single integration point — if Beads' JSON schema changes, only this package needs updating.

```
internal/
  datasource/
    bd.go           # CLI executor: runs bd commands, captures stdout
    parser.go       # JSON → typed Go structs
    types.go        # Issue, Dependency, Comment, Label structs
    cache.go        # Optional in-memory cache with TTL
```

**Key data types** (derived from `bd show --json` output):

```go
type Issue struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`      // open, in_progress, closed
    Priority    int       `json:"priority"`     // 0=critical .. 4=backlog
    IssueType   string    `json:"issue_type"`   // bug, feature, task, epic, chore
    Assignee    string    `json:"assignee"`
    Owner       string    `json:"owner"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    CreatedBy   string    `json:"created_by"`
    Parent      string    `json:"parent"`       // hierarchical parent (e.g. bd-a3f8)
}

type Dependency struct {
    ID             string `json:"id"`
    Title          string `json:"title"`
    Status         string `json:"status"`
    Priority       int    `json:"priority"`
    DependencyType string `json:"dependency_type"` // blocks, relates_to, duplicates, etc.
}

type IssueDetail struct {
    Issue        Issue        `json:"issue"`
    Labels       []string     `json:"labels"`
    Dependencies []Dependency `json:"dependencies"`  // what this issue depends on
    Dependents   []Dependency `json:"dependents"`     // what depends on this issue
    Comments     []Comment    `json:"comments"`
}
```

### Refresh Strategy

Two modes:

1. **Manual refresh** (default): Press `r` to re-fetch data from `bd`. Fast, predictable, no background processes.
2. **Watch mode** (opt-in via `--watch` or `w` key): Poll `bd list --json` on a configurable interval (default 5s). Useful when your agent is actively working and you want to see progress in near-real-time.

No filesystem watchers, no Dolt hooks, no daemons. Just CLI calls.

---

## Views

Loom has five views, navigable via tab bar or keyboard shortcuts.

### 1. Dashboard (`d`)

The landing view. At-a-glance project health.

```
┌─ Loom ─────────────────────────────────────────────────────┐
│ [Dashboard]  Issues  Detail  Tree  Critical Path           │
├────────────────────────────────────────────────────────────-┤
│                                                            │
│  Project: my-project    Issues: 47    Last refresh: 12s    │
│                                                            │
│  ── Status ──────────  ── Priority ─────────────────────   │
│  Open          18      P0 Critical    2  ██                │
│  In Progress    6      P1 High        8  ████████          │
│  Closed        23      P2 Medium     21  █████████████     │
│                        P3 Low        12  ██████████        │
│                        P4 Backlog     4  ████              │
│                                                            │
│  ── Ready Queue (5) ────────────────────────────────────   │
│  bd-a1b2  [P1] Fix auth validation          task           │
│  bd-c3d4  [P2] Add retry logic              feature        │
│  bd-e5f6  [P2] Update API docs              chore          │
│  bd-g7h8  [P3] Refactor config loader       task           │
│  bd-i9j0  [P3] Add telemetry hooks          feature        │
│                                                            │
│  ── Blocked (3) ────────────────────────────────────────   │
│  bd-k1l2  [P0] Deploy hotfix     ← blocked by bd-a1b2     │
│  bd-m3n4  [P1] Integration tests ← blocked by bd-c3d4     │
│  bd-o5p6  [P2] Release v2.1      ← blocked by bd-k1l2     │
│                                                            │
│  Longest chain: 3 deep (bd-o5p6 → bd-k1l2 → bd-a1b2)      │
│                                                            │
├────────────────────────────────────────────────────────────-┤
│ d:dashboard  i:issues  t:tree  c:critical  r:refresh  ?:help│
└────────────────────────────────────────────────────────────-┘
```

Data sources: `bd list --json`, `bd ready --json --explain`

### 2. Issue List (`i`)

Filterable, sortable table of all issues.

```
┌─ Loom ─────────────────────────────────────────────────────┐
│  Dashboard  [Issues]  Detail  Tree  Critical Path          │
├────────────────────────────────────────────────────────────-┤
│ Filter: status:open  Sort: priority ↑                      │
│                                                            │
│  ID        Pri  Type     Status       Assignee  Title      │
│ ─────────────────────────────────────────────────────────  │
│▸ bd-a1b2   P1   task     open         —         Fix auth…  │
│  bd-c3d4   P1   feature  in_progress  alice     Add retry… │
│  bd-e5f6   P2   chore    open         —         Update A…  │
│  bd-g7h8   P2   bug      open         bob       Handle t…  │
│  bd-i9j0   P2   feature  open         —         Add tele…  │
│  bd-k1l2   P0   task     open         —         Deploy h…  │
│  ...                                                       │
│                                                            │
│ 18 of 47 issues (filtered)                                 │
├────────────────────────────────────────────────────────────-┤
│ /:filter  s:sort  enter:detail  esc:clear  r:refresh       │
└────────────────────────────────────────────────────────────-┘
```

Features:
- Filter by status, priority, type, assignee, label (composable via `/` prompt)
- Sort by any column
- Enter on a row opens the Detail view for that issue
- Visual indicators for blocked vs. ready issues

Data sources: `bd list --json` with appropriate filter flags

### 3. Issue Detail (`enter` from list, or `g` + issue ID)

Full detail view for a single issue.

```
┌─ Loom ─────────────────────────────────────────────────────┐
│  Dashboard  Issues  [Detail: bd-a1b2]  Tree  Critical Path │
├────────────────────────────────────────────────────────────-┤
│                                                            │
│  bd-a1b2: Fix auth validation                              │
│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━                               │
│  Status: open        Priority: P1       Type: task         │
│  Assignee: —         Created: 2025-01-15  by: bob          │
│  Parent: bd-epic-auth                                      │
│  Labels: backend, security                                 │
│                                                            │
│  ── Description ────────────────────────────────────────   │
│  Users cannot log in after password reset. The auth        │
│  validation layer rejects valid tokens generated by the    │
│  new password flow.                                        │
│                                                            │
│  ── Dependencies (1) ──────────────── blocks this ──────   │
│  ✓ bd-124  [P1] Update auth library       in_progress      │
│                                                            │
│  ── Dependents (1) ───────────────── this blocks ───────   │
│  ○ bd-125  [P0] Deploy hotfix             open             │
│                                                            │
│  ── Comments (1) ───────────────────────────────────────   │
│  alice (2025-01-15 11:00): Root cause identified           │
│                                                            │
├────────────────────────────────────────────────────────────-┤
│ esc:back  t:tree from here  enter:jump to dep  r:refresh   │
└────────────────────────────────────────────────────────────-┘
```

Data source: `bd show <id> --json`

Navigation: pressing enter on a dependency/dependent jumps to that issue's detail view. Breadcrumb-style back navigation with `esc`.

### 4. Dependency Tree (`t`)

ASCII-rendered dependency tree, either for a specific issue or the full project graph.

```
┌─ Loom ─────────────────────────────────────────────────────┐
│  Dashboard  Issues  Detail  [Tree: bd-epic-auth]  Crit…    │
├────────────────────────────────────────────────────────────-┤
│                                                            │
│  bd-epic-auth: Auth System Overhaul [P1] (open)            │
│  ├── bd-a1b2: Fix auth validation [P1] (open)              │
│  │   └── bd-124: Update auth library [P1] (in_progress)    │
│  ├── bd-c3d4: Add retry logic [P2] (in_progress)           │
│  │   ├── bd-m3n4: Integration tests [P1] (open) ⊘          │
│  │   └── bd-q1r2: Load testing [P3] (open) ⊘               │
│  └── bd-e5f6: Update API docs [P2] (open)                  │
│                                                            │
│  Legend: ✓ closed  ● in_progress  ○ open  ⊘ blocked        │
│                                                            │
│  Tree stats: 7 issues, 3 ready, 2 blocked, 2 in progress  │
│                                                            │
├────────────────────────────────────────────────────────────-┤
│ enter:detail  /:search  a:all trees  e:expand  c:collapse  │
└────────────────────────────────────────────────────────────-┘
```

Two modes:
- **Rooted tree** (`t` from detail view, or `t` + issue ID): Shows the dependency tree from a specific issue
- **Forest view** (`t` then `a`): Shows all top-level issues (no parent, no dependents) as roots, each with their dependency sub-trees

Data sources: `bd dep tree <id>`, `bd list --json` + local graph construction for forest view

### 5. Critical Path (`c`)

The longest chain(s) of blocking dependencies from any open leaf issue to a terminal goal. This is the view that answers "what's the minimum sequence of work to reach completion?"

```
┌─ Loom ─────────────────────────────────────────────────────┐
│  Dashboard  Issues  Detail  Tree  [Critical Path]          │
├────────────────────────────────────────────────────────────-┤
│                                                            │
│  Longest blocking chains to completion:                    │
│                                                            │
│  Chain 1 (depth 4, all open) ───────────────────────────   │
│  bd-124: Update auth library [P1] ●                        │
│   └→ bd-a1b2: Fix auth validation [P1] ○                   │
│       └→ bd-k1l2: Deploy hotfix [P0] ○                     │
│           └→ bd-o5p6: Release v2.1 [P2] ○                  │
│                                                            │
│  Chain 2 (depth 3, 1 in progress) ──────────────────────   │
│  bd-c3d4: Add retry logic [P2] ●                           │
│   └→ bd-m3n4: Integration tests [P1] ○                     │
│       └→ bd-s3t4: Staging deploy [P1] ○                    │
│                                                            │
│  Chain 3 (depth 2) ─────────────────────────────────────   │
│  bd-e5f6: Update API docs [P2] ○                           │
│   └→ bd-u5v6: Public launch [P0] ○                         │
│                                                            │
│  Summary: 3 chains, max depth 4, 2 P0 goals blocked       │
│                                                            │
├────────────────────────────────────────────────────────────-┤
│ enter:detail  p:sort by priority  l:sort by length         │
└────────────────────────────────────────────────────────────-┘
```

**Algorithm:** Build a DAG from all issue dependencies where `dependency_type == "blocks"`. Find all sink nodes (issues with no outgoing "blocks" edges — these are the goals). For each sink, find the longest path from any source (issues with no incoming "blocks" edges). Sort chains by length descending, then by maximum priority in the chain.

This is computed locally from the full issue + dependency dataset, not from a single `bd` command.

Data sources: `bd list --json` + `bd show <id> --json` for all open issues (batched)

---

## Project Structure

```
loom/
├── cmd/
│   └── loom/
│       └── main.go              # CLI entry point, flag parsing
├── internal/
│   ├── datasource/
│   │   ├── bd.go                # bd CLI executor
│   │   ├── parser.go            # JSON response parsing
│   │   ├── types.go             # Issue, Dependency, Comment types
│   │   ├── cache.go             # In-memory TTL cache
│   │   └── graph.go             # DAG construction from issue data
│   ├── tui/
│   │   ├── app.go               # Root Bubble Tea model, view routing
│   │   ├── keys.go              # Key bindings
│   │   ├── styles.go            # Lip Gloss styles, color theme
│   │   ├── dashboard.go         # Dashboard view model
│   │   ├── list.go              # Issue list view model
│   │   ├── detail.go            # Issue detail view model
│   │   ├── tree.go              # Dependency tree view model
│   │   └── critical.go          # Critical path view model
│   └── graph/
│       ├── dag.go               # Directed acyclic graph operations
│       ├── critical_path.go     # Longest-path / critical chain finder
│       └── dag_test.go          # Graph algorithm tests
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## CLI Interface

```
loom                        # Launch TUI in current directory (discovers .beads/)
loom --watch                # Launch with auto-refresh enabled (5s default)
loom --watch --interval 10  # Custom refresh interval in seconds
loom --beads-dir /path      # Explicit .beads directory (same as BEADS_DIR env)
loom --version              # Print version
loom --help                 # Usage
```

Loom discovers the Beads database the same way `bd` does — by walking up from the current directory looking for `.beads/`. The `--beads-dir` flag (or `BEADS_DIR` env var) overrides this.

---

## Key Bindings (Global)

| Key | Action |
|---|---|
| `d` | Switch to Dashboard |
| `i` | Switch to Issue List |
| `t` | Switch to Dependency Tree |
| `c` | Switch to Critical Path |
| `r` | Manual refresh |
| `w` | Toggle watch mode |
| `?` | Help overlay |
| `q` / `ctrl+c` | Quit |

---

## Implementation Plan

### Phase 1: Foundation

- [ ] Project scaffolding (go mod, Makefile, CI)
- [ ] `internal/datasource` — bd CLI executor + JSON parser + typed structs
- [ ] Root Bubble Tea model with tab-based view routing
- [ ] Lip Gloss theme and shared styles
- [ ] Dashboard view (status/priority counts, ready queue)

### Phase 2: Core Views

- [ ] Issue List view with table, sorting, cursor navigation
- [ ] Issue Detail view with scrollable viewport
- [ ] Navigation flow: list → detail → back
- [ ] Filter prompt on issue list (`/` to filter)

### Phase 3: Graph Views

- [ ] `internal/graph` — DAG construction from issue dependency data
- [ ] Dependency Tree view (rooted + forest mode)
- [ ] Critical Path view (longest blocking chains)
- [ ] Jump-to-detail from tree/critical path nodes

### Phase 4: Polish

- [ ] Watch mode with configurable polling interval
- [ ] In-memory cache to avoid redundant bd calls within a refresh cycle
- [ ] Responsive layout (adapt to terminal width/height)
- [ ] Color-coded priority and status indicators
- [ ] Error handling for missing `bd` binary, uninitialized projects, etc.
- [ ] README, install instructions, release workflow

---

## Non-Goals

- **Writing to Beads.** Loom is read-only. Period.
- **Replacing bd.** Loom doesn't duplicate `bd` commands. It visualizes `bd` output.
- **Linking against Dolt.** The bd CLI is the integration boundary. No CGO dependency, no embedded database.
- **Agent features.** No MCP server, no JSON API mode for robots, no slash commands.
- **Web UI.** Terminal only. If you want a browser-based viewer, that's a different project.
- **Multi-project.** Loom operates on one `.beads/` database at a time, same as `bd`.

---

## Open Questions

1. **Batch fetching for critical path.** Computing the full dependency graph requires `bd show --json` for every open issue. For large projects (hundreds of issues), this could be slow. Worth investigating whether `bd list --json` includes enough dependency info, or whether `bd query` with a direct SQL query against the Dolt database would be faster (at the cost of coupling to the schema).

2. **Hierarchical vs. dependency relationships.** Beads has both parent-child hierarchy (`bd-a3f8.1.1`) and explicit dependency links (`blocks`, `relates_to`). The tree view needs a clear decision on whether to show hierarchy, dependencies, or both — and how to visually distinguish them.

3. **`bd graph` command.** Beads has a `bd graph` command that may already output structured graph data. Worth checking if it provides DOT or JSON output that could replace manual DAG construction.

4. **Terminal rendering of wide trees.** Deep or wide dependency trees will exceed terminal width. Need a strategy — horizontal scrolling, collapsible nodes, or a simplified "outline" mode for deep trees.
