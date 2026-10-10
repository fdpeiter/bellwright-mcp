# Proof-of-concept status

Audited 10 October 2026. This document separates repository evidence, previously reported runtime checks, and work still required. No game process or runtime configuration was changed by this audit.

## Current evidence

| Capability | Evidence | Status |
|---|---|---|
| Save ingestion and dashboard | `publisher/src/publisher.ts`, `compose.yaml`, Rust dump/parser pipeline; latest save ingestion previously reported successful | Implemented; reported working locally |
| Docker deployment | Companion, publisher, live-map services and opt-in stdio MCP service in `compose.yaml`; bind mounts read-only | Implemented; engine and containers previously reported working |
| Read-only MCP | `mcp-go/main.go` registers seven fixed-endpoint tools, rejects non-loopback origins/redirects, caps responses, reserves stdout for protocol | Implemented |
| MCP integration smoke | `mcp-go/smoke.mjs` initializes, lists and calls every tool; all seven previously reported passing | Transport/endpoints checked; not proof of fresh game data |
| Save/live separation | MCP tools label save, live actors, packaged resources and inferred placement independently | Implemented |
| Live UE4SS positions | Go live-map consumes UE4SS log via read-only mount; previous no-save movement test succeeded | Previously demonstrated; latest reported feed stale |
| Real map/resource overlay | Live-map supports configured image, calibration, resource and placement files | Implemented; inputs are private external artifacts |
| Assistant client session | Docker stdio example exists in `mcp-go/mcp-config.example.json` | Verified in a new Codex session: native snapshot and live-map tool calls succeeded; live freshness remained stale |
| Companion realtime integration | Web UI polls `/api/telemetry/latest`; UE4SS reader independently serves `/api/map` | Two separate live pipelines; bridge missing |
| Resource availability/output | Packaged baseline nodes and foraging coverage endpoint | Depletion, regrowth, seasonal availability, yield and output per hour unverified |

The seven tools succeeding means their HTTP responses were valid. For example, telemetry can legitimately return `data: null`, and the live map can legitimately report stale data. Neither response establishes an active game connection.

## Priority issues and acceptance criteria

### P0 — Connect the MCP to a new assistant session

Owner: MCP/client setup. Files: client configuration, `mcp-go/README.md`, integration checklist.

Use Docker stdio with `compose --project-directory <repo> run --rm --no-deps -T mcp`. The `-T` flag keeps the protocol free of terminal control characters. Register it in the actual client, start a new session, discover all seven tools, and call saved-world plus live-map tools. Record freshness honestly and keep session secrets/private save contents out of screenshots and issues. Acceptance: an actual new assistant session receives a structured result from this server.

### P0 — Repeat a fresh no-save live movement test through Docker and MCP

Owner: UE4SS/live integration. Files: `mcp-go/internal/livemap/`, private local probe deployment, test checklist.

Load the game and move without saving. Compare at least two captured timestamps/player positions through `read_live_map`. Stop receiving records (menu/exit) and verify stale indication; reload and verify new-session handling. Acceptance: fresh changing coordinates via MCP, then correct stale state without substituting save coordinates. Host mod installation and log location must be documented; Docker cannot itself inject UE4SS into the Windows game.

### P1 — Feed the companion realtime map from UE4SS

Owner: Go telemetry adapter. Files: new isolated Go bridge command; avoid concurrent editing of live-map UI.

Map the bounded actor stream to the companion telemetry schema and POST to its fixed local endpoint. Preserve capture time, age, session and ownership uncertainty: forwarding an old snapshot must not make it appear fresh merely because the receiver stamps a new `received_at`. Runtime NPCs cannot all be labeled recruited villagers. Add explicit building support or a documented fallback kind because the current `LiveActor.kind` union lacks buildings. Acceptance: companion realtime map shows fresh player movement without save ingestion, retains provenance, and visibly stops being live when the source is stale.

### P1 — Build a reproducible private-input setup

