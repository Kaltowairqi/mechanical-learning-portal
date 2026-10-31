# Baseline inspection and ZIP comparison

Inspected 2026-10-30. Source: the user's attached `mechanical-learning-portal(1).zip` and `Kaltowairqi/mechanical-learning-portal`.

## Repository state before work

The GitHub repository had size zero, no branches, and the contents API returned `This repository is empty.` No tracked application, agent conventions, open branch content or unrelated repository files existed to reconcile. A minimal README initialization commit created `master` (`cb0a90c711fb7df913586c4470401e47c5d3e5a1`) so the implementation could be proposed against a real base on `verified-baseline`.

## Supplied archive

The ZIP was inspected before extraction; paths and symbolic links were checked to prevent extraction outside a dedicated directory. It contained 11 nonempty files plus `data/.gitkeep` and empty placeholder directories. There were no third-party Go dependencies, secret/environment files or binary CAD artifacts. The original ZIP was preserved unchanged.

| ZIP path/group | Original finding | Baseline action |
| --- | --- | --- |
| `go.mod` | Go 1.23, no dependencies | Retained |
| `main.go` | One file for content types, handlers, persistence and startup | Retained behavior, split responsibilities into four small Go files; added validation and graceful HTTP shutdown |
| `main_test.go` | Two tests for basic pass/prerequisite/file-write and study-session validation | Executed original tests/vet; retained and updated for practical evidence/JSON contract; added regression coverage |
| `curriculum/helicopter.json` | 8 stages / 42 modules / 3,200 hours; title, resources, competency paragraph, prerequisites and same assessment boilerplate per module | Preserved IDs/order/hours/competencies; removed unused static learner status; labeled all resources unverified; removed inherited precise unverified chapter ranges |
| `static/index.html`, `app.js`, `style.css` | Working dashboard, curriculum/stage/module pages, sessions/export/resources; four statuses; minified code and decorative external fonts | Preserved visual direction and views; readable JS, five states, evidence/practical confirmation/completion date, prerequisite status, keyboard stage buttons, escaped text, no external font requests |
| `Dockerfile`, `compose.yaml` | Nonroot app, loopback publication and named volume; older base images; Docker built only `main.go` | Build whole Go package after splitting; current Go line, retained nonroot/volume; configurable host port and limited container capabilities |
| `.gitignore`, `README.md` | Minimal ignore/setup guidance; broad claims about curriculum completeness | Expanded data/config/build ignores; explicit tested commands, backup/restore, assessment/migration rules and known limits |
| Empty `web/src/`, `backend/` | No implementation | Removed empty placeholders only |

## Verified strengths and corrected gaps

The original code compiled and its two tests and `go vet` passed on Go 1.27.3. It already escaped user text, enforced ≥80 and prerequisites when passing, validated session dates/durations, serialized state and exposed JSON export. It was a viable baseline and did not warrant a stack replacement.

Corrections address score-only passing, missing revision/readiness state, missing completion dates/evidence fields, a possible phantom map entry on failed progress writes, inconsistent passed prerequisites after demotion, no curriculum graph validation, overly broad startup state acceptance, predictable temporary-file naming, missing explicit file sync, missing server timeouts and incomplete backup/restore documentation. The retained UI is now connected to those contracts.

The archive had no structured learning-objective lists, reading-verification records, exercise sets, per-module rubrics, reviewer/attempt history or portfolio model. Broad `criteria` paragraphs are useful starting points, not complete teaching material. Detailed educational content is deliberately deferred; see the gap analysis and roadmap.
