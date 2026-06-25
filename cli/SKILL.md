---
name: pp-agilix-buzz
description: "The first maintained, agent-native CLI for Agilix Buzz — every DLAP command plus offline activity analytics, content review, and a local database no other Buzz tool has. Trigger phrases: `check teacher grading activity in buzz`, `find students falling behind`, `review course content in agilix`, `audit buzz roles`, `list enrollments in domain`, `use agilix-buzz`, `run agilix buzz`."
author: "RyanGravetteIDLA"
license: "Apache-2.0"
argument-hint: "<command> [args] | install cli|mcp"
allowed-tools: "Read Bash"
metadata:
  openclaw:
    requires:
      bins:
        - agilix-buzz-pp-cli
    install:
      - kind: go
        bins: [agilix-buzz-pp-cli]
        module: github.com/mvanhorn/printing-press-library/library/productivity/agilix-buzz/cmd/agilix-buzz-pp-cli
---

# Agilix Buzz — Printing Press CLI

## Prerequisites: Install the CLI

This skill drives the `agilix-buzz-pp-cli` binary. **You must verify the CLI is installed before invoking any command from this skill.** If it is missing, install it first:

1. Install via the Printing Press installer. It defaults binaries to `$HOME/.local/bin` on macOS/Linux and `%LOCALAPPDATA%\Programs\PrintingPress\bin` on Windows:
   ```bash
   npx -y @mvanhorn/printing-press-library install agilix-buzz --cli-only
   ```
2. Verify: `agilix-buzz-pp-cli --version`
3. Ensure the reported install directory is on `$PATH` for the agent/runtime that will invoke this skill.

If the `npx` install fails (no Node, offline, etc.), fall back to a direct Go install (requires Go 1.26.4 or newer). This installs into `$GOPATH/bin` (default `$HOME/go/bin`), so add that directory to `$PATH` instead:

```bash
go install github.com/mvanhorn/printing-press-library/library/productivity/agilix-buzz/cmd/agilix-buzz-pp-cli@latest
```

If `--version` reports "command not found" after install, the runtime cannot see the binary directory on `$PATH`. Do not proceed with skill commands until verification succeeds.

Agilix Buzz's DLAP API has 260+ commands but every existing wrapper is abandoned, and its 15-minute login tokens make scripting painful. This CLI mirrors the full command surface with typed flags, --json, --dry-run, and typed exit codes; solves the token grind with auto-login and refresh; and syncs domains, users, courses, enrollments, roles, gradebook, and content into a local SQLite store so you can answer class- and domain-wide questions (idle students, teacher grading backlogs, grade distributions, content diffs, tree-wide role audits) in a single command.

## When to Use This CLI

Use this CLI for any Agilix Buzz administration or analysis task an agent or admin would script: provisioning and updating users, courses, and enrollments; copying courses; reading gradebook, submissions, and content; and especially macro analytics across a class or domain (idle/at-risk students, teacher grading activity, grade distributions, content review and diffs, tree-wide role audits). It is strongest when a question spans many domains, courses, or users at once.

## Anti-triggers

Do not use this CLI for:
- This build ships no delete or purge commands (omitted per least-privilege policy); to remove users, courses, enrollments, or domains, use the Buzz web UI or request a gated delete build.
- Do not use it as a student-facing gradebook UI; it is an admin/analyst and agent tool.
- Do not use it for real-time proctoring or live classroom interaction.

## Unique Capabilities

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

## Command Reference

**announcements** — Announcements

- `agilix-buzz-pp-cli announcements get` — This command gets an announcement, which is a file that conforms to the Announcement format.
- `agilix-buzz-pp-cli announcements get-info` — This command gets information about an announcement.
- `agilix-buzz-pp-cli announcements get-list` — This command lists an entity's announcements.
- `agilix-buzz-pp-cli announcements get-user-list` — This command lists course, section, and domain announcements for the current user.
- `agilix-buzz-pp-cli announcements list-restorable` — This command lists domainor course announcements that have been deleted and can be restored.
- `agilix-buzz-pp-cli announcements put` — This command posts an announcement to the domain or course specified by entityid .
- `agilix-buzz-pp-cli announcements restore` — This command restores one or more domain or course announcements.
- `agilix-buzz-pp-cli announcements update-viewed` — This command updates the viewed state of one or more domain announcements for the current signed-on user.

**assessments** — Agilix Buzz assessments commands