Owner: infrastructure/docs. Files: `compose.yaml`, `.env.example`, deployment scripts/docs.

The current live map needs external map-image/calibration/resource/placement paths and a host UE4SS log directory. Add a preflight that names missing inputs and distinguishes a missing asset from an inactive game. Document supported probe version/schema, acquisition/extraction steps, and start/stop/update instructions. Bind mounts remain read-only. Acceptance: another authorized game owner can configure the stack from documented steps without our old workspace; no game assets or saves committed.

### P1 — Make MCP resource tools useful at assistant scale

Owner: MCP API. Files: `mcp-go/main.go`, tool schemas/tests.

Current tools have empty argument schemas and fetch whole responses up to 32 MiB. Baseline node output is large. Add bounded summaries and filters (resource category, world bounds/radius, limit), plus compact placement details and a health/provenance summary. Acceptance: a normal assistant can ask for nearby resources and receive a bounded useful result; limits cannot silently imply complete coverage.

### P1 — Turn placement coverage into a verified first structure experiment

Owner: resource analysis. Files: placement analysis and docs, not UI shell.

Foraging is the first trial. Confirm resource classification, work-radius units, current camp identity, buildability and hauling route; compare a controlled work period before/after with similar workers and food targets. Baseline node count alone is not maximum resource output. Acceptance: a documented recommendation with assumptions, measured outcome or explicit remaining uncertainty, and no automatic game writes.

### P1 — Clarify UI freshness and source labels

Owner: UI. Files: `web/components/bw/` and independently `mcp-go/internal/livemap/index.html` in separate issues.

Replace misleading “live from your save” wording with save capture/ingest time. Show save-backed, genuinely live, stale and packaged-static sources distinctly. Provide an integration status view and link between the companion map and UE4SS map while the bridge is incomplete. Acceptance: stale log output is never represented as current game state; desktop/mobile navigation remains usable.

### P2 — Expand resources and settlement identity

Owner: game-data research. Files: isolated extraction/analysis tools and provenance docs.

Current baseline does not establish complete tree/ore coverage or current availability. Discover those classes separately, distinguish loaded-world census from packaged instances, and link runtime actors to saved building/villager IDs only with evidence. Acceptance: coverage tables state exactly which categories and world scope are known; unknown ownership persists until verified.

### P2 — Add CI and failure-path checks for our additions

Owner: tests/CI. Files: `.github/workflows/`, `mcp-go/*_test.go`, bounded integration fixtures.

Existing Go tests cover origin restrictions, a valid read and redirects. Add non-200/invalid/oversize JSON, timeout, stale/no-player/session transitions as appropriate, and Docker/MCP startup validation with synthetic non-personal fixtures. Acceptance: Go test/vet/build joins the existing web checks and failures cannot silently turn into successful fresh responses.

## Parallel agent boundaries

- Board/triage owns project fields, priorities, dependencies and issue assignment; it does not edit application code.
- Client setup owns client registration and session proof; coordinate before editing `mcp-go/main.go`.
- MCP/API agent owns tool schemas and server tests.
- Telemetry adapter agent owns a new bridge command and contract documentation.
- Companion UI and Go live-map UI are independent assignments; do not assign both agents the same file.
- Resource researcher owns classification and placement evidence, with private output files.
- Infrastructure owns Compose/setup/preflight; coordinate port/network changes with MCP and bridge owners.
- Naming agent proposes a name first, then performs the approved mechanical rename separately from functional changes.

Use one issue/branch (preferably isolated worktree) per assignment, name acceptance criteria and touched files before starting, and link dependent issues. A completed smoke test does not close the fresh-live or client-session issues.

## PoC completion gate

The usable read-only PoC is complete when the new assistant session can query an actual save, query fresh no-save movement, identify stale/disconnected state, obtain bounded nearby resource information, and explain which recommendation inputs remain unverified. Construction, hiring, priority changes and arbitrary Unreal commands are outside the current read-only scope.

