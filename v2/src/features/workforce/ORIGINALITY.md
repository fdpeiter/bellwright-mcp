# Issue 53 workforce originality and verification

Before implementation, individually fetched and inspected current upstream screen source through GitHub on 2026-10-10. Repository tree: `0a47289648d6a3584cae12dc036b7566bbc98e25`.

- Roster: [current source](https://github.com/claygorman/bellwright-companion/blob/main/web/components/bw/roster.tsx), blob `d8785707873d559044bd4e0694c63bf1b23b7560`. Upstream composition is a dense all-skills/gear table, clickable rows, sorting headers and mobile summary cards; compare-mode selection is embedded in rows.
- Person overlay: [current source](https://github.com/claygorman/bellwright-companion/blob/main/web/components/bw/drawer.tsx), blob `58ef37703c41590d41f5b9deb8909157271d5f45`. Upstream uses a fixed side drawer over a backdrop, avatar header, combat/work skill bars, squad avatars, equipment slot tiles and job rows.

After implementation, compared new feature source against each inspected screen. New v2 flow starts with extractor context and an explicitly chosen saved skill; candidate records expose only relevant evidence, then a labeled button opens inline contextual person evidence. No global all-skills roster, selected-row composition, portrait, side drawer, equipment paper doll, overlays or inherited interaction flow. Native controls and named inspect buttons replace clickable table rows. Skill-cap values remain evidence rather than an efficiency score. Local session staffing proposals explicitly make no game writes and are unavailable without a stable saved person/extractor identity.

No frontend source, styles, assets, logo, components or helper imports from upstream/web are used. Existing MCP/parser provenance remains inherited and attributed by the shared shell. UI uses coordinator-owned fresh DOM helpers only. This file is feature evidence for the coordinator to incorporate into `docs/UI-ORIGINALITY-AUDIT.md`; it is not a substitute for the integrated audit or visual acceptance gate.

Validation: feature-local Node tests cover explicit recruitment, missing/null versus zero, sorting unknown values last, extractor class/Player scope and local proposal constraints; production asset build succeeds. Integrated browser validation remains required for desktop/360px, 200% zoom, keyboard, stale/error/empty states, focus and storage failure.