- `agilix-buzz-pp-cli assessments get-attempt` — This command returns the data needed for an assessment or homework attempt.
- `agilix-buzz-pp-cli assessments get-attempt-file` — This command gets a uploaded file associated with a fileupload question.
- `agilix-buzz-pp-cli assessments get-attempt-review` — This command returns the data needed to review an assessment or homework attempt.
- `agilix-buzz-pp-cli assessments get-next-question` — Submits answers for the current question in an adpative assessment.
- `agilix-buzz-pp-cli assessments get-question` — This command gets one question from a course.
- `agilix-buzz-pp-cli assessments get-question-scores` — This command scores one or more question submissions.
- `agilix-buzz-pp-cli assessments get-question-stats` — This command gets system-wide statistics about a particular question.
- `agilix-buzz-pp-cli assessments get-submission-state` — Retrieves state information for an enrollment's assessment or homework submission
- `agilix-buzz-pp-cli assessments list-questions` — This command lists one or more questions in a course.
- `agilix-buzz-pp-cli assessments list-restorable-questions` — This command lists questions that have been deleted and can be restored.
- `agilix-buzz-pp-cli assessments put-attempt-file` — This command puts files on the server to be associated with a fileupload question as part of a student
- `agilix-buzz-pp-cli assessments put-questions` — This command puts (adds or updates) one or more questions in a course.
- `agilix-buzz-pp-cli assessments restore-questions` — This command restores one or more questions in a course.
- `agilix-buzz-pp-cli assessments save-attempt-answers` — Save answers for an assessment of homework group attempt for later use.
- `agilix-buzz-pp-cli assessments submit-attempt-answers` — Submits answers for an assessment of homework group attempt for grading.

**badges** — Badges

- `agilix-buzz-pp-cli badges create` — This command creates a badge.
- `agilix-buzz-pp-cli badges get` — This command gets the badge image.
- `agilix-buzz-pp-cli badges get-assertion` — This command gets the assertion associated with a badge.
- `agilix-buzz-pp-cli badges get-list` — This command gets a list of badges for a user.

**blogs** — Blogs

- `agilix-buzz-pp-cli blogs get` — This command gets a blog or journal message, which conforms to the Message format.
- `agilix-buzz-pp-cli blogs get-list` — This command returns a list of blog or journal messages for the specified enrollment and item
- `agilix-buzz-pp-cli blogs get-recent-posts` — This command returns recent posts to discussion boards, wikis, blogs and journals that have not been marked as viewed.
- `agilix-buzz-pp-cli blogs get-summary` — This command returns a blog or journal summary for each enrollment in the specified entity.
- `agilix-buzz-pp-cli blogs put` — This command puts a blog or journal message to the server.
- `agilix-buzz-pp-cli blogs update-viewed` — This command updates a user’s viewed state of one or more blog messages.

**conversion** — Format conversion

- `agilix-buzz-pp-cli conversion export-data` — This command convert structured post data to a tab or comma-delimited file.
- `agilix-buzz-pp-cli conversion get-converted-data` — Retrieves the converted data generated from either ImportData or ExportData and deletes the temporary file.
- `agilix-buzz-pp-cli conversion import-data` — This command imports and converts various data formats.

**courses** — Courses & course copy

- `agilix-buzz-pp-cli courses copy` — This command copies one or more courses.
- `agilix-buzz-pp-cli courses create` — This command creates one or more courses.
- `agilix-buzz-pp-cli courses create-demo` — This command creates a demo course.
- `agilix-buzz-pp-cli courses deactivate` — This command deactivates a course.
- `agilix-buzz-pp-cli courses get` — This command gets information for a course.
- `agilix-buzz-pp-cli courses list` — This command lists courses.
- `agilix-buzz-pp-cli courses merge` — This command merges the deltas from a derivative course into its immediate base course.
- `agilix-buzz-pp-cli courses restore` — This command restores a deleted course.
- `agilix-buzz-pp-cli courses update` — This command updates the title, reference, data (free-form XML), and other attributes of one or more courses.

**datastreams** — Data streams

- `agilix-buzz-pp-cli datastreams get-data-stream-configuration` — This command gets the data stream configuration previously set using SetDataStreamConfiguration for a specified domain.
- `agilix-buzz-pp-cli datastreams set-data-stream-configuration` — This command tests a given data stream configuration to see if it is valid and if the API servers can connect to it and

**discussions** — Discussion boards

- `agilix-buzz-pp-cli discussions get-message` — This command returns a discussion forum message.
- `agilix-buzz-pp-cli discussions get-message-list` — This command returns the messages associated with a discussion board of an entity.
- `agilix-buzz-pp-cli discussions list-restorable-messages` — This command lists messages that have been deleted from a entity and can be restored.
- `agilix-buzz-pp-cli discussions put-message` — This command puts a discussion board message to the server.
- `agilix-buzz-pp-cli discussions put-message-part` — This command puts individual parts of a discussion board message to the server.
- `agilix-buzz-pp-cli discussions restore-messages` — This command restores one or more messages in a discussion forum.
- `agilix-buzz-pp-cli discussions submit-message` — This command assembles changes from the PutMessagePart and DeleteMessagePart commands into a new message and posts the
- `agilix-buzz-pp-cli discussions update-message-viewed` — This command updates a user’s viewed state of one or more discussion board messages.

**domains** — Domains (schools/districts) & hierarchy

