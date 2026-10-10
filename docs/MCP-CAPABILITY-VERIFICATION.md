# MCP capability implementation and verification

Verified 10 October 2026 against synthetic data. The first batch implements eight
data-backed ideas: gear planning, item history, villager search, inventory
locations, squad inspection, village progression, bounded resources/placements
and independent connection health. The server has 14 tools, preserving its seven
original names and adding seven new names. Quest state, research unlocks,
production queues and durability remain investigations (#21–#24), with no
unsupported gameplay fields advertised.

## Implementation

Three agents used isolated branches and file ownership for #14–#20. Existing #6
resource-query work was reused and reviewed, then integrated in one MCP branch.
The coordinator owns registration, backend API provenance, smoke, CI and docs.
See [the backlog](MCP-CAPABILITY-BACKLOG.md) for all 32 ideas, including twenty
new ideas (#25–#44). All issues are on the Steward delivery board.

Saved query results retain snapshot identity and ingestion time where known.
Health keeps backend availability, telemetry receipt age and live capture
freshness separate. No live query falls back to a saved position. Packaged
resources remain static evidence with unknown availability. Results have hard
limits and counts before/after truncation; squad members are bounded separately.
Item history retains null observations, sampling metadata and save lineage.

Review found and corrected an inherited gear-demand bug: five equipped and five
unequipped villagers previously could produce zero craft shortage. Demand now
includes full desired allocations, retaining higher-tier gear and manual overrides.
Review also found JSON null-to-zero coercion in resource coordinates/placement
metrics, unknown squad membership, and nullable dictionary output-schema issues.
These now fail malformed evidence or preserve unknown state explicitly.

No screen, overlay or interaction layout changed. Web edits are API provenance,
gear calculation and a pure helper; existing Steward UI work remains separate.
The upstream code/game attribution and read-only save policy are retained.

## Verification evidence

- `go test ./...`: passed, including all integrated structured results validated
  against their advertised MCP output schemas and malformed/missing-data regressions.
- `go vet ./...` and binary build: passed.
- `node capability-smoke.mjs`: all 14 tools passed actual stdio calls; invalid
  limits, stale/null evidence, independent outages and EOF were checked.
- `node --test web/lib/gear-planning.test.mjs`: five meaningful gear regressions passed.
- Production Next build and TypeScript checks: passed using `next build web --webpack`.
  Local dependency junctions point outside the worktree, which Turbopack refuses;
  Webpack permits them. CI uses an ordinary isolated dependency installation.
- `node web/scripts/capability-api-smoke.mjs`: production HTTP routes passed
  against a temporary synthetic SQLite DB, checking actual shortages, same-snapshot
  metadata, nullable quantities/playtime, lineage and sampling fields.
- CI now runs the gear regressions, production API smoke and synthetic MCP stdio smoke.

Initial sandboxed tests failed on forbidden loopback sockets; reruns with local
test-server access passed. Successful synthetic tests do not establish fresh
gameplay, current resource availability, measured production or current runtime
ownership. Public issues contain abstract capability scopes and test summaries,
not private saves, logs, paths or extracted assets.

## Activation

Rebuild the web backend to obtain new gear/history metadata, then rebuild the MCP
and reconnect its client process to discover expanded tools. This implementation
does not install mods, alter gameplay, write saves or reconfigure existing volumes.
