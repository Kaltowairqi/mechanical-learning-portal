# Helicopter Engineering Learning Portal

A single-user, self-hosted helicopter engineering study and manual-assessment tracker. The original Go/static frontend and all **8 stages, 42 modules and 3,200 estimated study hours** are retained. This is a curriculum **outline**: detailed lessons, verified reading assignments and module-specific grading rubrics are planned.

## Start with Docker Compose

Requires Docker Engine/Desktop with Compose v2. From the repository root:

```bash
git clone https://github.com/Kaltowairqi/mechanical-learning-portal.git
cd mechanical-learning-portal
# Before the baseline PR is merged:
git switch verified-baseline
docker compose up --build -d
```

Open [http://localhost:3000](http://localhost:3000). View startup errors with `docker compose logs portal`. Data lives in the named `portal-data` volume, mounted at `/app/data`.

```bash
docker compose restart portal  # preserves data
docker compose down            # preserves data
docker compose up -d           # reuses the same volume
```

`docker compose down -v` deletes the volume and its progress. Run one portal instance per data file/volume. Keep the Compose project/directory name stable; changing it selects a different named volume.

Optional configuration: copy `.env.example` to `.env` and set `PORT=3001` if port 3000 is busy. Compose publishes only on the host's `127.0.0.1`. The app is intentionally unauthenticated for local single-user use; remote access needs an authenticated HTTPS reverse proxy that preserves the original Host and Origin.

## Start without Docker

Requires Go 1.23 or later; verification and CI use Go 1.27.3 / the 1.27 line. No npm installation, frontend build, CDN or third-party Go modules are required.

```bash
go run .
```

Open [http://localhost:3000](http://localhost:3000). Run from the repository root so `static/` and `curriculum/` are found. Stop with Ctrl+C. The server completes active requests during graceful shutdown.

For a reusable binary:

```bash
go build -o portal .
./portal
```

Native defaults: `ADDR=127.0.0.1:3000`, `DATA_FILE=data/progress.json`. Set environment variables in your shell to override them. Go does **not** automatically read `.env`; Compose does. Windows users can use `portal.exe` or Docker Desktop. Personal data and backups are ignored by Git and Docker's build context.

## What works now

- Dashboard: passed-module completion, per-stage progress, logged study time and estimated workload.
- Searchable curriculum, stage/module views, resource guidance and competency summaries.
- Five states: not started, in progress, ready for assessment, passed and needs revision.
- Latest manual assessment score, study notes, submitted-work references, practical-review confirmation, update and completion dates.
- Prerequisite status shown in module details; prerequisite enforcement occurs **when passing**, so browsing and study can happen in any order.
- Dated study sessions, downloadable JSON export, validated offline backup/restore.
- Server-side JSON persistence across restarts. There is no browser-local progress store.

To mark a module **passed**, all four conditions are required:

1. A score from 80 through 100.
2. Nonblank submitted-work/evidence references (paths, URLs or revision IDs; stored as text).
3. Confirmation that mandatory calculations/practical work were reviewed and critical errors resolved.
4. Every listed prerequisite is passed.

Review is manual, including self-review if appropriate. The app records that review; it does not inspect CAD files or validate engineering correctness. Other states allow a blank score. To revise a passed prerequisite, first revise its passed dependents, working backward through the dependency chain. This prevents contradictory completion records.

Legacy ZIP data is supported: `submitted` becomes `ready_for_assessment`; passes missing evidence, practical confirmation or a completion date become `needs_revision` with notes and scores preserved. Descendant passes with unmet prerequisites also require revision. Loading does not rewrite the original file; the normalized state is exported and persisted on the next successful change. Export/back up before upgrading.

## Backup, restore and updates

In **Study log**, choose **Export JSON**, or:

```bash
mkdir -p backups
curl --fail http://localhost:3000/api/export -o backups/helicopter-progress.json
```

The export contains study progress and sessions, not curriculum or external portfolio files. Save the repository revision and back up referenced files separately.

For native restore, stop the running portal and keep a copy of its current data first:

```bash
./portal -check-data backups/helicopter-progress.json
./portal -restore backups/helicopter-progress.json
./portal
```

`-restore` writes to `DATA_FILE` (default `data/progress.json`) only after loading and validating the backup. A missing/malformed backup exits with failure. No HTTP import or arbitrary file upload is exposed.

For Docker, keep backups under `backups/` in the repository root. Stop the service to avoid two writers, then mount that directory read-only into the temporary maintenance container:

```bash
docker compose stop portal
docker compose run --rm -T --no-deps --volume "${PWD}/backups:/backups:ro" portal -check-data /backups/helicopter-progress.json
docker compose run --rm -T --no-deps --volume "${PWD}/backups:/backups:ro" portal -restore /backups/helicopter-progress.json
docker compose up -d
```

The container runs as UID/GID 10001. On Linux, the backup directory/file must be readable by that UID. For example, when the backup contains no sensitive notes, `chmod 755 backups` and `chmod 644 backups/helicopter-progress.json` grant read access; choose narrower group/ACL permissions for private notes. The progress volume itself remains owned by the application user. PowerShell can use `${PWD}` for the directory path as well.

For updates: export a backup, stop the app, pull the intended reviewed revision, then run `docker compose up --build -d`. After this baseline merges, switch to `master` for updates. Re-check `/api/state` and the dashboard. Keep the backup until the update is verified.

## Verification

```bash
go test -race -cover ./...
go vet ./...
gofmt -l *.go
node --check static/app.js
CGO_ENABLED=0 go build -trimpath -o portal .
python3 scripts/smoke.py
# Requires a running Docker daemon; uses a NEW disposable project/volume:
docker compose config --quiet
python3 scripts/smoke.py --compose
```

Node and Python are development/check tools, not application runtime dependencies. The smoke test uses synthetic data, tests real-process restart/export/restore, and removes only its own disposable files/Compose volume. See [the verification report](docs/verification.md) for executed checks and any environment limitations.

## Architecture and project plan

- `main.go`: startup, environment configuration, maintenance CLI and graceful shutdown.
- `handlers.go`: standard-library HTTP API, request validation, security headers and static serving.
- `store.go`: validated state loading, legacy normalization and serialized atomic file replacement.
- `curriculum.go` and `curriculum/helicopter.json`: content types, prerequisite graph validation and curriculum data.
- `static/`: retained responsive HTML/CSS/JavaScript interface; no external font requests.

See [architecture and decisions](docs/architecture.md), [prioritized roadmap](docs/roadmap.md), [curriculum gap analysis](docs/curriculum-gap-analysis.md), [baseline inspection](docs/baseline-inspection.md) and [development conventions](AGENTS.md).

Known limits: one process/writer, latest assessment only (no attempt history), no structured portfolio or file uploads, no automatic engineering grading, no service-worker/offline cache, and all 42 modules' exact resource editions/reading locations still require verification. The app works without Internet while the local server is running; a fresh container build needs base-image access. Completion is educational progress and does not establish aircraft airworthiness.

## API

| Method | Route | Purpose |
| --- | --- | --- |
| GET | `/api/curriculum` | Curriculum with module prerequisites and reference-verification flags |
| GET | `/api/state` | Current progress and study sessions |
| PUT | `/api/progress/{moduleId}` | Replace a module's latest assessment |
| POST | `/api/sessions` | Append a validated study session |
| GET | `/api/export` | Download current state as JSON |

PUT/POST require `Content-Type: application/json`. Unknown fields, trailing JSON, invalid values, cross-origin browser requests and bodies above 64 KiB are rejected. Notes/evidence are limited to 5,000 bytes each, session notes to 2,000 bytes; durations are integer minutes from 1 through 1,440. Dates use `YYYY-MM-DD`; audit/completion timestamps use UTC RFC3339. Invalid input returns 400, unknown routes/modules 404, dependency conflicts 409 and failed persistence 500 (previous memory state retained). See `baseline_test.go` for request examples.