- `agilix-buzz-pp-cli domains create` — This command creates one or more domains and links them as children to the domain specified by parentid.
- `agilix-buzz-pp-cli domains get` — This command gets information for the specified domain.
- `agilix-buzz-pp-cli domains get-content` — This command returns the list of content items for the current signed-on user’s domain.
- `agilix-buzz-pp-cli domains get-enrollment-metrics` — This command gets enrollment metrics for courses in the specified domain.
- `agilix-buzz-pp-cli domains get-enrollment-metrics-parameters` — This command gets parameters that affect calculations used in Enrollment Metrics , Course EnrollmentMetrics
- `agilix-buzz-pp-cli domains get-parent-list` — This command gets the list of parent domains for a domain.
- `agilix-buzz-pp-cli domains get-stats` — This command gets statistics for a domain.
- `agilix-buzz-pp-cli domains list` — This command lists domains.
- `agilix-buzz-pp-cli domains restore` — This command restores a deleted domain.
- `agilix-buzz-pp-cli domains update` — This command updates the name, reference value, flags, and free-forn structured data of one or more domains.

**enrollments** — Enrollments (user-course bindings) & activity

- `agilix-buzz-pp-cli enrollments create` — This command enrolls a user with the specified rights in a course.
- `agilix-buzz-pp-cli enrollments get` — This command gets enrollment data for a particular enrollment.
- `agilix-buzz-pp-cli enrollments get-activity` — This command gets the activity detail for the specified user enrollment.
- `agilix-buzz-pp-cli enrollments get-gradebook` — This command gets the gradebook detail, including rolled-up period, category, and course grades
- `agilix-buzz-pp-cli enrollments get-group-list` — This command lists groups that the specified enrollment is a member of.
- `agilix-buzz-pp-cli enrollments get-metrics-report` — Returns report data for the enrollment metrics of a domain, teacher, or student.
- `agilix-buzz-pp-cli enrollments list` — This command lists enrollments.
- `agilix-buzz-pp-cli enrollments list-by-teacher` — This command lists enrollments in courses where the specified user is a teacher.
- `agilix-buzz-pp-cli enrollments list-entity` — This command lists enrollments in the specified entity.
- `agilix-buzz-pp-cli enrollments list-user` — This command lists enrollments for the specified user.
- `agilix-buzz-pp-cli enrollments put-self-assessment` — This command puts self-assessment ratings of understanding, interest, and effort into the student's Enrollment Metrics .
- `agilix-buzz-pp-cli enrollments restore` — This command restores a deleted enrollment.
- `agilix-buzz-pp-cli enrollments update` — This command updates the enrollment status, start date, end date, and rights granted to the specified user for a course.

**files** — Course resources & files

- `agilix-buzz-pp-cli files copy` — This command copies one or more resources from the domain or course specified by sourceentityid to the domain or course
- `agilix-buzz-pp-cli files get` — This command gets resource metadata and/or binary content for the course, section, enrollment, user, or domain.
- `agilix-buzz-pp-cli files get-document` — This command retrieves a user document from the server.
- `agilix-buzz-pp-cli files get-document-info` — This command retrieves information about one or more user documents from the server.
- `agilix-buzz-pp-cli files get-entity-id` — This command gets the resource entity ID of the specified entity.
- `agilix-buzz-pp-cli files get-info` — This command gets information about a resource in the domain, course, section, or enrollment specified by entityid .
- `agilix-buzz-pp-cli files get-list` — This command lists metadata for entity (domain, course, or enrollment) resources.
- `agilix-buzz-pp-cli files list-restorable` — This command lists resources that have been deleted and can be restored.
- `agilix-buzz-pp-cli files list-restorable-documents` — This command lists documents that have been deleted and can be restored.
- `agilix-buzz-pp-cli files put` — This command puts a resource to the server for the specified course, section, enrollment, user, or domain.
- `agilix-buzz-pp-cli files put-folders` — This command creates one or more resource folders for a domain or course.
- `agilix-buzz-pp-cli files restore` — This command restores one or more resources in a domain, course, or enrollment.
- `agilix-buzz-pp-cli files restore-documents` — This command restores one or more documents.

**gradebook** — Gradebook results & grading

