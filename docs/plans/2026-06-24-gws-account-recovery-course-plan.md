# Course Build Plan: Google Workspace Account Recovery for Technology Directors

Created: 2026-06-24

Target course: **Google Workspace Account Recovery** (Buzz course id `267910529`, ref `gwar-training-001`, domain `254591853` / testsandboxidla). Built and verified with the `agilix-buzz-pp-cli`.

---

## 1. Goal & Audience

Transform the existing thin, end-user-flavored course into a researched, assessment-driven course for **technology directors** — the people who *administer* Google Workspace for a district / online school, not the end users who forget their own passwords.

The reframe is total: the learner is the admin. Every lesson answers "what do *I*, the director, own, configure, and do when access is lost or an account is compromised." End-user mechanics appear only as "what your staff/students experience and how you support them."

**Design decisions (confirmed):**
- **Full admin/director reframe** — super-admin recovery, admin-console user recovery, org-wide policy, governance, incident response.
- **Fuller curriculum** — grow from 2 modules to **5 modules** (prevention → super-admin recovery → user recovery → org policy → governance/IR).
- **Knowledge checks + scenario final** — a short formative knowledge-check after each module, plus one graded, scenario-based final assessment.

---

## 2. Current State (live, captured 2026-06-24)

```
DEFAULT  Google Workspace Account Recovery
├─ mod1  Module 1: Recovery Fundamentals
│   ├─ t1  Lesson 1: Identify Your Recovery Options
│   └─ t2  Lesson 2: Set Up Recovery Email & Phone
└─ mod2  Module 2: Performing a Recovery
    ├─ t3  Lesson 3: Account Recovery Form Walkthrough
    └─ t4  Lesson 4: Admin-Assisted Recovery & Hardening
```

- 2 folders, 4 topic (leaf) items. **No assessments, no questions, no gradebook weighting** beyond the default category.
- Existing titles are end-user oriented ("Set Up Recovery Email & Phone", "Account Recovery Form Walkthrough").
- Gradebook scaffold exists: categories (weighted, "Category 1" @ 100%), grade tables (A–F, +/- , P/F), passing score 0.7.

---

## 3. Hard Constraints (read before building)

- **No deletes without explicit approval.** Per `CLAUDE.md`, only read/create/update are permitted. The existing `mod1`, `mod2`, `t1`–`t4` are therefore **updated in place / repurposed**, never deleted. Items no longer needed in their old role are retitled and re-contented to new roles; nothing is removed. If you later decide to remove orphaned items, that requires a separate explicit approval and a `deleteitems` call.
- **Secrets discipline.** Tokens are short-lived (≈15 min) and live only in ephemeral shell commands (`auth set-token`). No token value is ever written into a lesson, resource, or file.
- **Content accuracy / drift.** Google is mid-migration of Admin Help (`support.google.com/a/answer/*` → `knowledge.workspace.google.com/admin/*`; old links 301-redirect). Setting labels and default states drift. Every lesson that names a console path or default must carry a **"verify in your console — labels and defaults change"** callout, and cite the canonical Google URL.
- **Throttling.** A full build is many write calls. Respect Buzz time/rate limiting: run builds with `--rate-limit` (e.g. `--rate-limit 3`) and prefer batched bodies (multiple `<item>` / `<question>` per request) over one-call-per-object. The client already honors `Retry-After` / `X-Provisioned-Ms-Remaining`; don't fight it with parallel bursts.

---

## 4. Target Course Structure

5 modules, 16 lessons, 5 knowledge checks, 1 graded final. Existing items are mapped to their new roles in the build map (§6); new items are created.

### Module 1 — Recovery Readiness: Owning the Risk *(prevention)*
- **L1.1 The director owns the recovery problem.** Single points of failure; real district lockout cases (sole admin departs, lost DNS, no break-glass). Why prevention >> heroics.
- **L1.2 Multiple super admins & the break-glass account.** Google's "more than one super admin, each a separate person"; the dedicated, rarely-used emergency admin secured with hardware keys + offline printed backup codes; separate admin identity vs daily account (`alice-admin@` pattern). Note "3–5" is practitioner consensus, not an official Google number.
- **L1.3 The dependencies you must not lose: recovery info, DNS, registrar.** Recovery email/phone hygiene; domain-ownership = DNS control; the registrar credential as a recovery dependency; inventory & secure offline.
- **Knowledge Check 1** (5 Q): readiness fundamentals.

