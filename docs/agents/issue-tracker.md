# Issue tracker

GitHub Issues in `BenDManning/jetkvm-mcp` are the only execution state. Roadmap
issue #32 owns release-work ordering and dependencies; do not create a parallel
roadmap, lease log, handoff ledger, or agent-governance record.

Keep the active roadmap body limited to the current destination, accepted
specification pointer, live open work and dependencies, owner-only gates, and
completion rule. Reconcile stale state in place after specification acceptance;
GitHub history and comments retain superseded plans and rationale.

## Connect

In each new shell, target and verify the repository before any tracker command:

```sh
export GH_REPO=BenDManning/jetkvm-mcp
gh repo view --json nameWithOwner --jq .nameWithOwner
```

Proceed only when the command prints exactly `BenDManning/jetkvm-mcp`. Read the
complete named issue, its comments, live assignees, and GitHub dependency state
before relying on it. Tracker writes require authority from the current task.

## Execute one ticket

One root implementation owner selects tickets and changes tracker state.
Parallel agents may perform bounded investigation or review, but they do not
claim tickets, edit dependencies, or update roadmap state independently.

1. Read roadmap issue #32 and the candidate ticket from GitHub. Select the first
   open, unassigned ticket in the roadmap's recorded order whose blockers are
   closed.
2. Assign the ticket to the implementation owner, then re-read it. Start only
   when it remains open, assigned to that owner, and unblocked.
3. Work on an isolated branch and keep the issue and pull request linked. Before
   each consequential transition, re-read the live assignment, blockers, branch,
   pull request, and required checks that govern it.
4. After verified delivery, update the ticket and roadmap only when their
   material status, dependency, or order changed. Close an issue only when the
   accepted work is delivered and closure is authorized.

When abandoning work, remove the assignment and leave a concise issue comment
with the branch, commit, last verified command and result, blocker, and next
action. Preserve the repository's existing label vocabulary; label creation,
rename, or deletion requires maintainer direction.