- `agilix-buzz-pp-cli gradebook calculate-enrollment-scenario` — This command calculates a rolled-up (category, period, and course)
- `agilix-buzz-pp-cli gradebook get-calendar` — Gets the iCalendar (RFC2445) file of events for the user or enrollments specified by token.
- `agilix-buzz-pp-cli gradebook get-calendar-items` — This command gets duedates and blackoutdates for the specified enrollments.
- `agilix-buzz-pp-cli gradebook get-calendar-token` — Get a token to generate a URL that can retrieve the current user's calendar without any authentication.
- `agilix-buzz-pp-cli gradebook get-certificates` — This command gets the completion certificates associated with a course or section enrollment.
- `agilix-buzz-pp-cli gradebook get-due-soon-list` — This command gets list of items that will soon become due for a student.
- `agilix-buzz-pp-cli gradebook get-entity-summary` — This command gets grade summaries for students enrolled in the specified entity.
- `agilix-buzz-pp-cli gradebook get-entity-work` — Gets the list of items that have been graded or need grading in a course back to a specified date.
- `agilix-buzz-pp-cli gradebook get-grade` — This command gets the grade detail for the specified user enrollment and item.
- `agilix-buzz-pp-cli gradebook get-grade-history` — This command gets the history of grades for an item in the gradebook for the specified user enrollment.
- `agilix-buzz-pp-cli gradebook get-item-analysis` — Some items compute their score from scores assigned within the item.
- `agilix-buzz-pp-cli gradebook get-item-report` — Some items compute their score from scores assigned within the item.
- `agilix-buzz-pp-cli gradebook get-list` — This command lists entities (courses and sections) that have gradebooks.
- `agilix-buzz-pp-cli gradebook get-rubric-mastery` — This command gets report data that shows how students are performing for items related to the specified rubric.
- `agilix-buzz-pp-cli gradebook get-rubric-stats` — This command gets detailed information about responses and scores associated with the rules in a rubric.
- `agilix-buzz-pp-cli gradebook get-summary` — This command gets a summary of course participants that match the specified parameters.
- `agilix-buzz-pp-cli gradebook get-user` — This command gets gradebook detail for the specified user.
- `agilix-buzz-pp-cli gradebook get-weights` — This command gets the item and category gradebook weights for the specified entity
- `agilix-buzz-pp-cli gradebook getentitygradebook2` — This command gets grades for all students enrolled in the specified entity.
- `agilix-buzz-pp-cli gradebook getentitygradebook3` — This command gets grades for students enrolled in the specified entity.

**groups** — Groups

- `agilix-buzz-pp-cli groups add-members` — This command adds one or more member enrollments to an existing group.
- `agilix-buzz-pp-cli groups create` — This command creates one or more groups in the specified owner course.
- `agilix-buzz-pp-cli groups get` — This command gets information for a particular group.
- `agilix-buzz-pp-cli groups get-enrollment-list` — This command lists enrollments for the specified group.
- `agilix-buzz-pp-cli groups get-list` — This command lists groups in the specified owner entity (course).
- `agilix-buzz-pp-cli groups remove-members` — This command removes one or more member enrollments from an existing group.
- `agilix-buzz-pp-cli groups update` — This command updates the title, set ID, reference, and data of one or more groups.

**items** — Manifests & content items