### Module 2 — Super Admin Account Recovery *(break-glass execution)*
- **L2.1 The five recovery paths, in order.** (1) another super admin resets; (2) self-recovery via recovery email/phone; (3) domain-ownership self-recovery (CNAME/TXT); (4) support-assisted escalation; (5) user-promotion via the recovery form when admins are unreachable.
- **L2.2 The domain-ownership escalation, walked.** `toolbox.googleapps.com/apps/recovery/form`, reference number → CNAME/TXT on the **primary** domain, propagation (up to 24h), Dig verification, "Request for Password Reset", identity verification. The hard dependency: lose DNS = this path fails.
- **L2.3 The super-admin self-recovery toggle & Education defaults.** Security → Authentication → Account recovery → "Allow super admins to recover their account." **Default OFF for Education Standard/Plus.** Why OFF + multiple admins is the more secure posture; phone recovery is always on for super admins regardless.
- **Knowledge Check 2** (5 Q): paths & the DNS escalation.

### Module 3 — Admin-Assisted End-User Recovery *(daily help-desk reality)*
- **L3.1 Password resets from the Admin Console.** Directory → Users → Reset password; auto-generate vs set; "Ask the user to change their password"; **does not work under third-party SSO / Password Sync** (managed in the IdP/AD).
- **L3.2 Restoring deleted users.** **20-day window** (confirmed current); license must be available (the K-12 gotcha — incl. Archived User licenses needing manual reassignment); username/alias conflict resolution; ~24h processing.
- **L3.3 2SV lockouts: backup codes, lost keys & passkeys.** "Only the user can turn on 2SV" — the admin remedy is **Get backup verification codes** (Directory → Users → user → Security); removing a lost security key/passkey; user re-enrolls their own factor; only super admins can generate codes for another admin.
- **L3.4 Suspend/reactivate & the self-recovery toggle.** Reactivate suspended users (abuse/TOS suspensions need Google Support); "Allow users and non-super admins to recover their account" — **unavailable for under-18 Education users** (the single biggest K-12 point: student resets route through the admin).
- **Knowledge Check 3** (6 Q): user-recovery actions & constraints.

### Module 4 — Org-Wide Controls: 2SV, Passkeys & Policy *(configuration)*
- **L4.1 2SV methods in 2026.** Passkeys & security keys (phishing-resistant, "Only security key" now accepts both); Google prompt; Authenticator TOTP; backup codes; **SMS discouraged**. Allowed-methods policy options.
- **L4.2 Enforcement without lockouts.** Off / On / enforce-from-date; new-user enrollment grace (1 day–6 months); the config-group pattern (enforcement OFF for not-yet-enrolled, enroll, then move into enforcing OU); **group settings override OU**; track enrollment via security reports.
- **L4.3 Passkeys, passwordless & Advanced Protection.** Skip-passwords ("Allow users to skip their password and authenticate with a passkey", default OFF); APP for high-risk admins (security-key/passkey-only, non-overridable) and its **multi-day recovery delay** — pre-stage backup factors before enrolling.
- **L4.4 The backup-codes program.** Ten single-use codes; new set invalidates old; admin generation path & the super-admin-only rule for other admins; offline storage best practice. The linchpin that prevents enforcement-day lockouts.
- **Knowledge Check 4** (6 Q): policy & the enforce-without-lockout pattern.

### Module 5 — Governance, Audit & Incident Response *(oversight)*
- **L5.1 K-12 specifics.** Student vs staff OUs as the policy mechanism; recovery info **OFF by default for K-12 students** (Dec 2024 update); COPPA/FERPA implications (avoid personal recovery channels on student accounts; admin-driven resets).
- **L5.2 Audit & monitoring.** Reporting → Audit and investigation → Admin log events (config/role/recovery-setting changes) & User log events (sign-ins, 2SV); Alert center rules to keep on (Suspicious login, User granted Admin privilege, Government-backed attacks).
- **L5.3 Compromised vs locked-out: the 4-step order.** **Suspend** (resets cookies + OAuth tokens) → **Investigate** (admin audit log first, devices, login events) → **Revoke** (reset password, revoke OAuth tokens, remove app passwords) → **Restore** (unsuspend, temp password, re-enroll 2SV, Gmail checklist: forwarding/filters/delegation). Password reset alone is insufficient.
- **Knowledge Check 5** (6 Q): governance & IR ordering.

### Final — Scenario-Based Graded Assessment
- 10–12 scenario questions ("Your only super admin left Friday and you're locked out — what's the FIRST correct action?"; "A teacher's account is sending phishing — order the response"; "A student in 4th grade forgot their password — what's the supported path?"). Mix of multiple-choice (single best / first action) and multi-select (all that apply). Graded, weighted in the gradebook, passing 70%.

---

## 5. Assessment & Interaction Design

| Element | Type | Grading | Attempts | Question mix |
|--------|------|---------|----------|--------------|
| Knowledge Checks 1–5 | Per-module formative quiz | Practice / ungraded (or 0-weight) | Unlimited | MC, multi-select, true/false |
| Final Assessment | Scenario-based summative | Graded, weighted, pass ≥ 70% | 2 | Scenario MC ("first/best action") + multi-select ("all that apply") |

