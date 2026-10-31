# Development conventions

## Purpose and first reads

This repository is exclusively a local-first, self-hosted **helicopter engineering education** portal. Read this file, README.md, docs/architecture.md, docs/roadmap.md and relevant curriculum modules before editing. Keep the distinction between an educational concept design and airworthiness certification explicit.

## Architecture

Go standard library only; root package split into startup/CLI (`main.go`), HTTP (`handlers.go`), persistence (`store.go`) and curriculum types/graph validation (`curriculum.go`). `static/` contains the existing browser UI. `curriculum/helicopter.json` is source-controlled content. Progress is runtime data, not curriculum content, and must stay out of Git and images.

## Responsibilities

- The developer owning a PR is responsible for implementation, relevant regression checks, documentation and an honest verification report.
- Content changes require source/edition verification and explicit measurable passing criteria. If not checked, mark `resourcesVerification: requires_verification`; never invent chapter numbers, ISBNs, quotes or assessment outcomes.
- Keep IDs stable: saved progress and prerequisites refer to module IDs. Validate the graph and estimate totals on every curriculum edit. Do not delete/rename IDs without a migration plan.
- A human records grades and practical review. Software must not represent a score or checked box as independent engineering validation.
- Store portfolio artifacts externally with revision references. Do not add large CAD binaries or learner records to this repository.

## Workflow

Use small feature branches and pull requests. Preserve unrelated work. Do not introduce React, a database, new runtime dependencies or a generalized LMS without a concrete requirement and a recorded decision. Keep Docker Compose runnable. Update docs/roadmap.md when a milestone changes status; distinguish planned and implemented behavior.

Progress writes must stay serialized through Server.mu and Server.save, with memory committed only after disk replacement succeeds. Use one process per data file; offline restore requires stopping other writers. Backward compatibility and failure preservation need tests for persistence changes. Escape all user text before HTML interpolation. Evidence is text; do not dereference arbitrary file paths/URLs on the server.

## Verification

From the repository root:

```bash
gofmt -w *.go
go vet ./...
go test -race -cover ./...
node --check static/app.js
CGO_ENABLED=0 go build -trimpath -o portal .
python3 scripts/smoke.py
docker compose config --quiet
python3 scripts/smoke.py --compose
```

The last smoke command requires Docker and creates/cleans its own disposable project/volume. Never point verification at personal learner data. Use targeted tests for changed behavior, then the listed baseline checks before opening a PR. UI changes also need browser inspection where available; record viewport/platform limits. Do not say a check passed unless executed. CI runs on PRs and pushes to master, with cancel-in-progress concurrency and read-only repository permissions.

Record observed results, blockers, and residual limitations in docs/verification.md and the PR. Do not merge without the owner's instruction.
