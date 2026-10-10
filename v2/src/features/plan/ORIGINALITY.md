# Plan feature originality and verification (#52)

Independent v2 implementation, 2026-10-10. No inherited frontend imports, components, stylesheet, imagery, map texture, logos or game assets. Existing MCP/parser provenance remains inherited.

## Current upstream inspection before and after

Inspected upstream main via GitHub connector before implementation:

- [Map screen](https://github.com/claygorman/bellwright-companion/blob/main/web/components/bw/map-tab.tsx), Git blob `6bc139dfb3c0cf1c3b9932b134b6e0155da9efb8`.
- Selected building overlay is the `selPin` block in that same file, individually inspected: floating top-right panel, icon hero, source badge, description, coordinates and close control.
- Planner interaction is the building-type/radius selector and draggable ring in the same file, individually inspected.

The upstream map uses many grouped layer switches in a sidebar, imagery-based pan/zoom canvas, wiki/game icons, planner ring and floating selected-object popup. The v2 Plan page starts with resource and stable saved extractor selection, then persistent Location/Balance/Workforce/Observations sections. It uses an original equal-scale XY schematic, a labeled coordinate-list alternative, a native keyboard coordinate proposal form and paired evidence summaries. No upstream compositions or interaction sequences were copied. The planner creates a local URL proposal, not a draggable game-map ring. Extractor detail is part of the persistent document, not an overlay.

After implementation, compared independently authored feature structure against those source sections; no layer drawer, hover-tooltip machinery, floating building panel, game imagery, planner ring, profile action or upstream projected-pin system is present. Rechecked upstream file identity after implementation. This is a current-source composition comparison; it is not a claim of browser screenshots or upstream runtime inspection. Coordinator must add integrated desktop/mobile/keyboard evidence to the repository originality audit before acceptance.

## Data and behavior verification

- Exactly `ForagingCamp_C` + `Player` faction + stable saved ID qualify; producers, hostile or unclassified containers do not enter the extractor directory.
- Missing stock remains unknown, known zero stays zero. Missing containers do not imply demolition.
- Tool metadata decides whether bounded arguments are supported. Bounded `items` payload and legacy/demo `nodes` payload are handled. Plot samples are capped and disclosed.
- Static node categories do not establish berry/mushroom item yields. No coverage score, measured output, route, buildability or automatic efficiency is fabricated.
- Runtime placement candidates remain separate unlinked evidence. No suffix, name or proximity join to selected saved camp.
- Current and local proposal share XY scale and plot frame. Coordinates use engine units without an asserted metres conversion. The storage estimate is straight-line XY distance, never a route.
- Full balance includes consumption, losses and reconciliation; flow terms remain unmeasured, and no residual is computed from incomplete evidence.
- Proposal coordinate inputs are plans, not manually entered measurements. Production observations link to automatic MCP readiness; no capture is invented.

Commands: `node --test v2/src/features/plan/model.test.mjs` (8 tests passing); `node v2/build.mjs` (passing). DOM contract checks cover player scope, stock unknown/zero, coordinates, equal-scale comparison, supported/legacy MCP query arguments and meaningful empty states. Full browser acceptance remains for coordinator integration: 360px, desktop, 200% zoom, keyboard proposals and anchors, stale/failed sources, source drawer, and no page overflow.
