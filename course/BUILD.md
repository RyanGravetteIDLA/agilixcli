# Rebuilding the course from source

These files reconstruct the **Google Workspace Account Recovery** course in a
Buzz course via `agilix-buzz-pp-cli`. The target sandbox course used during
authoring was id `267910529` (domain `254591853`); change the `--entityid` to
target a different course.

## Prerequisites

- `agilix-buzz-pp-cli` built and on PATH.
- A valid Buzz API token: `agilix-buzz-pp-cli auth set-token <token>`
  (token value only ever in the shell — never committed).

## Build order

All write batches respect API throttling with `--rate-limit 3` and are
idempotent (re-running updates in place; nothing is deleted).

```sh
CID=267910529

# 1. Structure: module folders + lesson items (titles, parents, sequence, href).
cat structure/skeleton.json | agilix-buzz-pp-cli items put --entityid $CID --stdin --rate-limit 3 --agent

# 2. Lesson content: upload each HTML resource to Assets/<id>.html
for f in lessons/*.html; do
  id=$(basename "$f" .html)
  agilix-buzz-pp-cli files put --entityid $CID --path "Assets/$id.html" --file "$f" --rate-limit 3 --agent
done

# 3. Question banks: push schema-2 questions (course-level bank).
for q in questions/q_kc1 questions/q_kc2 questions/q_kc3 questions/q_kc4 questions/q_kc5 questions/q_final; do
  cat "$q.json" | agilix-buzz-pp-cli assessments put-questions --entityid $CID --stdin --rate-limit 3 --agent
done

# 4. Assessment items: knowledge checks + graded final, linking the questions.
cat structure/assess_items.json | agilix-buzz-pp-cli items put --entityid $CID --stdin --rate-limit 3 --agent
```

## Verify

```sh
agilix-buzz-pp-cli content tree $CID                              # structure
agilix-buzz-pp-cli items getmanifest --entityid $CID --agent     # items + weights
agilix-buzz-pp-cli assessments list-questions --entityid $CID --agent
```

## Model notes (DLAP)

- **Items** (`putitems`): each item object carries `entityid`, `itemid`, and a
  `data` block. Updates merge into existing item data. Lessons are
  `data.type = "Resource"` with `data.href = "Assets/<id>.html"`; modules are
  folders; assessments are `data.type = "Assessment"`.
- **Questions** (`putquestions`): schema `"2"`, live in a course-level bank keyed
  by `questionid` (no item binding on the question itself). `interaction.type`
  is `choice`; one `answer.value` entry = single-answer, multiple = multi-select
  (`partial: true`).
- **Linkage**: an assessment item references its questions via
  `data.questions.question[]`, each `{ "id": "<questionid>", "type": 1, "score": N }`
  (`type: 1` = Question Link). Knowledge checks use `weight: 0` (practice); the
  final uses `weight: 100`.
