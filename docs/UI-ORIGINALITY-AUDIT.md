# UI originality audit

## Stock target planner — 2026-10-10

Before changing the dashboard, inspected authoritative claygorman/bellwright-companion web/components/bw/insights.tsx via GitHub raw API. Upstream Insights stacks equipment/housing/village/recruit panels and a severity-card grid with portrait chips; it has no demand-and-replenishment calculator. Also inspected the active local Steward Insights briefing/details interface to avoid overwriting its concurrent redesign. This isolated branch adds only a new module and a narrow import/render hook; it does not replace that local interface.

New planner composition: explicit planning fields grouped by item, followed by an inline calculation/result strip and source status. No severity cards, portrait chips, recruitment flows or copied upstream interactions. Responsive fields stack on mobile. Inputs have programmatic labels, fieldsets, visible focus and keyboard-operable buttons; status/recommendation changes are announced. Upstream attribution stays intact.

After integration, inspected the resulting Insights screen in a browser at desktop and 390×844. Fields and results are readable without horizontal scrolling. Keyboard Tab advances from the last input to Add item, with visible focus. Verified retained assumptions and fresh stock on reload: synthetic Iron stock3→5 updates shortfall7→5 at target10. Verified null stock remains unknown, absent snapshot shows an empty-data explanation and offline backend shows Retry. No targets are applied or saved. Only browser planning inputs are retained.

Scope of originality acceptance is the new planner section. Existing inherited sections are unchanged in this PR; this audit does not claim completion of the separate full-interface redesign. The current unpublished local Steward interface and its existing audit must be preserved when integrating that work.
