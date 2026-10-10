# Issue #7 implementation evidence

Ownership: this session implements the existing infrastructure claim for #7 on `codex/7-private-input-setup`, isolated from baseline `5d1b607`. Files owned: Compose, environment example, Docker setup guide, preflight/scripts, private-data ignore rules, README setup link and this evidence document. No UI or game source was edited; unrelated active-repository edits remain untouched. The committed baseline exists locally; #1 remains open remotely. No issue messages, board transitions or deployment mutations were made from this session.

Checked 10 October 2026:

- PowerShell syntax parse passed. `pwsh -NoProfile -File scripts/preflight.tests.ps1` passed synthetic absent-env, missing-log, complete quoted paths with spaces, empty save directory and process-environment precedence cases. Temporary synthetic fixtures were removed.
- Clean worktree without `.env` produced `MISSING_INPUT`, separate `GAME_INACTIVE`, and exit 1 with Docker skipped. Process state is an observation at check time, not proof of a loaded session.
- `docker compose --env-file <owner-local-env> config --quiet` passed. Existing project `bellwright`, volume identity, loopback binding and shared network design are preserved. Bind mounts now refuse host-directory creation.
- Read-only full preflight using the owner's external configuration found required private files and an empty configured save directory. It independently detected a game process. Companion and prepared companion map endpoints passed. Live-map/calibration/resource/placement endpoint requests inside companion failed to connect; shared-network attachment or service availability requires deployment-owner repair. No containers were recreated and no saves ingested.
- `git diff --check` passed. `git check-ignore` verified private capture JSON, terrain PNG in the documented private directory, resource baseline, saves, UE4SS log, package/mapping files and secret environment variants. Tracked filenames contain no save/log/package/mapping inputs. No private input contents or paths are recorded here.

Unverified: deployment from an empty owner's machine with legally extracted valid assets, live producer installation/compatibility, actual current save ingestion, live endpoint contents, fresh movement/menu/exit captures and reload session identity. Probe source, texture exporter and full resource extractor are absent from the baseline; documentation makes these prerequisites explicit. This focused change diagnoses infrastructure and documents reproducible configuration, but does not satisfy complete end-to-end clean-owner acceptance. Keep #7 open until those prerequisites and checks are addressed. Go source and UI are unchanged, so Go/UI checks are outside this change.

## Continuation and superseding results

Under the user's instruction to continue, inspected deployment timing: live-map predated companion's latest restart. Recreated only live-map using the active repository's existing Compose configuration, with `--no-deps --force-recreate`. Companion, publisher and data volume were not recreated. This restored all six endpoint checks. Current protocol records yielded fresh feed/player timestamps, waiting=false and invalid_records=0. This supersedes the earlier unreachable live endpoint result; it does not establish two moving captures or reload identity.

The isolated negative Docker test exposed a platform limitation: Docker Desktop accepted a missing bind source despite resolved `create_host_path: false`. Its disposable project/volume was removed. Added `start-local.ps1` to reject missing private inputs on the host before any Docker command and to attach dependents after companion updates. Direct Compose startup alone is not the documented safety gate.

Checks passed after continuation:

- PowerShell fixture tests, now including rejection of Windows drive-relative paths.
- Three Node tests covering absent probe, stale feed, menu heartbeat, stale/missing/fresh player, malformed response, HTTP/connection failures, and suppression of private response/error content. Gameplay inactivity remains distinct from infrastructure failure.
- Compose configuration with independent synthetic quoted owner paths; five private mounts retained read-only/create-host-path flags. Startup wrapper rejected missing paths with exit 1 without creating directories/deploying services.
- Full read-only preflight against the existing owner deployment: dashboard, prepared image, live feed, calibration, baseline and placements all passed. Configured save directory was empty, so new ingestion remains unverified.

Remaining acceptance limitations: full successful startup-wrapper deployment on a new owner's machine, actual compatible save ingestion, lawful asset extraction and producer installation from independently obtained inputs, movement/menu/exit and reload-session lifecycle. Only the failure gate and existing-owner endpoints have been exercised; the complete build/start path of the new wrapper has not been run against the shared deployment. No private files/assets were added to Git.
