# Local Docker setup with private inputs

Run on the Windows gaming host with PowerShell 7 and Docker Desktop in Linux-container mode. A clean clone contains source, not saves, a deployed UE4SS probe, terrain images or resource captures. No old workspace path is required. Dashboard/parser attribution to claygorman/bellwright-companion remains intact; Bellwright assets belong to Donkey Crew and are not covered by this repository's MIT license.

## Prepare your own inputs

Copy `.env.example` to `.env` and replace all four directories. Use literal absolute paths, forward slashes and quotes around paths with spaces. Do not use interpolation, inline comments or multiline values for these settings. Process environment overrides `.env` in both Compose and preflight. Keep private files outside the checkout, or under ignored `private/`.

| Setting | Required contents | Origin |
| --- | --- | --- |
| `SAVE_DIR` | Existing directory; `*.sav` in this directory or nested character directories needed for ingestion | Your own saves, normally `%LOCALAPPDATA%/Bellwright/Saved/SaveGames` |
| `UE4SS_LOG_DIR` | `UE4SS.log` | Actual log directory of your UE4SS installation |
| `MAP_ASSETS_DIR` | `exports/Karvenia_08_map_overlay.png`, `exports/Karvenia_08_companion.png`, `karvenia-background.json` | Lawful extraction and calibration |
| `RESOURCE_ANALYSIS_DIR` | `resource-baseline.json`, `foraging-placement-candidates.json` | Private packaged-node baseline and runtime placement analysis |

Compose mounts these directories read-only and requests `create_host_path: false`. Docker Desktop accepted a nonexistent source in the negative test despite that flag, so use the startup wrapper below: it checks host files before any Docker action. Preflight checks presence; service endpoints validate contents. An empty save directory reports `WAITING_FOR_SAVE`, not missing infrastructure. Do not create a dummy log to make checks pass.

```powershell
Copy-Item .env.example .env
# Edit .env, then:
pwsh -File scripts/preflight.ps1 -SkipDocker
docker compose config --quiet
pwsh -File scripts/start-local.ps1
docker compose build mcp
pwsh -File scripts/preflight.ps1
```

