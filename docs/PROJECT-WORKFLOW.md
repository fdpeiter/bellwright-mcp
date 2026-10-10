# Project board and parallel agent workflow

Use one GitHub Project linked to `fdpeiter/bellwright-mcp`. The project is named **Bellwright Steward** and is linked to this repository. [Open the delivery board](https://github.com/users/fdpeiter/projects/1).

## Board columns

| Status | Entry requirement | Next action |
| --- | --- | --- |
| Backlog | A concrete problem or investigation is recorded. | Clarify scope, dependencies and acceptance criteria. |
| Ready | Acceptance criteria, dependencies and permitted file scope are clear; dependencies are complete. | Claim one issue and assign a working branch/worktree. |
| In progress | One named agent or human owns the issue. | Implement and record meaningful findings or blockers. |
| In review | A reviewable PR is linked and required checks pass, or investigation evidence is ready. | Review the diff and verify acceptance criteria. |
| Done | Acceptance criteria are verified and implementation is merged, or investigation result is recorded. | Close the issue with evidence. |

Use a board view grouped by **Status**. Keep blocked work in its current column and add a `blocked` label with the precise dependency in the issue; do not create a separate state that hides the work's stage.

## Parallel ownership

- Each agent claims one issue before editing and records its name, branch, file scope and dependency assumptions in the issue.
- Use an isolated Git worktree and a `codex/<issue-number>-<short-topic>` branch for each implementation issue. All agents must start from a reviewed common base containing the MCP and Docker source.
- Avoid concurrent edits to shared configuration, lockfiles or the same UI component. Name a single integration owner when scopes overlap.
- Limit **In progress** to the number of available implementation agents. Keep one coordinator available for integration and review; this environment currently supports four concurrent agents including the coordinator.
- Hand off shared-file changes explicitly. An agent may review another agent's change, but must not merge or rewrite it while that agent is still editing.
- Never include game saves, raw telemetry, extracted game assets, `.env` secrets or local credentials in issue attachments or commits.

## Issue and PR expectations

Each implementation issue should include: problem, scope, acceptance criteria, verification method, dependencies, and suggested ownership/file boundaries. Link related issues instead of repeating their scope.

A PR should explain the resulting behavior, relevant checks and remaining limitations, and reference its issue with `Closes #<number>` only when all acceptance criteria are satisfied. Keep partial implementations open and explicitly list unfinished criteria.

For save-derived and Unreal-derived data, preserve source provenance, capture time and stale/unknown states. Packaged resource positions must not be presented as current harvest availability or measured production yield.

## Suggested views

1. **Delivery board:** all open work grouped by Status.
2. **PoC blockers:** issues labeled `area:poc` or `blocked`, sorted by priority.
3. **UI work:** issues labeled `area:ui`, grouped by Status.
4. **Ready for agents:** Status = Ready, sorted by priority and then issue creation time.

## GitHub setup

Project creation requires GitHub CLI authentication with the `project` scope in addition to repository access. The GitHub connector can manage repository issues, but does not expose GitHub Projects operations in this session.

Authenticate with:

```powershell
& 'C:\Program Files\GitHub CLI\gh.exe' auth login --hostname github.com --web --scopes repo,project
```

Reuse a matching existing project before creating a new one. Link this repository, add the scoped issues, configure the five status options, and confirm the board view groups by Status. Do not treat a prepared local workflow document as proof that the remote board exists.

