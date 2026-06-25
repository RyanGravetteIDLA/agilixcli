# Agilix Buzz — Course Source & Plans

Working repository for Agilix Buzz (DLAP) course-authoring work, built with the
`agilix-buzz-pp-cli`. Course content is authored here as source files and pushed
into a live Buzz course via the CLI, so the build is reproducible and reviewable.

## Contents

```
cli/               agilix-buzz-pp-cli — the Go CLI + MCP server (full DLAP surface)
  cmd/agilix-buzz-pp-cli   CLI entrypoint
  cmd/agilix-buzz-pp-mcp   MCP server (stdio | http) for agents
  internal/                client, commands, analytics, store, mcp orchestration
docs/plans/        Planning documents (ce-plan output)
course/
  lessons/         Standalone lesson HTML (uploaded as course resources)
  questions/       Assessment question banks (schema-2 PutQuestions JSON)
  structure/       Manifest batches: module/lesson skeleton + assessment items
  BUILD.md         How to rebuild the course from these files
```

## The CLI & MCP server

`cli/` holds **agilix-buzz-pp-cli**, an agent-native CLI covering the full Agilix
Buzz DLAP command surface (238 commands) plus offline activity analytics, content
review, and a local cache. It ships an **MCP server** (`cmd/agilix-buzz-pp-mcp`,
stdio or HTTP transport) that exposes the same surface to AI agents via
code-orchestration. Build with `cd cli && make build` (or `go build ./...`);
binaries are git-ignored. Read/create/update only — delete commands are gated.

## Current course: Google Workspace Account Recovery (for Technology Directors)

A researched, assessment-driven course aimed at K-12 technology directors who
administer Google Workspace. **5 modules · 17 lessons · 5 knowledge checks ·
1 graded scenario final.** See `docs/plans/2026-06-24-gws-account-recovery-course-plan.md`
for the design, sources, and scope.

- Lessons are grounded in current (2026) official Google documentation, with
  per-lesson citations and "verify in your console" callouts (Google's admin UI
  and defaults drift).
- Knowledge checks are formative (ungraded practice); the final is graded
  (weighted, 70% pass, scenario-based, includes multi-select).

## Conventions

- **No secrets in source.** API tokens live only in ephemeral shell commands,
  never in any file here.
- **No deletes.** Per project policy, the CLI uses read/create/update only;
  delete commands are gated behind explicit approval.

See `course/BUILD.md` to (re)build the course from these source files.