**Question authoring principles:**
- Every question carries answer-level **feedback** that teaches (why the right answer is right, why a tempting wrong answer is wrong) — feedback is where the recovery nuance lands.
- Scenario questions test *ordering and judgment* (suspend-before-investigate, second-admin-before-support), not trivia.
- Distractors are realistic director mistakes (e.g., "reset the password" as the first IR step — wrong because it skips containment).
- Keep stems console-label-light so question content survives Google UI drift; put the volatile labels in the lessons (with verify callouts), not the graded items.

**Build mechanism:** questions are pushed to the course-level bank with `assessments put-questions --entityid 267910529 --stdin` (batched JSON, multiple questions per call), then bound to their assessment item. Verify with `assessments list-questions`. Exact question markup (Buzz question schema) is an execution-time detail — author declaratively (stem, options, correct, feedback, points) and translate to the schema during `ce-work`.

---

## 6. Build Map: Existing Items → New Roles (update, don't delete)

| Existing id | Old title | New role (UPDATE in place) |
|-------------|-----------|----------------------------|
| `mod1` | Module 1: Recovery Fundamentals | **Module 1 — Recovery Readiness** |
| `t1` | Lesson 1: Identify Your Recovery Options | **L1.1 The director owns the recovery problem** |
| `t2` | Lesson 2: Set Up Recovery Email & Phone | **L1.2 Multiple super admins & break-glass** |
| `mod2` | Module 2: Performing a Recovery | **Module 2 — Super Admin Account Recovery** |
| `t3` | Lesson 3: Account Recovery Form Walkthrough | **L2.2 Domain-ownership escalation, walked** |
| `t4` | Lesson 4: Admin-Assisted Recovery & Hardening | **L2.1 The five recovery paths, in order** |

Everything else (L1.3, L2.3, all of Modules 3–5, all knowledge checks, the final) is **created new**. New module folders and lesson topics are added with `items put` (parent = manifest or module folder; itemtype folder/topic); assessment items added with `items put` (itemtype assessment). New lesson resource pages uploaded with `files put --file`.

---

## 7. Build Sequence (runbook)

Ordered by dependency. Use `--agent` for machine-readable output, `--rate-limit 3`, and `--dry-run` first on each new write shape.

1. **Auth & snapshot.** `auth set-token <token>`; capture a baseline `items getmanifest --entityid 267910529 --agent > baseline-manifest.json` so the build is reversible-by-update and diffable.
2. **Module skeleton.** Update `mod1`/`mod2` titles; create module folders M3, M4, M5 under the manifest root (`items put`, batched). Verify with `content tree 267910529`.
3. **Lesson items.** Update `t1`–`t4` to their new roles/titles; create the remaining lesson topic items under each module, in sequence order (`items put`, batched per module).
4. **Lesson content.** Author each lesson as an HTML resource (grounded in §8 research + citations, with verify-in-console callouts); upload with `files put --entityid 267910529 --path <Resources/...> --file <local.html>`; bind the resource to its topic item. Author content offline first, review, then upload.
5. **Assessment items.** Create 5 knowledge-check assessment items (one per module) + 1 final assessment item (`items put`, itemtype assessment; set attempt limits, practice vs graded, category/weight).
6. **Questions.** Author question banks declaratively, then `assessments put-questions --entityid 267910529 --stdin` (batched per assessment); bind questions to their assessment items. Verify with `assessments list-questions`.
7. **Gradebook.** Confirm the final assessment is in a graded, weighted category with passing score 0.7; knowledge checks practice/0-weight. Verify via `items getmanifest` categories/weights.
8. **Verification pass.** `content tree 267910529` for structure; `items getmanifest` for items/weights; `assessments list-questions` for question counts; spot-check one lesson resource fetch. Confirm 5 modules / 16 lessons / 5 KCs / 1 final all present and ordered.

**Execution posture:** author-and-review content offline before each upload batch — content quality is the point, and re-uploading (update) is cheap while wrong content shipped to a live course is the failure mode. Build incrementally module-by-module and verify after each, rather than one big push.

---

## 8. Research Foundation (sourced, current as of 2026-06)

Authoring must draw on these findings; cite the canonical Google URL in each relevant lesson. Flag any default/label as "verify in console."