Open [the dashboard](http://127.0.0.1:8710). Publisher polls your read-only save directory every 15 seconds. A compatible save is required; startup alone does not establish successful ingestion. Connect using `mcp-go/mcp-config.example.json`, which runs `docker compose run --rm --no-deps -T mcp` per connection. Project `bellwright`, volume `companion-data` and upstream client identifier are preserved.

`start-local.ps1` validates inputs/configuration, builds/starts companion, then recreates live-map and publisher against its current network namespace. It checks endpoints up to three times for startup readiness. Use `-NoBuild` with already-built images. It deliberately leaves companion/data running if a dependent check fails. Unlike read-only preflight, this command changes the deployment and starts save publishing. Run it from the checkout intended for deployment. Missing inputs return exit 1 before Docker is contacted.

## Interpret preflight

The script does not start/stop the game or containers, ingest saves or write inputs. Output contains classifications, never private paths, actor IDs, coordinates or response bodies. Exit 1 means missing/configured inputs; exit 2 means Docker/configuration/service failure; exit 0 means selected infrastructure checks passed. Gameplay readiness is reported separately and does not fail infrastructure setup. `-SkipDocker` leaves container/content/endpoint checks unverified.

- `MISSING_INPUT`: directory/file absent or inaccessible; fix inputs.
- `GAME_INACTIVE`: no Windows Bellwright process detected, independently of file presence. `GAME_RUNNING` proves only process presence.
- `DOCKER_UNAVAILABLE` / `COMPOSE_INVALID`: fix Docker Desktop/configuration.
- `SERVICE_UNAVAILABLE` / `SERVICE_ERROR` / `SERVICE_INVALID_RESPONSE`: connection, HTTP or response-shape failure respectively; inspect logs privately.
- `PROBE_UNAVAILABLE`: endpoint works but has no accepted protocol 0.3 records; check deployment/log location.
- `FEED_STALE`: source timestamp older than 10 seconds; game may be inactive or probe stopped. Log modification time is not freshness evidence.
- `WAITING_FOR_PLAYER`: fresh heartbeat, no loaded player; main menu is not gameplay.
- `PLAYER_UNAVAILABLE`: feed active but no player point fresh within 45 seconds.
- `LIVE_FRESH`: feed/player timestamps fresh; movement/reload identity still require manual checks.

Preflight checks dashboard, companion background, live map, calibration, resource baseline and placements. Live-map/MCP share companion's network namespace; port 18126 stays private and is checked inside companion. When companion is recreated, recreate dependents and reconnect MCP clients:

```powershell
docker compose up -d --force-recreate live-map publisher
```

Do not use `down -v` for repair: it deletes settlement data. Coordinate deployment ownership before running an isolated development worktree against an existing `bellwright` stack.

## Deploy the read-only probe

Docker reads a log; it does not install UE4SS/inject a mod. Install a UE4SS release compatible with your legally obtained Bellwright build using its distribution instructions. Obtain a reviewed read-only probe from the live-integration owner. Place the entry script in the installation's `Mods/<ProbeName>/Scripts/main.lua`; enable `<ProbeName> : 1` in `Mods/mods.txt`, or follow that UE4SS version's documented enablement mechanism. Confirm actual `UE4SS.log` location and set `UE4SS_LOG_DIR`. The user manually launches/loads the game. Disable a failing probe; never mutate saves/gameplay to satisfy telemetry checks.

**Probe source and a pinned compatible UE4SS build are not shipped here. Live production cannot be reproduced from this clone alone.** Record game/UE4SS/probe versions and hashes privately. An arbitrary UE4SS install or old log is not a compatible producer. Require a new session after reload; the previous private probe retained its session, so lifecycle acceptance remains unresolved.

Reader contract (`mcp-go/internal/livemap/server.go`): complete newline-terminated JSON after `BWADVISOR `; an earlier UE4SS prefix is allowed. Envelope: `protocol_version: "0.3"`, `source: "ue4ss_live"`, `event` (`scan_start`, `heartbeat`, `actor`), `data`. Every event needs equal nonempty `scan_id`/`session_id`, nonnegative integer `sequence`, UTC RFC3339 `captured_at` no more than 5 seconds in the future. Actor fields: nonempty `actor_id`, `query_class` (`player` identifies the player), `class_name`, `ownership: "unknown"`, `coordinate_system: "unreal_world"`, `unit: "engine_units_unverified"`, finite x/y/z `position` bounded to ±1e9. Sequence increases per actor. `continuous` is a producer convention, not enforced by the reader. Heartbeat `waiting_for_player: true` and changed sessions clear cached actors.

Synthetic heartbeat shape (not a probe or real capture):

```json
{"protocol_version":"0.3","source":"ue4ss_live","event":"heartbeat","data":{"scan_id":"synthetic-session","session_id":"synthetic-session","sequence":0,"captured_at":"2026-01-01T00:00:00Z","continuous":true,"waiting_for_player":true}}
```

Reader bounds: last 4 MiB of log, complete lines up to 64 KiB, at most 256 actors. Feed becomes stale after 10 seconds; actor points after 45 seconds. Runtime actors do not establish recruitment/ownership. Acceptance needs two fresh changing player captures without saving, menu/exit checks, and a new reload session. Keep raw captures private.

## Lawful extraction

Use only your own legally obtained game/official Modkit data where your license and applicable permissions allow local extraction. Do not bypass encryption/access controls, obtain unauthorized keys/assets or redistribute extracted assets. This setup provides no rights grant or third-party game-asset download.

The included [owner-local extraction pipeline](tools/asset-extract/README.md) builds with locked .NET dependencies, exports the authorized map texture/package, converts BC1 to PNG, prepares the companion image and generates the static resource baseline. Follow its commands with absolute private output paths and compatible lawfully obtained mappings. It supports the observed BC1 map format; unsupported formats fail. It does not acquire mappings or encryption keys. Record game build, package hashes and calibration evidence privately.

Follow [map calibration](docs/MAP-BACKGROUND.md) for the companion's projection. Live-map's `karvenia-background.json` must supply PNG `width`/`height`, lowercase `sha256` matching the overlay image, `source_package`, nonempty `evidence`, and `projection_status`. Use `unverified` until independently calibrated; `verified` requires a finite nondegenerate six-number affine `matrix`. Do not copy the companion's matrix without verifying its coordinate frame. See `mcp-go/internal/livemap/background.go` and tests for exact validation.

Consult `mcp-go/internal/livemap/resources.go` and tests for baseline/placement validation. Baseline has `schema: "bellwright.harvestable-baseline.v1"`, `source_kind: "packaged_game_asset"`, nodes with unique `id`, `baseline_category`, `class_asset_path`, three-number `position`, and nonempty `unknowns`. Placements are an array of `bellwright.foraging-placement-candidate.v1` entries with runtime target identity, assumed radius, current/practical sites and assumptions; `[]` truthfully means no placement evidence. Never fabricate identities or label packaged nodes as measured harvest availability. The included extractor preserves observed component translations and explicitly leaves parent transforms, streamed-world coverage, tree/ore attackables and runtime availability unverified.

Keep saves/logs/databases/mappings/packages and extracted images/JSON outside Git and Docker build contexts. Ignored `private/`, `map-assets/`, `resource-analysis/`, `saves/`, `logs/` are local-data conventions; never force-add inputs. Review staged filenames before committing. `.env.example` contains placeholders only. Synthetic checks cannot prove actual extraction, producer compatibility or a fresh gameplay session.

## Setup regression checks

```powershell
pwsh -File scripts/preflight.tests.ps1
pwsh -File scripts/start-local.tests.ps1
node --test scripts/check-services.test.mjs
pwsh -File scripts/preflight-compose.tests.ps1
```

The last check needs Docker CLI/Compose. It resolves an isolated synthetic owner configuration and verifies missing inputs stop the startup gate without deploying containers. Node tests verify classifications and that private bodies/errors are suppressed; PowerShell fixtures test presence/path parsing, not real asset content.

For an isolated deployment, use `-ProjectName owner-test` with both scripts and set `COMPANION_PORT=0` in its private env file for a random loopback port. Pass that same project and env file to every Compose command and MCP connection. The default project remains `bellwright`. The publisher recursively discovers character saves. Only remove volumes belonging to a disposable test project; preserve the normal settlement volume.
