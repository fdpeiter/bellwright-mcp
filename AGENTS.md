# Agent working rules

This repository is Bellwright Steward, built on claygorman/bellwright-companion.
Preserve upstream attribution and the read-only save policy.

Use the GitHub project and docs/PROJECT-WORKFLOW.md to coordinate parallel work.
Claim one issue, record files you own, and use an isolated codex/ branch/worktree
once the baseline is committed. Do not edit another active agent's owned files.
If the repository has no initial commit, coordinate local file ownership with
an integrator; do not create competing baseline commits or stage others' work.

Keep saved snapshots, live Unreal captures, static packaged nodes and estimates
separate. Never describe stale/null telemetry as fresh, runtime NPCs as recruited
villagers, or node counts as measured productivity. Keep private saves, logs,
extracted assets, .env files and credentials out of Git and public issues.

Go MCP changes: run tests and vet; verify output schemas through MCP smoke.
UI changes: run the production build and inspect desktop/mobile behavior.
Docker changes: validate compose and verify service endpoints. Preserve the
bellwright compose project/data volume and client identifier during branding.
Move issues to In review with evidence; Done requires acceptance checks and
integrated source. Communicate blockers and dependencies in the issue.
