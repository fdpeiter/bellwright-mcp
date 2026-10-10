# V2 team workflow

## Roles

Researcher: decide which player decisions warrant screens, verify actual MCP schemas/current evidence, disclose gaps and source changes. Deliver useful-screen inventory and data gates, not a list of every exposed field.

Designer: turn research into one navigation, source vocabulary, responsive component rules and explicit feature contracts. Own specifications; imagery is exploratory. Require fresh frontend source and no inherited assets.

UX critic: independently challenge complexity, dead ends, map trust, keyboard/mobile comparison, failure recovery and unsupported measurements. Record actionable findings with measurable acceptance before implementation; cross-review implemented work rather than self-approve.

Integrator: create issues/dependencies, own shared files, commit common contracts, dispatch nonoverlapping worktrees, inspect integrated results and attach PRs. Current implementation uses three feature agents plus coordinator. An independent review pass follows feature completion.

## Gates

Research → written design → UX critique → scoped GitHub issues → committed shared foundation → parallel feature edits → cross-review → build/browser checks → PR review. Reuse existing data-service issues. Production measurements need #54 and validated attribution, not a manually entered substitute.

## Current issues

#49 delivery epic; #50 foundation/spec; #51 MCP/Journal/Workspace; #52 location/extractor; #53 workforce/supplies; #54 automatic measurement contract; #55 independent QA. Project https://github.com/users/fdpeiter/projects/1 . Source contracts: V2-SPEC.md; UX gates: V2-UX-REVIEW.md.