**Super-admin recovery (M1, M2)**
- Five recovery paths; the formal escalation proves **DNS control of the primary domain** (CNAME/TXT, up to 24h propagation, Dig verification). Recovering administrator access: https://support.google.com/a/answer/33561
- Super-admin self-recovery **default OFF for Education Standard/Plus**; phone recovery always on for super admins. Toggle: Security → Authentication → Account recovery. https://support.google.com/a/answer/9436964
- Multiple super admins + separate admin identity + hardware keys + printed backup codes offline. Security best practices for admin accounts: https://support.google.com/a/answer/9011373 · GCP super-admin best practices: https://docs.cloud.google.com/resource-manager/docs/super-admin-best-practices
- Recovery info settings (new Dec 2024; K-12 students OFF by default): https://workspaceupdates.googleblog.com/2024/12/new-admin-settings-account-recovery-information.html

**Admin-assisted user recovery (M3)**
- Reset a user's password (SSO/Password-Sync constraint): https://support.google.com/a/answer/6236387
- Restore a recently deleted user — **20-day window**, license & alias gotchas: https://knowledge.workspace.google.com/admin/users/restore-a-recently-deleted-user
- Manage a user's security settings — **Get backup verification codes**, remove lost key/passkey: https://knowledge.workspace.google.com/admin/security/manage-a-users-security-settings
- User self-recovery toggle **unavailable for under-18 Education users**: https://knowledge.workspace.google.com/admin/users/set-up-password-recovery-for-users
- Restore a suspended user (abuse/TOS → Support): https://support.google.com/a/answer/1110339

**Org-wide 2SV / passkeys / policy (M4)**
- Deploy 2SV — allowed methods, enforcement, enrollment grace, SMS discouraged: https://knowledge.workspace.google.com/admin/security/deploy-2-step-verification
- Avoid lockouts when 2SV enforced — config-group pattern: https://knowledge.workspace.google.com/admin/security/avoid-account-lockouts-when-2-step-verification-is-enforced-by-your-organization
- **Mandatory admin 2SV** for Education edition (Google-set, non-bypassable): https://knowledge.workspace.google.com/admin/security/about-2sv-enforcement-for-admins
- Skip passwords with passkeys (default OFF): https://knowledge.workspace.google.com/admin/users/allow-users-to-skip-passwords-at-sign-in
- Advanced Protection Program (recovery delay; passkeys added Jul 2024): https://knowledge.workspace.google.com/admin/security/protect-users-with-the-advanced-protection-program
- Backup codes (ten single-use; storage): https://support.google.com/accounts/answer/1187538

**Governance, audit, IR (M5)**
- Identify and secure compromised accounts — the 4-step order: https://knowledge.workspace.google.com/admin/support/troubleshooting/identify-and-secure-compromised-accounts
- Admin log events / User log events: https://knowledge.workspace.google.com/admin/reports/admin-log-events · https://knowledge.workspace.google.com/admin/reports/user-log-events
- FERPA (Google Cloud): https://cloud.google.com/security/compliance/ferpa

**Verify-before-publish flags** (research-noted likely drift): super-admin recovery default-state edition list; whether the Dec-2024 recovery-info toggles consolidated further; skip-passwords default label/state; the per-user "turn off login challenges" duration (distinct from 2SV enforcement); APP default self-enroll. Treat the "3–5 super admins" figure as practitioner consensus, not official.

---

## 9. Risks & Mitigations

- **Content goes stale / wrong default stated.** → Verify-in-console callouts on every settings reference; cite canonical URLs; keep volatile labels out of graded questions.
- **Accidental deletion to "clean up."** → Update-in-place mapping (§6); deletes gated behind explicit approval.
- **Throttling during a large build.** → `--rate-limit`, batched bodies, incremental module-by-module builds; rely on the client's `Retry-After` handling.
- **Token expiry mid-build (~15 min).** → Build in small verifiable batches; re-auth between batches; the runbook is resumable because each step is an idempotent update.
- **Over-scoping.** → 5 modules is the ceiling; if build effort runs long, ship Modules 1–3 + final first (the core recovery competency), then add 4–5.
- **Question quality drift to trivia.** → Scenario/judgment questions with teaching feedback; distractors = realistic director mistakes.

---

## 10. Out of Scope / Deferred

- Deleting or archiving the old item structure (gated; not needed — items are repurposed).
- Enrolling test students or running analytics against the course (separate task; the analytics commands already exist in the CLI).
- SCORM/xAPI interactive packages, video production, or external H5P-style interactions — Buzz-native assessments cover the "quizzes and interactions" ask.
- Multi-language / accessibility localization passes.
- Publishing/copying the course to a production domain.

---

## 11. Definition of Done

- 5 modules, 16 lessons, 5 knowledge checks, 1 graded final present and correctly ordered (`content tree` confirms).
- Every lesson has researched HTML content with citations and verify-in-console callouts.
- Each knowledge check has its questions (MC/multi-select/T-F) with teaching feedback; final has 10–12 scenario questions, graded, pass ≥ 70%.
- Existing items repurposed per §6; nothing deleted.
- Build done within throttle limits; final verification pass clean.
