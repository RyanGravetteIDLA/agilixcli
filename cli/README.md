# Agilix Buzz CLI

**The first maintained, agent-native CLI for Agilix Buzz — every DLAP command plus offline activity analytics, content review, and a local database no other Buzz tool has.**

Agilix Buzz's DLAP API has 260+ commands but every existing wrapper is abandoned, and its 15-minute login tokens make scripting painful. This CLI mirrors the full command surface with typed flags, --json, --dry-run, and typed exit codes; solves the token grind with auto-login and refresh; and syncs domains, users, courses, enrollments, roles, gradebook, and content into a local SQLite store so you can answer class- and domain-wide questions (idle students, teacher grading backlogs, grade distributions, content diffs, tree-wide role audits) in a single command.

## Install

The recommended path installs both the `agilix-buzz-pp-cli` binary and the `pp-agilix-buzz` agent skill (Claude Code, Codex, Cursor, Gemini CLI, GitHub Copilot, and other agents supported by the upstream [`skills`](https://github.com/vercel-labs/skills) CLI) in one shot:

```bash
npx -y @mvanhorn/printing-press-library install agilix-buzz
```

For CLI only (no skill):

```bash
npx -y @mvanhorn/printing-press-library install agilix-buzz --cli-only
```

For skill only — installs the skill into the same agents as the default command above, but skips the CLI binary (use this to update or reinstall just the skill):

```bash
npx -y @mvanhorn/printing-press-library install agilix-buzz --skill-only
```

To constrain the skill install to one or more specific agents (repeatable — agent names match the [`skills`](https://github.com/vercel-labs/skills) CLI):

```bash
npx -y @mvanhorn/printing-press-library install agilix-buzz --agent claude-code
npx -y @mvanhorn/printing-press-library install agilix-buzz --agent claude-code --agent codex
```

### Without Node (Go fallback)

If `npx` isn't available (no Node, offline), install the CLI directly via Go (requires Go 1.26.4 or newer):

```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/agilix-buzz/cmd/agilix-buzz-pp-cli@latest
```

This installs the CLI only — no skill.

### Pre-built binary

Download a pre-built binary for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/agilix-buzz-current). On macOS, clear the Gatekeeper quarantine: `xattr -d com.apple.quarantine <binary>`. On Unix, mark it executable: `chmod +x <binary>`.

<!-- pp-hermes-install-anchor -->
## Install for Hermes

Install the CLI binary first. The installer writes binaries to a per-user managed bin directory by default: `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows.

```bash
npx -y @mvanhorn/printing-press-library install agilix-buzz --cli-only
```

Then install the focused Hermes skill.

From the Hermes CLI:

```bash
hermes skills install mvanhorn/printing-press-library/cli-skills/pp-agilix-buzz --force
```

Inside a Hermes chat session:

```bash
/skills install mvanhorn/printing-press-library/cli-skills/pp-agilix-buzz --force
```

Restart the Hermes session or gateway if the newly installed skill is not visible immediately.

## Install for OpenClaw
Install both the CLI binary and the focused OpenClaw skill. The installer defaults binaries to a per-user bin directory (`$HOME/.local/bin` on macOS/Linux, `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows):

```bash
npx -y @mvanhorn/printing-press-library install agilix-buzz --agent openclaw
```

Restart the OpenClaw session or gateway if the newly installed skill is not visible immediately.

## Use with Claude Desktop

This CLI ships an [MCPB](https://github.com/modelcontextprotocol/mcpb) bundle — Claude Desktop's standard format for one-click MCP extension installs (no JSON config required).

To install:

1. Download the `.mcpb` for your platform from the [latest release](https://github.com/mvanhorn/printing-press-library/releases/tag/agilix-buzz-current).
2. Double-click the `.mcpb` file. Claude Desktop opens and walks you through the install.
3. Fill in `BUZZ_TOKEN` when Claude Desktop prompts you.

Requires Claude Desktop 1.0.0 or later. Pre-built bundles ship for macOS Apple Silicon (`darwin-arm64`) and Windows (`amd64`, `arm64`); for other platforms, use the manual config below.

<details>
<summary>Manual JSON config (advanced)</summary>

If you can't use the MCPB bundle (older Claude Desktop, unsupported platform), install the MCP binary and configure it manually.


```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/agilix-buzz/cmd/agilix-buzz-pp-mcp@latest
```

Add to your Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "agilix-buzz": {
      "command": "agilix-buzz-pp-mcp",
      "env": {
        "BUZZ_TOKEN": "<your-key>"
      }
    }
  }
}
```

</details>

## Authentication

Buzz issues short-lived (~15 minute) session tokens that auto-extend on every API call. Quickest path: export BUZZ_TOKEN with a token (mint one with 'session login --username domain/user --password ...'), or save it with 'agilix-buzz-pp-cli auth set-token <token>'. Run 'agilix-buzz-pp-cli auth status' to confirm and 'agilix-buzz-pp-cli doctor' to check connectivity. Delete commands are intentionally omitted from this build per least-privilege policy.

## Quick Start

```bash
# Health check — verifies config and reachability without needing credentials
agilix-buzz-pp-cli doctor --dry-run

# Read a domain to confirm auth works
agilix-buzz-pp-cli domains get --domainid 254591853 --json

# Mirror the domain subtree into the local SQLite store
agilix-buzz-pp-cli sync --resources domains,users,courses,enrollments,roles --global-param domainid=254591853

# Offline full-text search over synced users
agilix-buzz-pp-cli search "smith" --type users

# Macro engagement and grade distribution across the domain
agilix-buzz-pp-cli class engagement --type domain --domainid 254591853 --agent

```

## Unique Features

These capabilities aren't available in any other tool for this API.

### Activity analytics
- **`gradebook stale`** — Find teachers and courses with the largest ungraded backlogs and the oldest unscored submissions, tree-wide.

  _Reach for this to catch teacher grading backlogs before a parent complaint; it ranks by oldest unscored work, which no API endpoint exposes._

  ```bash
  agilix-buzz-pp-cli gradebook stale --type domain --domainid 254591853 --agent
  ```
- **`teacher activity`** — Per-teacher last login, days since last grade, median submit-to-grade turnaround, feedback count, and active-section count — flags inactive teachers.

  _Use it to surface teacher disengagement (no logins, slow grading) across a whole program in one call._

  ```bash
  agilix-buzz-pp-cli teacher activity --domainid 254591853 --agent --select teacherid,lastlogin,turnaround_days,ungraded
  ```
- **`students falling-behind`** — Ranks students by a composite risk score: idle days, low completion ratio, low submission rate, and grades below a threshold.

  _Reach for this mid-term to triage which students need intervention, ranked, not one record at a time._

  ```bash
  agilix-buzz-pp-cli students falling-behind --type course --entityid <courseid> --agent
  ```
- **`class engagement`** — Per-section active-vs-idle counts, submission and completion rates, A-F grade-distribution histogram, per-item completion funnel, and mean time-on-task.

  _Use it for a macro read on class health and grade distributions instead of clicking section by section._

  ```bash
  agilix-buzz-pp-cli class engagement --type domain --domainid 254591853 --select distribution,funnel --agent
  ```

### Content review
- **`content tree`** — Renders a course's full item hierarchy (folders, assignments, assessments) with type, points, due dates, and resource-count rollups.

  _Reach for this to review what a course actually contains before term start, in one view._

  ```bash
  agilix-buzz-pp-cli content tree <courseid> --agent --select title,type,points,due
  ```
- **`content diff`** — Diffs two courses' content: added/removed/changed items, points-total drift, and resource-count drift.

  _Use it to verify a derived course copy against its master, or compare a course across terms._

  ```bash
  agilix-buzz-pp-cli content diff <masterid> <copyid> --agent
  ```

### Audit & reconciliation
- **`roles audit`** — Lists every role and right a user holds across the whole domain tree and builds a role co-occurrence matrix.

  _Reach for this when restructuring roles or auditing access — it answers 'what does this person have, everywhere.'_

  ```bash
  agilix-buzz-pp-cli roles audit --type user --userid 267876066 --agent
  ```
- **`reconcile`** — Read-only diff between a provided roster file and live enrollments: who is missing, extra, or has a role mismatch.

  _Use it before provisioning to see exactly what would change, without writing anything._

  ```bash
  agilix-buzz-pp-cli reconcile roster.csv --domainid 254591853 --agent
  ```

## Recipes


### Find teacher grading backlogs

```bash
agilix-buzz-pp-cli gradebook stale --type domain --domainid 254591853 --agent --select teacherid,course,ungraded,oldest_days
```

Ranks courses/teachers by ungraded backlog and oldest unscored submission across the domain.

### Triage at-risk students mid-term

```bash
agilix-buzz-pp-cli students falling-behind --type course --entityid <courseid> --agent
```

Returns students ranked by a composite risk score from idle days, completion, submissions, and grades.

### Review a course's content before term start

```bash
agilix-buzz-pp-cli content tree <courseid> --agent --select title,type,points,due
```

Renders the full item hierarchy with points and resource rollups; pair --agent with --select to keep the payload small on large manifests.

### Verify a course copy against its master

```bash
agilix-buzz-pp-cli content diff <masterid> <copyid> --agent
```

Shows added/removed/changed items and points drift between the master and a derived copy.

### Audit everything a user can access

```bash
agilix-buzz-pp-cli roles audit --type user --userid 267876066 --agent
```

Lists every role and right the user holds across the whole domain tree.

## Usage

Run `agilix-buzz-pp-cli --help` for the full command reference and flag list.

## Paths & environment variables

This CLI separates local files into four path kinds:

| Kind | Contents |
|------|----------|
| `config` | User-editable settings such as `config.toml` and saved profiles |
| `data` | Durable local data: `credentials.toml`, `data.db`, cookies, browser-session proof files, and other auth sidecars |
| `state` | Runtime state such as persisted queries, jobs, and `teach.log` |
| `cache` | Regenerable HTTP/cache files |

Each kind resolves independently. The ladder is:

1. Per-kind env var: `AGILIX_BUZZ_CONFIG_DIR`, `AGILIX_BUZZ_DATA_DIR`, `AGILIX_BUZZ_STATE_DIR`, or `AGILIX_BUZZ_CACHE_DIR`
2. `--home <dir>` for this invocation
3. `AGILIX_BUZZ_HOME` for a flat relocated root
4. XDG env vars: `XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`
5. Platform defaults matching existing installs

For containers and agent sandboxes, prefer a single relocated root:

```bash
export AGILIX_BUZZ_HOME=/srv/agilix-buzz
agilix-buzz-pp-cli doctor
```

Under `AGILIX_BUZZ_HOME=/srv/agilix-buzz`, the four dirs resolve to `/srv/agilix-buzz/config`, `/srv/agilix-buzz/data`, `/srv/agilix-buzz/state`, and `/srv/agilix-buzz/cache`.

MCP servers do not receive CLI flags from the host. Put relocation in the host `env` block:

```json
{
  "mcpServers": {
    "agilix-buzz": {
      "command": "agilix-buzz-pp-mcp",
      "env": {
        "AGILIX_BUZZ_HOME": "/srv/agilix-buzz"
      }
    }
  }
}
```

Precedence matters in fleets: an ambient per-kind variable such as `AGILIX_BUZZ_DATA_DIR` overrides an explicit `--home` for that kind. Use `AGILIX_BUZZ_HOME` or the per-kind variables for durable fleet relocation; treat `--home` as the weaker per-invocation lever.

Relocation is one-way. Unsetting `AGILIX_BUZZ_HOME` does not move files back to platform defaults, and `doctor` cannot find credentials left under a former root. Move the files manually before unsetting relocation variables.

Existing installs keep working because the platform-default rung matches the legacy layout. On the first auth write, stored secrets leave `config.toml` and are consolidated into `credentials.toml` under the data directory. Run `agilix-buzz-pp-cli doctor --fail-on warn` to check path and credential-location warnings in automation.

## Commands

### announcements

Announcements

- **`agilix-buzz-pp-cli announcements get`** - This command gets an announcement, which is a file that conforms to the Announcement format.
- **`agilix-buzz-pp-cli announcements get-info`** - This command gets information about an announcement.
- **`agilix-buzz-pp-cli announcements get-list`** - This command lists an entity's announcements.
- **`agilix-buzz-pp-cli announcements get-user-list`** - This command lists course, section, and domain announcements for the current user.
- **`agilix-buzz-pp-cli announcements list-restorable`** - This command lists domainor course announcements that have been deleted and can be restored.
- **`agilix-buzz-pp-cli announcements put`** - This command posts an announcement to the domain or course specified by entityid .
- **`agilix-buzz-pp-cli announcements restore`** - This command restores one or more domain or course announcements.
- **`agilix-buzz-pp-cli announcements update-viewed`** - This command updates the viewed state of one or more domain announcements for the current signed-on user.

### assessments

Agilix Buzz assessments commands

- **`agilix-buzz-pp-cli assessments get-attempt`** - This command returns the data needed for an assessment or homework attempt.
- **`agilix-buzz-pp-cli assessments get-attempt-file`** - This command gets a uploaded file associated with a fileupload question.
- **`agilix-buzz-pp-cli assessments get-attempt-review`** - This command returns the data needed to review an assessment or homework attempt.
- **`agilix-buzz-pp-cli assessments get-next-question`** - Submits answers for the current question in an adpative assessment.
- **`agilix-buzz-pp-cli assessments get-question`** - This command gets one question from a course.
- **`agilix-buzz-pp-cli assessments get-question-scores`** - This command scores one or more question submissions.
- **`agilix-buzz-pp-cli assessments get-question-stats`** - This command gets system-wide statistics about a particular question.
- **`agilix-buzz-pp-cli assessments get-submission-state`** - Retrieves state information for an enrollment's assessment or homework submission, including whether the enrollment can start, resume, or re
- **`agilix-buzz-pp-cli assessments list-questions`** - This command lists one or more questions in a course.
- **`agilix-buzz-pp-cli assessments list-restorable-questions`** - This command lists questions that have been deleted and can be restored.
- **`agilix-buzz-pp-cli assessments put-attempt-file`** - This command puts files on the server to be associated with a fileupload question as part of a student assessment/homework attempt.
- **`agilix-buzz-pp-cli assessments put-questions`** - This command puts (adds or updates) one or more questions in a course.
- **`agilix-buzz-pp-cli assessments restore-questions`** - This command restores one or more questions in a course.
- **`agilix-buzz-pp-cli assessments save-attempt-answers`** - Save answers for an assessment of homework group attempt for later use.
- **`agilix-buzz-pp-cli assessments submit-attempt-answers`** - Submits answers for an assessment of homework group attempt for grading.

### badges

Badges

- **`agilix-buzz-pp-cli badges create`** - This command creates a badge.
- **`agilix-buzz-pp-cli badges get`** - This command gets the badge image.
- **`agilix-buzz-pp-cli badges get-assertion`** - This command gets the assertion associated with a badge.
- **`agilix-buzz-pp-cli badges get-list`** - This command gets a list of badges for a user.

### blogs

Blogs

- **`agilix-buzz-pp-cli blogs get`** - This command gets a blog or journal message, which conforms to the Message format.
- **`agilix-buzz-pp-cli blogs get-list`** - This command returns a list of blog or journal messages for the specified enrollment and item, ordered by message creationdate .
- **`agilix-buzz-pp-cli blogs get-recent-posts`** - This command returns recent posts to discussion boards, wikis, blogs and journals that have not been marked as viewed.
- **`agilix-buzz-pp-cli blogs get-summary`** - This command returns a blog or journal summary for each enrollment in the specified entity.
- **`agilix-buzz-pp-cli blogs put`** - This command puts a blog or journal message to the server.
- **`agilix-buzz-pp-cli blogs update-viewed`** - This command updates a user’s viewed state of one or more blog messages.

### conversion

Format conversion

- **`agilix-buzz-pp-cli conversion export-data`** - This command convert structured post data to a tab or comma-delimited file.
- **`agilix-buzz-pp-cli conversion get-converted-data`** - Retrieves the converted data generated from either ImportData or ExportData and deletes the temporary file.
- **`agilix-buzz-pp-cli conversion import-data`** - This command imports and converts various data formats.

### courses

Courses & course copy

- **`agilix-buzz-pp-cli courses copy`** - This command copies one or more courses.
- **`agilix-buzz-pp-cli courses create`** - This command creates one or more courses.
- **`agilix-buzz-pp-cli courses create-demo`** - This command creates a demo course.
- **`agilix-buzz-pp-cli courses deactivate`** - This command deactivates a course.
- **`agilix-buzz-pp-cli courses get`** - This command gets information for a course.
- **`agilix-buzz-pp-cli courses list`** - This command lists courses.
- **`agilix-buzz-pp-cli courses merge`** - This command merges the deltas from a derivative course into its immediate base course.
- **`agilix-buzz-pp-cli courses restore`** - This command restores a deleted course.
- **`agilix-buzz-pp-cli courses update`** - This command updates the title, reference, data (free-form XML), and other attributes of one or more courses.

### datastreams

Data streams

- **`agilix-buzz-pp-cli datastreams get-data-stream-configuration`** - This command gets the data stream configuration previously set using SetDataStreamConfiguration for a specified domain.
- **`agilix-buzz-pp-cli datastreams set-data-stream-configuration`** - This command tests a given data stream configuration to see if it is valid and if the API servers can connect to it and if successful, sets 

### discussions

Discussion boards

- **`agilix-buzz-pp-cli discussions get-message`** - This command returns a discussion forum message.
- **`agilix-buzz-pp-cli discussions get-message-list`** - This command returns the messages associated with a discussion board of an entity.
- **`agilix-buzz-pp-cli discussions list-restorable-messages`** - This command lists messages that have been deleted from a entity and can be restored.
- **`agilix-buzz-pp-cli discussions put-message`** - This command puts a discussion board message to the server.
- **`agilix-buzz-pp-cli discussions put-message-part`** - This command puts individual parts of a discussion board message to the server.
- **`agilix-buzz-pp-cli discussions restore-messages`** - This command restores one or more messages in a discussion forum.
- **`agilix-buzz-pp-cli discussions submit-message`** - This command assembles changes from the PutMessagePart and DeleteMessagePart commands into a new message and posts the new message to the se
- **`agilix-buzz-pp-cli discussions update-message-viewed`** - This command updates a user’s viewed state of one or more discussion board messages.

### domains

Domains (schools/districts) & hierarchy

- **`agilix-buzz-pp-cli domains create`** - This command creates one or more domains and links them as children to the domain specified by parentid.
- **`agilix-buzz-pp-cli domains get`** - This command gets information for the specified domain.
- **`agilix-buzz-pp-cli domains get-content`** - This command returns the list of content items for the current signed-on user’s domain.
- **`agilix-buzz-pp-cli domains get-enrollment-metrics`** - This command gets enrollment metrics for courses in the specified domain.
- **`agilix-buzz-pp-cli domains get-enrollment-metrics-parameters`** - This command gets parameters that affect calculations used in Enrollment Metrics , Course EnrollmentMetrics , and GetDomainEnrollmentMetrics
- **`agilix-buzz-pp-cli domains get-parent-list`** - This command gets the list of parent domains for a domain.
- **`agilix-buzz-pp-cli domains get-stats`** - This command gets statistics for a domain.
- **`agilix-buzz-pp-cli domains list`** - This command lists domains.
- **`agilix-buzz-pp-cli domains restore`** - This command restores a deleted domain.
- **`agilix-buzz-pp-cli domains update`** - This command updates the name, reference value, flags, and free-forn structured data of one or more domains.

### enrollments

Enrollments (user-course bindings) & activity

- **`agilix-buzz-pp-cli enrollments create`** - This command enrolls a user with the specified rights in a course.
- **`agilix-buzz-pp-cli enrollments get`** - This command gets enrollment data for a particular enrollment.
- **`agilix-buzz-pp-cli enrollments get-activity`** - This command gets the activity detail for the specified user enrollment.
- **`agilix-buzz-pp-cli enrollments get-gradebook`** - This command gets the gradebook detail, including rolled-up period, category, and course grades, for the specified user enrollment.
- **`agilix-buzz-pp-cli enrollments get-group-list`** - This command lists groups that the specified enrollment is a member of.
- **`agilix-buzz-pp-cli enrollments get-metrics-report`** - Returns report data for the enrollment metrics of a domain, teacher, or student.
- **`agilix-buzz-pp-cli enrollments list`** - This command lists enrollments.
- **`agilix-buzz-pp-cli enrollments list-by-teacher`** - This command lists enrollments in courses where the specified user is a teacher.
- **`agilix-buzz-pp-cli enrollments list-entity`** - This command lists enrollments in the specified entity.
- **`agilix-buzz-pp-cli enrollments list-user`** - This command lists enrollments for the specified user.
- **`agilix-buzz-pp-cli enrollments put-self-assessment`** - This command puts self-assessment ratings of understanding, interest, and effort into the student's Enrollment Metrics .
- **`agilix-buzz-pp-cli enrollments restore`** - This command restores a deleted enrollment.
- **`agilix-buzz-pp-cli enrollments update`** - This command updates the enrollment status, start date, end date, and rights granted to the specified user for a course.

### files

Course resources & files

- **`agilix-buzz-pp-cli files copy`** - This command copies one or more resources from the domain or course specified by sourceentityid to the domain or course specified by destina
- **`agilix-buzz-pp-cli files get`** - This command gets resource metadata and/or binary content for the course, section, enrollment, user, or domain.
- **`agilix-buzz-pp-cli files get-document`** - This command retrieves a user document from the server.
- **`agilix-buzz-pp-cli files get-document-info`** - This command retrieves information about one or more user documents from the server.
- **`agilix-buzz-pp-cli files get-entity-id`** - This command gets the resource entity ID of the specified entity.
- **`agilix-buzz-pp-cli files get-info`** - This command gets information about a resource in the domain, course, section, or enrollment specified by entityid .
- **`agilix-buzz-pp-cli files get-list`** - This command lists metadata for entity (domain, course, or enrollment) resources.
- **`agilix-buzz-pp-cli files list-restorable`** - This command lists resources that have been deleted and can be restored.
- **`agilix-buzz-pp-cli files list-restorable-documents`** - This command lists documents that have been deleted and can be restored.
- **`agilix-buzz-pp-cli files put`** - This command puts a resource to the server for the specified course, section, enrollment, user, or domain.
- **`agilix-buzz-pp-cli files put-folders`** - This command creates one or more resource folders for a domain or course.
- **`agilix-buzz-pp-cli files restore`** - This command restores one or more resources in a domain, course, or enrollment.
- **`agilix-buzz-pp-cli files restore-documents`** - This command restores one or more documents.

### gradebook

Gradebook results & grading

- **`agilix-buzz-pp-cli gradebook calculate-enrollment-scenario`** - This command calculates a rolled-up (category, period, and course) grade scenario for the specified user enrollment given the supplied input
- **`agilix-buzz-pp-cli gradebook get-calendar`** - Gets the iCalendar (RFC2445) file of events for the user or enrollments specified by token.
- **`agilix-buzz-pp-cli gradebook get-calendar-items`** - This command gets duedates and blackoutdates for the specified enrollments.
- **`agilix-buzz-pp-cli gradebook get-calendar-token`** - Get a token to generate a URL that can retrieve the current user's calendar without any authentication.
- **`agilix-buzz-pp-cli gradebook get-certificates`** - This command gets the completion certificates associated with a course or section enrollment.
- **`agilix-buzz-pp-cli gradebook get-due-soon-list`** - This command gets list of items that will soon become due for a student.
- **`agilix-buzz-pp-cli gradebook get-entity-summary`** - This command gets grade summaries for students enrolled in the specified entity.
- **`agilix-buzz-pp-cli gradebook get-entity-work`** - Gets the list of items that have been graded or need grading in a course back to a specified date.
- **`agilix-buzz-pp-cli gradebook get-grade`** - This command gets the grade detail for the specified user enrollment and item.
- **`agilix-buzz-pp-cli gradebook get-grade-history`** - This command gets the history of grades for an item in the gradebook for the specified user enrollment.
- **`agilix-buzz-pp-cli gradebook get-item-analysis`** - Some items compute their score from scores assigned within the item.
- **`agilix-buzz-pp-cli gradebook get-item-report`** - Some items compute their score from scores assigned within the item.
- **`agilix-buzz-pp-cli gradebook get-list`** - This command lists entities (courses and sections) that have gradebooks.
- **`agilix-buzz-pp-cli gradebook get-rubric-mastery`** - This command gets report data that shows how students are performing for items related to the specified rubric.
- **`agilix-buzz-pp-cli gradebook get-rubric-stats`** - This command gets detailed information about responses and scores associated with the rules in a rubric.
- **`agilix-buzz-pp-cli gradebook get-summary`** - This command gets a summary of course participants that match the specified parameters.
- **`agilix-buzz-pp-cli gradebook get-user`** - This command gets gradebook detail for the specified user.
- **`agilix-buzz-pp-cli gradebook get-weights`** - This command gets the item and category gradebook weights for the specified entity, optionally filtered by a specific grading period.
- **`agilix-buzz-pp-cli gradebook getentitygradebook2`** - This command gets grades for all students enrolled in the specified entity.
- **`agilix-buzz-pp-cli gradebook getentitygradebook3`** - This command gets grades for students enrolled in the specified entity.

### groups

Groups

- **`agilix-buzz-pp-cli groups add-members`** - This command adds one or more member enrollments to an existing group.
- **`agilix-buzz-pp-cli groups create`** - This command creates one or more groups in the specified owner course.
- **`agilix-buzz-pp-cli groups get`** - This command gets information for a particular group.
- **`agilix-buzz-pp-cli groups get-enrollment-list`** - This command lists enrollments for the specified group.
- **`agilix-buzz-pp-cli groups get-list`** - This command lists groups in the specified owner entity (course).
- **`agilix-buzz-pp-cli groups remove-members`** - This command removes one or more member enrollments from an existing group.
- **`agilix-buzz-pp-cli groups update`** - This command updates the title, set ID, reference, and data of one or more groups.

### items

Manifests & content items

- **`agilix-buzz-pp-cli items assign`** - Assigning an item to a folder changes the item’s parent and sequence to the specified values.
- **`agilix-buzz-pp-cli items copy`** - This command copies one or more items from one course to another.
- **`agilix-buzz-pp-cli items find-personalized-entities`** - This command finds the course, enrollment, and group entities that contain items which have been personalized in the way specified by the qu
- **`agilix-buzz-pp-cli items get-data`** - This command gets the manifest data (see Course Data ) for the specified entity (course, section, group, or enrollment) except for the <item
- **`agilix-buzz-pp-cli items get-links`** - This command gets a list of manifest items that link to the specified item through course or item chaining.
- **`agilix-buzz-pp-cli items get-list`** - This command lists items.
- **`agilix-buzz-pp-cli items getitem`** - This command gets a manifest item.
- **`agilix-buzz-pp-cli items getiteminfo`** - This command gets information about an item in the course, section or group specified by entityid .
- **`agilix-buzz-pp-cli items getmanifest`** - This command gets the manifest for the specified entity (course, section, group, or enrollment).
- **`agilix-buzz-pp-cli items getmanifestinfo`** - This command gets information about the manifest of one or more courses or sections.
- **`agilix-buzz-pp-cli items getmanifestitem`** - This command gets an item and optionally some of its descendents from the manifest of a course or section.
- **`agilix-buzz-pp-cli items list-assignable`** - ListAssignableItems uses the assignableitemsquery attribute on the folderid item to list items in the course that may be assigned to folderi
- **`agilix-buzz-pp-cli items list-restorable`** - This command lists items that have been deleted and can be restored.
- **`agilix-buzz-pp-cli items navigate`** - This command attempts to navigate a user's enrollment to a course item.
- **`agilix-buzz-pp-cli items put`** - This command puts one or more items in a manifest.
- **`agilix-buzz-pp-cli items restore`** - This command restores one or more items in a manifest.
- **`agilix-buzz-pp-cli items search`** - This command searches course content that has been marked by course creators as searchable and returns a paginated list of results.
- **`agilix-buzz-pp-cli items unassign`** - Unassigning an item changes the item’s parent and sequence to their default values (the values they would have if they had never been assign
- **`agilix-buzz-pp-cli items update-data`** - This command updates the manifest data (see Course Data ) on the specified entity (course or section.) Only the elements specified in the PO

### objectives

Learning objectives & progress

- **`agilix-buzz-pp-cli objectives create-sets`** - This command creates one or more objective sets or objective map sets, which are containers for either objectives or objective maps, respect
- **`agilix-buzz-pp-cli objectives get-list`** - This command gets a list of learning objectives.
- **`agilix-buzz-pp-cli objectives get-map-list`** - This command gets the list of maps defined in an objective map set.
- **`agilix-buzz-pp-cli objectives get-mastery`** - This command gets objective mastery report data for the specified entity using student scores from objective-aligned items and questions.
- **`agilix-buzz-pp-cli objectives get-mastery-detail`** - This command summarizes a single learning objective's mastery for each enrollment in the specified entities.
- **`agilix-buzz-pp-cli objectives get-mastery-summary`** - This command lists learning objectives and summarizes their mastery for enrollments in the specified entities.
- **`agilix-buzz-pp-cli objectives get-set`** - This command gets information for a objective set or objective map set.
- **`agilix-buzz-pp-cli objectives get-subject-list`** - This command gets the list of subjects covered by the learning objectives for an objective set.
- **`agilix-buzz-pp-cli objectives list-sets`** - This command gets a list of objective sets or objective map sets.
- **`agilix-buzz-pp-cli objectives put`** - This creates or updates one or more learning objectives and puts them in an objective set.
- **`agilix-buzz-pp-cli objectives put-maps`** - This command creates or updates one or more objective maps and puts them into an objective map set.
- **`agilix-buzz-pp-cli objectives restore-set`** - This command restores a deleted objective set.
- **`agilix-buzz-pp-cli objectives update-sets`** - This command updates the name, reference, and other attributes of one or more objective sets or objective map sets.

### peergrading

Peer grading

- **`agilix-buzz-pp-cli peergrading get-response`** - This command gets a peer's response to a student's submission.
- **`agilix-buzz-pp-cli peergrading get-response-info`** - This command retrieves information about peer responses from the server.
- **`agilix-buzz-pp-cli peergrading get-response-list`** - This command lists peer's responses to the specified student's submission to a course item.
- **`agilix-buzz-pp-cli peergrading get-review-list`** - This command lists peers that may be reviewed by the specified enrollment user.
- **`agilix-buzz-pp-cli peergrading put-response`** - This command puts peer response data including comments, rubric score, and likert responses to the server.

### ratings

Ratings

- **`agilix-buzz-pp-cli ratings get-item`** - This command gets a user rating on a manifest item.
- **`agilix-buzz-pp-cli ratings get-item-summary`** - This command gets the summary of user ratings on a manifest item.
- **`agilix-buzz-pp-cli ratings put-item`** - This command puts a user rating on a manifest item.

### reports

Reports

- **`agilix-buzz-pp-cli reports get-info`** - This command lists the information about a report including the parameters required to run it.
- **`agilix-buzz-pp-cli reports get-list`** - This command lists all reports defined on a domain.
- **`agilix-buzz-pp-cli reports get-runnable-list`** - This command lists all reports available to the currently logged-on user.
- **`agilix-buzz-pp-cli reports run`** - This command runs the specified report.

### rights

Rights & role assignments

- **`agilix-buzz-pp-cli rights create-role`** - This command creates a role on a given domain.
- **`agilix-buzz-pp-cli rights get`** - This command gets the rights granted to the specified actor (user or role) for the specified entity (domain, course, or section).
- **`agilix-buzz-pp-cli rights get-actor`** - This command lists entities that an actor (user) has rights for.
- **`agilix-buzz-pp-cli rights get-effective`** - This command gets the effective rights granted to the current user for the specified entity (domain, course, or section).
- **`agilix-buzz-pp-cli rights get-effective-subscription-list`** - This command lists effective subscriptions for the current signed-in user, including those inherited from the user's domain, and excluding a
- **`agilix-buzz-pp-cli rights get-entity`** - This command lists users and the rights granted to those users for the specified entity (domain, course, section, enrollment or user.)
- **`agilix-buzz-pp-cli rights get-entity-subscription-list`** - This command lists subscriptions to the specified entity (course or domain).
- **`agilix-buzz-pp-cli rights get-list`** - This command lists rights granted to users on entities (domains, users, and enrollments).
- **`agilix-buzz-pp-cli rights get-personas`** - This command gets the list of personas associated with a user.
- **`agilix-buzz-pp-cli rights get-role`** - This command gets information for a role.
- **`agilix-buzz-pp-cli rights get-subscription-list`** - This command lists subscriptions explicitly assigned to the specified subscriber (user or domain).
- **`agilix-buzz-pp-cli rights list-roles`** - This command lists roles defined for a domain.
- **`agilix-buzz-pp-cli rights restore-role`** - This command restores a role that has been previously deleted.
- **`agilix-buzz-pp-cli rights update`** - This command updates the rights (flags) granted to the specified user (actorid) for the domain or enrollment specified by entityid.
- **`agilix-buzz-pp-cli rights update-role`** - This command updates the name, domainid, reference, or privileges for a role.
- **`agilix-buzz-pp-cli rights update-subscriptions`** - This command creates or updates the subscription for the specified subscriber.

### session

Authentication & session (login, token, proxy/login-as)

- **`agilix-buzz-pp-cli session clear-second-factor-authentication`** - Clears 2FA (second factor authentication) settings for the specified user account, allowing the user to login without 2FA (the user will hav
- **`agilix-buzz-pp-cli session create-second-factor-authentication-secret`** - This command creates an RFC 6238 second factor authentication secret for use with 2FA authentication.
- **`agilix-buzz-pp-cli session extend-session`** - Each API command automatically refreshes the authorization token, extending the expiration for the duration originally specified at login, b
- **`agilix-buzz-pp-cli session finish-password-reset`** - This command updates the password of the specified user.
- **`agilix-buzz-pp-cli session force-password-change`** - Forces a user to change their password the next time they login, exactly as if their password had expired.
- **`agilix-buzz-pp-cli session get-key`** - This command retrieves a name/value pair from an entity.
- **`agilix-buzz-pp-cli session get-password-login-attempt-history`** - Gets the record of password login attempts for the specified user account.
- **`agilix-buzz-pp-cli session get-password-question`** - This command gets the question to ask a user who has forgotten their password.
- **`agilix-buzz-pp-cli session login`** - For server-side API integrations, OAuth 2.0 Application Identity is now available as an alternative authentication mechanism.
- **`agilix-buzz-pp-cli session logout`** - This command terminates the session state for an authenticated user.
- **`agilix-buzz-pp-cli session proxy`** - This command starts a session as a different user.
- **`agilix-buzz-pp-cli session put-key`** - This command stores a name/value pair on an entity.
- **`agilix-buzz-pp-cli session reset-lockout`** - This command resets an account that has been locked out due to too many contiguous password failures.
- **`agilix-buzz-pp-cli session reset-password`** - This command sends an email to the specified user with a time-limited link that can be used to reset their password.
- **`agilix-buzz-pp-cli session second-factor-authenticate`** - This command completes the two factor part of authentication for users who have opted for two factor authentication.
- **`agilix-buzz-pp-cli session setup-second-factor-authentication`** - Turns on second factor authentication for an account.
- **`agilix-buzz-pp-cli session unproxy`** - This command ends a proxy session.
- **`agilix-buzz-pp-cli session update-password`** - This command updates the password of the specified user.
- **`agilix-buzz-pp-cli session update-password-question-answer`** - This command updates the password question and answer for the specified user.

### submissions

Student submissions & responses

- **`agilix-buzz-pp-cli submissions get-sco-data`** - This command gets a user's SCORM data for a SCO activity from the server.
- **`agilix-buzz-pp-cli submissions get-student`** - This command gets a student's submission.
- **`agilix-buzz-pp-cli submissions get-student-history`** - This command gets the history of submissions for an item in the gradebook for the specified user enrollment.
- **`agilix-buzz-pp-cli submissions get-student-info`** - This command retrieves information about one or more student submissions from the server.
- **`agilix-buzz-pp-cli submissions get-teacher-response`** - This command gets a teacher's response to a student's submission.
- **`agilix-buzz-pp-cli submissions get-teacher-response-info`** - This command retrieves information about teacher responses from the server.
- **`agilix-buzz-pp-cli submissions get-work-in-progress`** - This command gets a work-in-progress file.
- **`agilix-buzz-pp-cli submissions put-item-activity`** - This command reports per-item student time spent to the server.
- **`agilix-buzz-pp-cli submissions put-sco-data`** - This command puts a user's SCORM data for a SCO activity to the server.
- **`agilix-buzz-pp-cli submissions put-student`** - This command puts a student submission for an activity to the server.
- **`agilix-buzz-pp-cli submissions put-teacher-response`** - This command puts teacher response data including scores, comments, and grade status flags to the server.
- **`agilix-buzz-pp-cli submissions put-teacher-responses`** - This command puts one or more teacher responses including scores and grade status flags to the server.
- **`agilix-buzz-pp-cli submissions put-work-in-progress`** - This command puts a work-in-progress file on the server in preparation for a student submission.
- **`agilix-buzz-pp-cli submissions submit-work-in-progress`** - This command submits work-in-progress files on the server as a completed submission for the student.

### system

Server status & general utilities

- **`agilix-buzz-pp-cli system get-command-list`** - This returns a complete list of all commands available on the API Server.
- **`agilix-buzz-pp-cli system get-entity-type`** - This command gets the entity type (Course, Domain, Group, or Section) for a given entity ID.
- **`agilix-buzz-pp-cli system get-status`** - The API servers do fairly extensive self-testing for all subsystems.
- **`agilix-buzz-pp-cli system send-mail`** - This command sends an e-mail to recipients that are enrolled in the same entity as that specified by the sender's enrollmentid .

### tokens

Command tokens (delegated links)

- **`agilix-buzz-pp-cli tokens create-command`** - This command creates one or more command tokens.
- **`agilix-buzz-pp-cli tokens get-command`** - This command gets details about a previously created command token.
- **`agilix-buzz-pp-cli tokens get-command-info`** - This command gets the description and other non-sensitive info about a previously created command token from the code.
- **`agilix-buzz-pp-cli tokens list-command`** - This command lists all the command tokens created by a particular user and associated with a specified domain, course, group, or user.
- **`agilix-buzz-pp-cli tokens redeem-command`** - This command redeems a command token using the code given to the currently authenticated user by the creator.
- **`agilix-buzz-pp-cli tokens update-command`** - This command updates one or more existing command tokens, regenerating codes if the desired length or per-user settings change.

### users

User accounts

- **`agilix-buzz-pp-cli users create`** - This command creates one or more users.
- **`agilix-buzz-pp-cli users get`** - This command gets information for a particular user.
- **`agilix-buzz-pp-cli users get-active-count`** - Gets the number of active users either currently or for the given date range.
- **`agilix-buzz-pp-cli users get-activity`** - This command lists the login activity for a user.
- **`agilix-buzz-pp-cli users get-activity-stream`** - This command lists activities in a user's activity stream.
- **`agilix-buzz-pp-cli users get-domain-activity`** - This command lists activity for all users in a specified domain.
- **`agilix-buzz-pp-cli users get-profile-picture`** - This command gets binary content for the user's profile picture if it exists.
- **`agilix-buzz-pp-cli users list`** - This command lists users.
- **`agilix-buzz-pp-cli users restore`** - This command restores a deleted user.
- **`agilix-buzz-pp-cli users update`** - This command updates the first name, last name, reference field, e-mail address, and flags for one or more users.
- **`agilix-buzz-pp-cli users verify-email`** - This command verifies a messaging account associated with a user.

### wikis

Wikis

- **`agilix-buzz-pp-cli wikis copy-pages`** - This command copies one or more wiki pages from the specified source course, item and group to the specified destination course, item and gr
- **`agilix-buzz-pp-cli wikis get-page-list`** - This command lists metadata for wiki pages.
- **`agilix-buzz-pp-cli wikis list-restorable-pages`** - This command lists wiki pages that have been deleted from a course or section and can be restored.
- **`agilix-buzz-pp-cli wikis put-page`** - This command puts a wiki page to the server.
- **`agilix-buzz-pp-cli wikis restore-pages`** - This command restores one or more wiki pages in a course or section.
- **`agilix-buzz-pp-cli wikis update-page-viewed`** - This command updates a user’s viewed state of one or more wiki pages.


## Output Formats

```bash
# Human-readable table (default in terminal, JSON when piped)
agilix-buzz-pp-cli announcements get

# JSON for scripting and agents
agilix-buzz-pp-cli announcements get --json

# Filter to specific fields
agilix-buzz-pp-cli announcements get --json --select id,name,status

# Dry run — show the request without sending
agilix-buzz-pp-cli announcements get --dry-run

# Agent mode — JSON + compact + no prompts in one flag
agilix-buzz-pp-cli announcements get --agent
```

## Agent Usage

This CLI is designed for AI agent consumption:

- **Non-interactive** - never prompts, every input is a flag
- **Pipeable** - `--json` output to stdout, errors to stderr
- **Filterable** - `--select id,name` returns only fields you need
- **Previewable** - `--dry-run` shows the request without sending
- **Explicit retries** - add `--idempotent` to create retries when a no-op success is acceptable
- **Confirmable** - `--yes` for explicit confirmation of destructive actions
- **Piped input** - write commands can accept structured input when their help lists `--stdin`
- **Offline-friendly** - sync/search commands can use the local SQLite store when available
- **Agent-safe by default** - no colors or formatting unless `--human-friendly` is set

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
agilix-buzz-pp-cli doctor
```

Verifies configuration, credentials, and connectivity to the API.

## Configuration

Run `agilix-buzz-pp-cli doctor` to see the resolved config, data, state, and cache directories. The platform-default config path is `~/.config/agilix-buzz/config.toml`; `--home`, `AGILIX_BUZZ_HOME`, and per-kind env vars can relocate it.

Static request headers can be configured under `headers`; per-command header overrides take precedence.

Environment variables:

| Name | Kind | Required | Description |
| --- | --- | --- | --- |
| `BUZZ_TOKEN` | per_call | Yes | Set to your API credential. |

### agentcookie (optional)

If you use agentcookie to sync secrets across machines, this CLI auto-adopts agentcookie-managed credentials with no extra setup. When the daemon writes to this CLI's config, `agilix-buzz-pp-cli doctor` reports `agentcookie: detected` and `auth-status` labels the source as `agentcookie`. Skip this section if you don't use agentcookie - the CLI works the same as any other.

## Troubleshooting
**Authentication errors (exit code 4)**
- Run `agilix-buzz-pp-cli doctor` to check credentials
- Verify the environment variable is set: `echo $BUZZ_TOKEN`
**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

### API-specific
- **code:Unauthorized or token expired** — Mint a fresh token (it auto-extends on each call): set BUZZ_TOKEN or run 'agilix-buzz-pp-cli auth set-token <token>'; 'session login' performs a username/password Login3.
- **code:BadRequest 'requires a domainid parameter'** — Most list/report commands need --domainid; add --includedescendantdomains to walk the subtree.
- **HTTP 429 / ServerOverwhelmed** — The client honors Retry-After with exponential backoff automatically; lower --limit or sync off-peak for very large trees.

## API Throttling

The client implements Agilix Buzz's documented throttling model (see
[API Time Limiting](https://api.agilixbuzz.com/docs/entry/Concept/ApiTimeLimiting.md),
[API Rate Limiting](https://api.agilixbuzz.com/docs/entry/Concept/ApiRateLimiting.md), and
[Server Busy Limiting](https://api.agilixbuzz.com/docs/entry/Concept/ServerBusyLimiting.md)):

- **429 Too Many Requests** — honors the `Retry-After` header and adapts the request rate (covers
  both time-limiting and the rate-limiting that applies to commands like `GetUserGradebook2`).
- **503 Server Busy** — honors `Retry-After` when present, otherwise exponential backoff.
- **Proactive pacing** — reads the `X-Provisioned-Ms-Remaining` budget header and pauses briefly as
  the budget runs low, so large tree-wide analytics runs stay under the limit instead of cascading
  into 429s. This is automatic; no flags required.

## API Reference

This CLI mirrors the DLAP command surface. For full per-command details, Agilix publishes
AI-optimized docs at `https://api.agilixbuzz.com/llms.txt` (and `/llms.md`), plus per-command pages
at `https://api.agilixbuzz.com/docs/entry/Command/<CommandName>.md`.

## Sources & Inspiration

This CLI was built by studying these projects and resources:

- [**beneggett/agilix**](https://github.com/beneggett/agilix) — Ruby (3 stars)
- [**StrongMind/agilix-buzz-client**](https://github.com/StrongMind/agilix-buzz-client) — Ruby (2 stars)
- [**AgilixLabs/BuzzApiSample-CSharp**](https://github.com/AgilixLabs/BuzzApiSample-CSharp) — C#
- [**rockymadden/brainhoney.js**](https://github.com/rockymadden/brainhoney.js) — CoffeeScript

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
