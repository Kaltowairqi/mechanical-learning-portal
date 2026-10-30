# Baseline verification report

Date: 2026-10-30. Platform: managed Linux amd64. Native compiler/toolchain: official Go 1.27.3 archive, checksum checked against the Go download manifest. Original ZIP preserved under a separate extraction directory during verification. Checks use disposable synthetic data, not learner progress.

## Executed checks

| Check | Result / evidence |
| --- | --- |
| Repository/branches before import | Empty repository; no branches or unrelated content to preserve. Minimal master initialization followed by feature branch. |
| Original ZIP `go test ./...` | PASS — original two tests executed before changes. |
| Original ZIP `go vet ./...` | PASS |
| Original and updated `node --check static/app.js` | PASS |
| Updated `go test -race -cover ./...` | PASS — 10 top-level tests plus table/subtests; 75.7% statement coverage in the final local run (CLI startup is additionally covered by process smoke checks). |
| Updated `go vet ./...` | PASS |
| `CGO_ENABLED=0 go build -trimpath -o portal .` | PASS — actual native binary produced and exercised. |
| Documented `go run .` startup | PASS — actual Go startup serves the dashboard, then shuts down on SIGTERM. |
| `python3 scripts/smoke.py` | PASS — actual HTTP process serves dashboard/assets and 8 stages/42 modules; rejects score-only pass and unmet prerequisite; saves assessment/session; terminates/restarts; state matches; export, offline validation, restore to a new file and restart preserve all records. Malformed/missing backups fail without overwriting current data. |
| `docker compose config --quiet` | PASS — Compose v2.40.3 validates the supplied configuration. |
| Docker image/build/runtime | PASS in GitHub Actions — image built and the actual Compose smoke tested grading gates, restart persistence, invalid-backup rejection, export and offline restore on a disposable named volume. No local Docker daemon was available. |
| Browser functional smoke | PASS in cloud Chrome via the local preview: dashboard, first module form, all five status choices, score-only rejection, synthetic pass with evidence/practical confirmation and completion date, study-session logging with literal `<script>` text, curriculum and keyword search. |
| Narrow viewport / other browsers / native macOS/Windows | NOT RUN — responsive CSS retained, but no cross-platform or mobile-viewport result claimed. |

## Regression coverage

- Score boundaries, five statuses, submitted alias, evidence/practical gates and stable completion date on pass edit.
- Prerequisite pass ordering and prevention of revising a passed prerequisite while a passed dependent remains.
- Unknown modules, invalid dates/durations, unknown fields, trailing JSON, oversized input, foreign Origin and cross-site requests.
- Failed disk writes preserve memory and do not create phantom module records or study sessions.
- Restart/load/export consistency, mode-0600 data files and temporary-file cleanup.
- Legacy normalization without startup disk rewrite, malformed/unknown saved state rejected without modifying its source.
- Twenty concurrent study writes retained and reloaded under the race detector.
- Curriculum totals, duplicate IDs, missing prerequisites, graph cycles and missing resource-verification flags.
- Static/API routes accessible and files outside the static root not exposed through static routing.

## Scope limits

This verifies application behavior, not the correctness/completeness of the 42-module engineering curriculum. Resource edition/section claims are unverified and marked accordingly. Assessment confirmation records a human assertion; evidence is not inspected. The JSON store supports one process, has no historical attempt/audit database and does not claim full power-loss durability. Docker verification was performed on a GitHub-hosted Linux runner; native verification was performed both locally and in CI.

## Final check and CI record

Final local checks pass: Go race/coverage tests, vet, gofmt, JavaScript syntax, static native build, real process smoke and Compose configuration validation.

[GitHub Actions run 41022696325](https://github.com/Kaltowairqi/mechanical-learning-portal/actions/runs/41022696325) tested implementation commit `7ace2a0149593ae28e5bc417f9665dc1af367252`. Both the **go** and **compose** job steps completed successfully. The container log confirms `Progress backup is valid`, `Restored validated progress to /app/data/progress.json`, and the smoke PASS covering dashboard/assets, 8 stages/42 modules, grading, prerequisites, real restart, export and restore. Subsequent documentation/ignore-only updates do not alter the tested implementation.

The implementation is available for review in [PR #1](https://github.com/Kaltowairqi/mechanical-learning-portal/pull/1), targeting `master` from `verified-baseline`. It has not been merged.
