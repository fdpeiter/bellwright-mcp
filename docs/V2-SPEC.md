# Steward v2 specification

Research, design and independent UX review completed 2026-10-10. This specification supersedes v1 presentation rules for v2 only. The existing MCP/parser remain inherited and attributed. New frontend code lives exclusively in v2/ and imports no web/ source, styles, imagery, logos or game assets.

## Useful MVP screens

Plan answers where an extractor might work better; extractor record explains known location, stock and unknown flows. Workforce compares extraction-relevant saved skills/tools/condition without assignment claims. Supplies locates resources in explicit Player containers, with storage record. Journal holds snapshot evidence and automatic observation readiness/results. Workspace exposes source health. Recruitment, expedition, squads and supply graphs are deferred.

## Decisions and evidence

Current MCP data may advance while designing. Discover tools at runtime. Keep saved, runtime, static and estimated sources separate. Runtime suffixes cannot establish save GUID identity: candidates remain unlinked until #39 proves correlation. Combined food-node counts cannot become separate berry/mushroom counts. Containers absent from parser output are not zero/demolished. All-world totals are not player settlement totals.

Each value preserves value, unit, source, capture time, retrieval time and status. Unknown, unsupported, not-measured, stale, unavailable and zero differ. Net stock change is never production. Full balance: opening + extracted + transfers in - transfers out - consumption - losses = closing. Compute reconciliation only when every term known.

## Automatic measurement

No manual quantity/duration/worker-hours forms. Capabilities load automatically on extractor selection. MCP must own capture and recorded results. Currently attributed deliveries and worker game-time are unsupported; do not display a working Start capture or derived efficiency. Snapshot metadata/stock comparisons are separate supported evidence. #54 defines missing capture contract; #44 capability manifest. Worker game-hours and simulation duration differ from wall-clock capture/import time.

## Design system

Canonical Plan / Workforce / Supplies / Journal navigation; Workspace utility. Desktop horizontal links, mobile one native section select. Georgia headings, Segoe UI body, monospace numeric values. Body 16px/1.5; h1 40px/30px mobile. Canvas #F7F5EE, surface white, ink #152D27, muted #52625B, primary #1D4638, accent #7A5A1D, divider #D7DCD4. Spacing 4/8/12/16/24/32/48; radius 8px controls/12px panels; 44px controls; max 1440px, gutters 24px/16px. Flat surfaces, no decoration/fake terrain. One page scroll. Extractor sections Location/Balance/Workforce/Observations use anchors at all widths. No floating selected-item popup or inherited roster composition.

Compare at most two locations in a common frame/units/scale. Mobile retains both summaries, even when selecting the plotted location. Camp triangle, storage square, proposal diamond, resource circle, all labeled. Coordinate/list alternative and keyboard proposal entry. Straight-line distance is an estimate, never a walkable route. Missing baseline keeps directory and coordinates accessible.

## Feature contract and ownership

All feature modules export async render(context) returning HTMLElement. context = { ui, loadTool, navigate, params: URLSearchParams, tools: discovered tool array, refreshTools }. loadTool(name,args?,{refresh?}?) returns { status:'ok'|'stale'|'error', data, source, note, retrievedAt, error? }. Transport /api/tools returns {tools, demo}; POST /api/tools/call accepts {name,arguments} and returns that envelope. data unwraps structured output.data when present, otherwise is complete structured output. No browser raw command arguments; read-only tool allowlist on server. Cached evidence survives failed refresh.

Shared ui exports el(tag,attrs,...children), button(text,action,attrs), link(text,route,params), heading(title,subtitle), notice(text,tone), empty(title,detail,action), field(label,control,hint), table(headers,rows,caption), value(input,fallback), source(result). el creates DOM text safely, never raw HTML; events onClick/onChange etc. Shared classes panel, comparison/two-columns/grid, controls/actions/section-links, field, notice, muted, badge, list, table-wrap. Feature-local styles may live within exclusive feature folder and load via link or style scoped to feature container; do not edit global styles without handoff.

Integrator owns app/router/ui/styles/build/package/docs. #51 owns server.mjs, src/data/**, journal/**, workspace/**. #52 owns plan/**. #53 owns workforce/** and supplies/**. Each uses isolated worktree from foundation commit and records claim. No root configuration or lockfile changes by feature agents. Synthetic fixtures only; no private source data in Git. Build: node build.mjs; start: node server.mjs. No network package installation required.

## Release gates

Actual current upstream screen source/UI individually inspected before and after changes, recorded with hashes/URLs in audit. Exact-source comparison is available through GitHub connector now; earlier historical-only comparisons are provisional. Build and meaningful data/transport tests; desktop/360px, keyboard, 200% zoom, empty/loading/error/stale/missing baseline. No private screenshots. Current game capture/worker instrumentation is a separate service dependency, not a reason to invent UI metrics.
