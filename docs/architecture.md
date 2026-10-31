# Architecture and technical decisions

## Current runtime

One Go HTTP process serves static HTML/CSS/JavaScript and a JSON API. It loads the curriculum separately from learner state. The browser retrieves both, renders dashboard/curriculum/module/study-log/resource views, and sends manual assessment or session changes to the server. All modules remain viewable regardless of prerequisite status.

`main.go` configures environment defaults, offline validation/restore and signal shutdown. `curriculum.go` validates unique stage/module IDs, positive module hours, stage totals, resource-verification flags, existing prerequisites and an acyclic graph. `handlers.go` owns HTTP validation and pass rules. `store.go` owns record validation, legacy normalization and persistence.

## Persistence contract

`State.progress` maps module IDs to the latest status, optional score, notes, evidence references, practical confirmation, update date and (while passed) completion date. `State.sessions` contains append-only records with random IDs, module, minutes, date and note. Completion dates are assigned on passing; editing an existing pass retains its date. Revising clears that active completion date; future attempt-history work must preserve prior attempts separately.

A mutex serializes mutation and export. Each change copies state, serializes JSON, writes a unique mode-0600 temporary file beside the target, syncs/closes it, then renames it over the target. Only then does memory change. Failed persistence leaves the previous logical state intact. Parent directories are created with mode 0750; existing ownership/modes are retained. Atomic replacement assumes a local filesystem with atomic same-directory rename. Directory metadata is not fsynced; this is not a guarantee against all power-loss/filesystem failures. Backups remain necessary.

Startup rejects malformed JSON, unknown fields/modules, invalid statuses/scores/timestamps/sessions and invalid curriculum graphs. It preserves the source file on failure. Legacy `submitted` statuses normalize to `ready_for_assessment`. Legacy/incomplete passes and dependent passes lacking passed prerequisites normalize to `needs_revision`, preserving work and grades. Normalization occurs in memory; export or a subsequent successful write carries it forward. The maintenance CLI uses the same validation and does not serve HTTP.

The supported deployment is **one process per data file**. There is no cross-process file lock, transactional database, multi-user isolation or conflict detection between stale browser tabs. Two writers or simultaneous native and Docker instances sharing a store are unsupported. The latest successful request replaces that module's latest assessment.

## Assessment contract

States are `not_started`, `in_progress`, `ready_for_assessment`, `passed`, `needs_revision`; the old API name `submitted` is accepted as an alias. Passing requires score ≥80 and ≤100, evidence text, a practical-review confirmation, and all listed prerequisites passed. Returning a passed prerequisite to another status is blocked while a passed immediate dependent exists. Revise descendants in reverse dependency order. This keeps the passed graph consistent without silently invalidating records during ordinary edits.

The practical checkbox is a human attestation to the existing broad competency criteria. There are no per-deliverable mandatory rubrics or evidence inspection in this baseline. Detailed criteria, independent calculation checks, reviewer identities and attempt history remain planned.

## Security and deployment

Native binding defaults to 127.0.0.1. Compose publishes only the host loopback address, and runs the image as UID/GID 10001 with capabilities dropped and no-new-privileges. No application authentication is supplied. An authenticated HTTPS reverse proxy is required before remote use.

API JSON is size-limited and rejects unknown fields/trailing values. Browser API requests with foreign Origin or cross-site fetch metadata are rejected; no CORS is enabled. The UI escapes dynamic text. Submitted paths/URLs remain plain text: the server does not read/download them. CSP limits scripts/assets to self, allows inline styles for the retained UI, and blocks embedding and objects. HTTP header/read/write/idle timeouts and bounded graceful shutdown are configured. A proxy must preserve Host/Origin. These controls do not replace authentication or trusted deployment boundaries.

## Decisions

| Decision | Reason and revisit trigger |
| --- | --- |
| Retain Go + static frontend | The supplied application runs with no third-party runtime dependencies. Revisit a framework only if demonstrated interaction complexity warrants it. |
| Retain JSON | Small, single-user local dataset; easy export and inspection. Revisit SQLite when attempt history/portfolio indexing or multi-writer requirements justify migrations and transactions. |
| Separate content from progress | Curriculum remains reviewed/versioned in Git; personal study records remain local. Never use curriculum `status` fields as learner state. |
| Manual grading with explicit evidence and practical review | Prevents reading-only or score-only completion without pretending to verify engineering correctness. |
| Offline restore instead of HTTP upload | Uses the same validation, keeps arbitrary file handling out of the HTTP surface, and allows a stopped single-writer workflow. |
| All inherited reading references unverified | None came with source evidence or an edition-specific verification record. Remove inherited precise unverified chapter ranges; research them in content PRs. |
| Docker build uses Go 1.27 line and current Alpine 3 tag | Moves from the inherited older images; the same Go line is used in CI. Tags are not digest-pinned, so rebuilds can change; pin digests with a deliberate update process in a follow-up. |

The app's 3,200 hours are inherited planning estimates, not a validated time-to-proficiency claim. The linear prerequisite chain is inherited, not a proven instructional sequence. See the gap analysis and roadmap before extending it.
