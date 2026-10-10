# Steward v2 originality and verification audit

The v2 frontend is independently authored under v2/. It imports no inherited web/ frontend source or assets. Existing MCP/parser attribution remains retained. Historical v1 audit material in the active checkout is outside this isolated branch and must be preserved at integration.

## Exact upstream comparison evidence

Current upstream source inspected through GitHub connector on 2026-10-10; comparison only, no upstream source copied into v2.

| Screen | Source and blob | Upstream structure | V2 replacement | Verification |
| --- | --- | --- | --- | --- |
| Shared shell | claygorman/bellwright-companion web/components/bw/app.tsx, fd5bf5cebc19b25887e9cf370c0e4f4761cd6776 | Metadata masthead, category tab strip and conditional toolbar, fixed scrolling body, person drawer and compare tray | Task links Plan/Workforce/Supplies/Journal, independent text mark, source context, normal page scroll, mobile native section selector | Before-source comparison completed; production assets build. Integrated browser/mobile/keyboard and after-comparison pending |
| Connection and preferences | app.tsx above; upload-modal.tsx, 8a03a355c2128f50e2a1e553d509086d74cb3435 | Header toggle dropdown and modal upload/drop-zone/ingest flow | Read-only source health full page; no upload or game-write flow, tool discovery/capability explanation | Before-source comparison complete; feature implementation and after-review pending |

Feature owners record map/extractor, workforce/person, supplies/storage and journal/history comparisons individually in their exclusive feature docs; coordinator integrates entries after review. No global claim substitutes for individual comparison.

## Release checks

Concept rasters are exploratory and cannot prove accessibility. Required checks: production build, desktop/360px, 200% zoom, keyboard and focus, useful loading/empty/error/stale/missing-baseline states, source capture versus retrieval time, unchanged saved evidence on runtime outage, explicit unknown values and unsupported measurement. All public test fixtures/screenshots must be synthetic. Do not mark features Done until integrated and verified.