- `agilix-buzz-pp-cli items assign` — Assigning an item to a folder changes the item’s parent and sequence to the specified values.
- `agilix-buzz-pp-cli items copy` — This command copies one or more items from one course to another.
- `agilix-buzz-pp-cli items find-personalized-entities` — This command finds the course, enrollment
- `agilix-buzz-pp-cli items get-data` — This command gets the manifest data (see Course Data ) for the specified entity (course, section, group, or enrollment)
- `agilix-buzz-pp-cli items get-links` — This command gets a list of manifest items that link to the specified item through course or item chaining.
- `agilix-buzz-pp-cli items get-list` — This command lists items.
- `agilix-buzz-pp-cli items getitem` — This command gets a manifest item.
- `agilix-buzz-pp-cli items getiteminfo` — This command gets information about an item in the course, section or group specified by entityid .
- `agilix-buzz-pp-cli items getmanifest` — This command gets the manifest for the specified entity (course, section, group, or enrollment).
- `agilix-buzz-pp-cli items getmanifestinfo` — This command gets information about the manifest of one or more courses or sections.
- `agilix-buzz-pp-cli items getmanifestitem` — This command gets an item and optionally some of its descendents from the manifest of a course or section.
- `agilix-buzz-pp-cli items list-assignable` — ListAssignableItems uses the assignableitemsquery attribute on the folderid item to list items in the course that may
- `agilix-buzz-pp-cli items list-restorable` — This command lists items that have been deleted and can be restored.
- `agilix-buzz-pp-cli items navigate` — This command attempts to navigate a user's enrollment to a course item.
- `agilix-buzz-pp-cli items put` — This command puts one or more items in a manifest.
- `agilix-buzz-pp-cli items restore` — This command restores one or more items in a manifest.
- `agilix-buzz-pp-cli items search` — This command searches course content that has been marked by course creators as searchable and returns a paginated list
- `agilix-buzz-pp-cli items unassign` — Unassigning an item changes the item’s parent and sequence to their default values (the values they would have if they
- `agilix-buzz-pp-cli items update-data` — This command updates the manifest data (see Course Data ) on the specified entity (course or section.

**objectives** — Learning objectives & progress

- `agilix-buzz-pp-cli objectives create-sets` — This command creates one or more objective sets or objective map sets
- `agilix-buzz-pp-cli objectives get-list` — This command gets a list of learning objectives.
- `agilix-buzz-pp-cli objectives get-map-list` — This command gets the list of maps defined in an objective map set.
- `agilix-buzz-pp-cli objectives get-mastery` — This command gets objective mastery report data for the specified entity using student scores from objective-aligned
- `agilix-buzz-pp-cli objectives get-mastery-detail` — This command summarizes a single learning objective's mastery for each enrollment in the specified entities.
- `agilix-buzz-pp-cli objectives get-mastery-summary` — This command lists learning objectives and summarizes their mastery for enrollments in the specified entities.
- `agilix-buzz-pp-cli objectives get-set` — This command gets information for a objective set or objective map set.
- `agilix-buzz-pp-cli objectives get-subject-list` — This command gets the list of subjects covered by the learning objectives for an objective set.
- `agilix-buzz-pp-cli objectives list-sets` — This command gets a list of objective sets or objective map sets.
- `agilix-buzz-pp-cli objectives put` — This creates or updates one or more learning objectives and puts them in an objective set.
- `agilix-buzz-pp-cli objectives put-maps` — This command creates or updates one or more objective maps and puts them into an objective map set.
- `agilix-buzz-pp-cli objectives restore-set` — This command restores a deleted objective set.
- `agilix-buzz-pp-cli objectives update-sets` — This command updates the name, reference, and other attributes of one or more objective sets or objective map sets.

**peergrading** — Peer grading

- `agilix-buzz-pp-cli peergrading get-response` — This command gets a peer's response to a student's submission.
- `agilix-buzz-pp-cli peergrading get-response-info` — This command retrieves information about peer responses from the server.
- `agilix-buzz-pp-cli peergrading get-response-list` — This command lists peer's responses to the specified student's submission to a course item.
- `agilix-buzz-pp-cli peergrading get-review-list` — This command lists peers that may be reviewed by the specified enrollment user.
- `agilix-buzz-pp-cli peergrading put-response` — This command puts peer response data including comments, rubric score, and likert responses to the server.

**ratings** — Ratings

- `agilix-buzz-pp-cli ratings get-item` — This command gets a user rating on a manifest item.
- `agilix-buzz-pp-cli ratings get-item-summary` — This command gets the summary of user ratings on a manifest item.
- `agilix-buzz-pp-cli ratings put-item` — This command puts a user rating on a manifest item.

**reports** — Reports

- `agilix-buzz-pp-cli reports get-info` — This command lists the information about a report including the parameters required to run it.
- `agilix-buzz-pp-cli reports get-list` — This command lists all reports defined on a domain.
- `agilix-buzz-pp-cli reports get-runnable-list` — This command lists all reports available to the currently logged-on user.
- `agilix-buzz-pp-cli reports run` — This command runs the specified report.

**rights** — Rights & role assignments

- `agilix-buzz-pp-cli rights create-role` — This command creates a role on a given domain.
- `agilix-buzz-pp-cli rights get` — This command gets the rights granted to the specified actor (user or role) for the specified entity (domain, course
- `agilix-buzz-pp-cli rights get-actor` — This command lists entities that an actor (user) has rights for.
- `agilix-buzz-pp-cli rights get-effective` — This command gets the effective rights granted to the current user for the specified entity (domain, course, or section)
- `agilix-buzz-pp-cli rights get-effective-subscription-list` — This command lists effective subscriptions for the current signed-in user
- `agilix-buzz-pp-cli rights get-entity` — This command lists users and the rights granted to those users for the specified entity (domain, course, section
- `agilix-buzz-pp-cli rights get-entity-subscription-list` — This command lists subscriptions to the specified entity (course or domain).
- `agilix-buzz-pp-cli rights get-list` — This command lists rights granted to users on entities (domains, users, and enrollments).
- `agilix-buzz-pp-cli rights get-personas` — This command gets the list of personas associated with a user.
- `agilix-buzz-pp-cli rights get-role` — This command gets information for a role.
- `agilix-buzz-pp-cli rights get-subscription-list` — This command lists subscriptions explicitly assigned to the specified subscriber (user or domain).
- `agilix-buzz-pp-cli rights list-roles` — This command lists roles defined for a domain.
- `agilix-buzz-pp-cli rights restore-role` — This command restores a role that has been previously deleted.
- `agilix-buzz-pp-cli rights update` — This command updates the rights (flags) granted to the specified user (actorid)
- `agilix-buzz-pp-cli rights update-role` — This command updates the name, domainid, reference, or privileges for a role.
- `agilix-buzz-pp-cli rights update-subscriptions` — This command creates or updates the subscription for the specified subscriber.

**session** — Authentication & session (login, token, proxy/login-as)

- `agilix-buzz-pp-cli session clear-second-factor-authentication` — Clears 2FA (second factor authentication) settings for the specified user account
- `agilix-buzz-pp-cli session create-second-factor-authentication-secret` — This command creates an RFC 6238 second factor authentication secret for use with 2FA authentication.
- `agilix-buzz-pp-cli session extend-session` — Each API command automatically refreshes the authorization token
- `agilix-buzz-pp-cli session finish-password-reset` — This command updates the password of the specified user.
- `agilix-buzz-pp-cli session force-password-change` — Forces a user to change their password the next time they login, exactly as if their password had expired.
- `agilix-buzz-pp-cli session get-key` — This command retrieves a name/value pair from an entity.
- `agilix-buzz-pp-cli session get-password-login-attempt-history` — Gets the record of password login attempts for the specified user account.
- `agilix-buzz-pp-cli session get-password-question` — This command gets the question to ask a user who has forgotten their password.
- `agilix-buzz-pp-cli session login` — For server-side API integrations, OAuth 2.
- `agilix-buzz-pp-cli session logout` — This command terminates the session state for an authenticated user.
- `agilix-buzz-pp-cli session proxy` — This command starts a session as a different user.
- `agilix-buzz-pp-cli session put-key` — This command stores a name/value pair on an entity.
- `agilix-buzz-pp-cli session reset-lockout` — This command resets an account that has been locked out due to too many contiguous password failures.
- `agilix-buzz-pp-cli session reset-password` — This command sends an email to the specified user with a time-limited link that can be used to reset their password.
- `agilix-buzz-pp-cli session second-factor-authenticate` — This command completes the two factor part of authentication for users who have opted for two factor authentication.
- `agilix-buzz-pp-cli session setup-second-factor-authentication` — Turns on second factor authentication for an account.
- `agilix-buzz-pp-cli session unproxy` — This command ends a proxy session.
- `agilix-buzz-pp-cli session update-password` — This command updates the password of the specified user.
- `agilix-buzz-pp-cli session update-password-question-answer` — This command updates the password question and answer for the specified user.

**submissions** — Student submissions & responses

- `agilix-buzz-pp-cli submissions get-sco-data` — This command gets a user's SCORM data for a SCO activity from the server.
- `agilix-buzz-pp-cli submissions get-student` — This command gets a student's submission.
- `agilix-buzz-pp-cli submissions get-student-history` — This command gets the history of submissions for an item in the gradebook for the specified user enrollment.
- `agilix-buzz-pp-cli submissions get-student-info` — This command retrieves information about one or more student submissions from the server.
- `agilix-buzz-pp-cli submissions get-teacher-response` — This command gets a teacher's response to a student's submission.
- `agilix-buzz-pp-cli submissions get-teacher-response-info` — This command retrieves information about teacher responses from the server.
- `agilix-buzz-pp-cli submissions get-work-in-progress` — This command gets a work-in-progress file.
- `agilix-buzz-pp-cli submissions put-item-activity` — This command reports per-item student time spent to the server.
- `agilix-buzz-pp-cli submissions put-sco-data` — This command puts a user's SCORM data for a SCO activity to the server.
- `agilix-buzz-pp-cli submissions put-student` — This command puts a student submission for an activity to the server.
- `agilix-buzz-pp-cli submissions put-teacher-response` — This command puts teacher response data including scores, comments, and grade status flags to the server.
- `agilix-buzz-pp-cli submissions put-teacher-responses` — This command puts one or more teacher responses including scores and grade status flags to the server.
- `agilix-buzz-pp-cli submissions put-work-in-progress` — This command puts a work-in-progress file on the server in preparation for a student submission.
- `agilix-buzz-pp-cli submissions submit-work-in-progress` — This command submits work-in-progress files on the server as a completed submission for the student.

**system** — Server status & general utilities

- `agilix-buzz-pp-cli system get-command-list` — This returns a complete list of all commands available on the API Server.
- `agilix-buzz-pp-cli system get-entity-type` — This command gets the entity type (Course, Domain, Group, or Section) for a given entity ID.
- `agilix-buzz-pp-cli system get-status` — The API servers do fairly extensive self-testing for all subsystems.
- `agilix-buzz-pp-cli system send-mail` — This command sends an e-mail to recipients that are enrolled in the same entity as that specified by the sender's

**tokens** — Command tokens (delegated links)

- `agilix-buzz-pp-cli tokens create-command` — This command creates one or more command tokens.
- `agilix-buzz-pp-cli tokens get-command` — This command gets details about a previously created command token.
- `agilix-buzz-pp-cli tokens get-command-info` — This command gets the description and other non-sensitive info about a previously created command token from the code.
- `agilix-buzz-pp-cli tokens list-command` — This command lists all the command tokens created by a particular user and associated with a specified domain, course
- `agilix-buzz-pp-cli tokens redeem-command` — This command redeems a command token using the code given to the currently authenticated user by the creator.
- `agilix-buzz-pp-cli tokens update-command` — This command updates one or more existing command tokens

**users** — User accounts

- `agilix-buzz-pp-cli users create` — This command creates one or more users.
- `agilix-buzz-pp-cli users get` — This command gets information for a particular user.
- `agilix-buzz-pp-cli users get-active-count` — Gets the number of active users either currently or for the given date range.
- `agilix-buzz-pp-cli users get-activity` — This command lists the login activity for a user.
- `agilix-buzz-pp-cli users get-activity-stream` — This command lists activities in a user's activity stream.
- `agilix-buzz-pp-cli users get-domain-activity` — This command lists activity for all users in a specified domain.
- `agilix-buzz-pp-cli users get-profile-picture` — This command gets binary content for the user's profile picture if it exists.
- `agilix-buzz-pp-cli users list` — This command lists users.
- `agilix-buzz-pp-cli users restore` — This command restores a deleted user.
- `agilix-buzz-pp-cli users update` — This command updates the first name, last name, reference field, e-mail address, and flags for one or more users.
- `agilix-buzz-pp-cli users verify-email` — This command verifies a messaging account associated with a user.

**wikis** — Wikis

- `agilix-buzz-pp-cli wikis copy-pages` — This command copies one or more wiki pages from the specified source course
- `agilix-buzz-pp-cli wikis get-page-list` — This command lists metadata for wiki pages.
- `agilix-buzz-pp-cli wikis list-restorable-pages` — This command lists wiki pages that have been deleted from a course or section and can be restored.
- `agilix-buzz-pp-cli wikis put-page` — This command puts a wiki page to the server.
- `agilix-buzz-pp-cli wikis restore-pages` — This command restores one or more wiki pages in a course or section.
- `agilix-buzz-pp-cli wikis update-page-viewed` — This command updates a user’s viewed state of one or more wiki pages.


### Finding the right command

When you know what you want to do but not which command does it, ask the CLI directly:

```bash
agilix-buzz-pp-cli which "<capability in your own words>"
```

`which` resolves a natural-language capability query to the best matching command from this CLI's curated feature index. Exit code `0` means at least one match; exit code `2` means no confident match — fall back to `--help` or use a narrower query.

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

## Auth Setup

Buzz issues short-lived (~15 minute) session tokens that auto-extend on every API call. Quickest path: export BUZZ_TOKEN with a token (mint one with 'session login --username domain/user --password ...'), or save it with 'agilix-buzz-pp-cli auth set-token <token>'. Run 'agilix-buzz-pp-cli auth status' to confirm and 'agilix-buzz-pp-cli doctor' to check connectivity. Delete commands are intentionally omitted from this build per least-privilege policy.

Run `agilix-buzz-pp-cli doctor` to verify setup.

## Agent Mode

Add `--agent` to any command. Expands to: `--json --compact --no-input --no-color --yes`.

- **Pipeable** — JSON on stdout, errors on stderr
- **Filterable** — `--select` keeps a subset of fields. Dotted paths descend into nested structures; arrays traverse element-wise. Critical for keeping context small on verbose APIs:

  ```bash
  agilix-buzz-pp-cli announcements get --agent --select id,name,status
  ```
- **Previewable** — `--dry-run` shows the request without sending
- **Offline-friendly** — sync/search commands can use the local SQLite store when available
- **Non-interactive** — never prompts, every input is a flag
- **Explicit retries** — use `--idempotent` only when an already-existing create should count as success

### Response envelope

Commands that read from the local store or the API wrap output in a provenance envelope:

```json
{
  "meta": {"source": "live" | "local", "synced_at": "...", "reason": "..."},
  "results": <data>
}
```

Parse `.results` for data and `.meta.source` to know whether it's live or local. A human-readable `N results (live)` summary is printed to stderr only when stdout is a terminal AND no machine-format flag (`--json`, `--csv`, `--compact`, `--quiet`, `--plain`, `--select`) is set — piped/agent consumers and explicit-format runs get pure JSON on stdout.

## Paths and state

Agents should treat the CLI's path resolver as part of the runtime contract:

- Use `--home <dir>` for one invocation, or set `AGILIX_BUZZ_HOME=<dir>` to relocate all four path kinds under one root.
- Use per-kind env vars only when a specific kind must diverge: `AGILIX_BUZZ_CONFIG_DIR`, `AGILIX_BUZZ_DATA_DIR`, `AGILIX_BUZZ_STATE_DIR`, `AGILIX_BUZZ_CACHE_DIR`.
- Resolution order is per-kind env var, `--home`, `AGILIX_BUZZ_HOME`, XDG (`XDG_CONFIG_HOME`, `XDG_DATA_HOME`, `XDG_STATE_HOME`, `XDG_CACHE_HOME`), then platform defaults.
- `config` contains settings like `config.toml` and profiles. `data` contains `credentials.toml`, `data.db`, cookies, and auth sidecars. `state` contains persisted queries, jobs, and `teach.log`. `cache` contains regenerable HTTP/cache files.
- Stored secrets live in `credentials.toml` under the data dir. Existing legacy `config.toml` secrets are read for compatibility and leave `config.toml` on the first auth write.
- Run `agilix-buzz-pp-cli doctor --fail-on warn` to surface path and credential-location warnings. `agent-context` exposes a schema v4 `paths` block for agents that need the resolved dirs.
- For MCP, pass relocation through the MCP host config. The MCP binary does not inherit CLI flags:

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

Fleet precedence: an inherited per-kind env var overrides an explicit `--home` for that kind. Use `AGILIX_BUZZ_HOME` or per-kind vars as durable fleet levers, and use `--home` only for a single invocation. Relocation is not reversible by unsetting env vars; move files manually before clearing `AGILIX_BUZZ_HOME`, or `doctor` will not find credentials left under the former root.

## Agent Feedback

When you (or the agent) notice something off about this CLI, record it:

```
agilix-buzz-pp-cli feedback "the --since flag is inclusive but docs say exclusive"
agilix-buzz-pp-cli feedback --stdin < notes.txt
agilix-buzz-pp-cli feedback list --json --limit 10
```

Entries are stored locally as `feedback.jsonl` under the resolved data dir. They are never POSTed unless `AGILIX_BUZZ_FEEDBACK_ENDPOINT` is set AND either `--send` is passed or `AGILIX_BUZZ_FEEDBACK_AUTO_SEND=true`. Default behavior is local-only.

Write what *surprised* you, not a bug report. Short, specific, one line: that is the part that compounds.

## Output Delivery

Every command accepts `--deliver <sink>`. The output goes to the named sink in addition to (or instead of) stdout, so agents can route command results without hand-piping. Three sinks are supported:

| Sink | Effect |
|------|--------|
| `stdout` | Default; write to stdout only |
| `file:<path>` | Atomically write output to `<path>` (tmp + rename) |
| `webhook:<url>` | POST the output body to the URL (`application/json` or `application/x-ndjson` when `--compact`) |

Unknown schemes are refused with a structured error naming the supported set. Webhook failures return non-zero and log the URL + HTTP status on stderr.

## Named Profiles

A profile is a saved set of flag values, reused across invocations. Use it when a scheduled agent calls the same command every run with the same configuration - HeyGen's "Beacon" pattern.

```
agilix-buzz-pp-cli profile save briefing --json
agilix-buzz-pp-cli --profile briefing announcements get
agilix-buzz-pp-cli profile list --json
agilix-buzz-pp-cli profile show briefing
agilix-buzz-pp-cli profile delete briefing --yes
```

Explicit flags always win over profile values; profile values win over defaults. `agent-context` lists all available profiles under `available_profiles` so introspecting agents discover them at runtime.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 2 | Usage error (wrong arguments) |
| 3 | Resource not found |
| 4 | Authentication required |
| 5 | API error (upstream issue) |
| 7 | Rate limited (wait and retry) |
| 10 | Config error |

## Argument Parsing

Parse `$ARGUMENTS`:

1. **Empty, `help`, or `--help`** → show `agilix-buzz-pp-cli --help` output
2. **Starts with `install`** → ends with `mcp` → MCP installation; otherwise → see Prerequisites above
3. **Anything else** → Direct Use (execute as CLI command with `--agent`)

## MCP Server Installation

1. Install the MCP server:
   ```bash
   go install github.com/mvanhorn/printing-press-library/library/productivity/agilix-buzz/cmd/agilix-buzz-pp-mcp@latest
   ```
2. Register with Claude Code:
   ```bash
   claude mcp add agilix-buzz-pp-mcp -- agilix-buzz-pp-mcp
   ```
3. Verify: `claude mcp list`

## Direct Use

1. Check if installed: `which agilix-buzz-pp-cli`
   If not found, offer to install (see Prerequisites at the top of this skill).
2. Match the user query to the best command from the Unique Capabilities and Command Reference above.
3. Execute with the `--agent` flag:
   ```bash
   agilix-buzz-pp-cli <command> [subcommand] [args] --agent
   ```
4. If ambiguous, drill into subcommand help: `agilix-buzz-pp-cli <command> --help`.

## Authoritative API Reference (for deeper questions)

This CLI mirrors the Agilix Buzz DLAP command surface. For full per-command parameter and
response details beyond `--help`, fetch Agilix's AI-optimized documentation:

- `https://api.agilixbuzz.com/llms.txt` (plain text, llms.txt standard — best for most assistants)
- `https://api.agilixbuzz.com/llms.md` (Markdown variant)
- Per-command docs: `https://api.agilixbuzz.com/docs/entry/Command/<CommandName>.md`

## API Throttling (handled automatically)

The client implements Buzz's documented throttling model so reporting fan-outs stay safe:
- **429 Too Many Requests** (time-limiting and rate-limiting) — honors `Retry-After` and adapts the
  request rate.
- **503 Server Busy** — honors `Retry-After` when present, else exponential backoff.
- **Proactive pacing** — reads the `X-Provisioned-Ms-Remaining` budget header and slows down before
  hitting the limit (Buzz's recommended speedometer approach), so large `--max-pages` analytics runs
  avoid cascading 429s. No flags needed.
