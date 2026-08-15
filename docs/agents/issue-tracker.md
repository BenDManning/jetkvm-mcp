# Issue tracker

Issues and specs live in GitHub Issues.

- Repository: `BenDManning/jetkvm-mcp`
- CLI: `gh`
- Pull requests as a triage request surface: no

Every tracker command must target the configured repository. In each new shell, run:

```sh
export GH_REPO=BenDManning/jetkvm-mcp
gh repo view --json nameWithOwner --jq .nameWithOwner
```

Proceed only when the verification command prints exactly `BenDManning/jetkvm-mcp`. The export scopes bare `gh issue` commands and the `{owner}` / `{repo}` placeholders used by `gh api` below.

Read complete tickets in non-interactive sessions with `gh issue view <number> --json number,title,body,state,stateReason,author,labels,assignees,milestone,createdAt,updatedAt,closedAt,url,parent,blockedBy,blocking,comments`. Create or change issues only when the current task explicitly authorizes that tracker write.

Preserve the existing label vocabulary. Do not create, rename, or delete GitHub labels without maintainer approval.

## Triage discovery

Before placing an unlabeled issue in a triage discovery bucket, fetch its body and exclude it when the first nonblank line is `<!-- wayfinder:map -->` or matches `<!-- wayfinder:type=<type> -->`. These are Wayfinder artifacts even when no `wayfinder:*` label exists. This filter is discovery-only; an explicitly named issue remains addressable.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a single issue with **child** issues as tickets.

- **Labels**: before creating issues, list existing names with `gh label list --limit 1000 --json name --jq '.[].name'`. Apply `wayfinder:map` or `wayfinder:<type>` only when that exact label exists. When it does not and label creation is not explicitly authorized, omit `--label`; make `<!-- wayfinder:map -->` or `<!-- wayfinder:type=<type> -->` the body's first nonblank line. Never create labels implicitly.
- **Map**: a single issue holding the Notes / Decisions-so-far / Fog body. Create it with `gh issue create`, adding `--label wayfinder:map` only when the label check permits it. Before charting completes, record exactly one child backend marker (`<!-- wayfinder-child-backend:native -->` or `<!-- wayfinder-child-backend:task-list -->`) and one blocker backend marker (`<!-- wayfinder-blocker-backend:native -->` or `<!-- wayfinder-blocker-backend:body -->`) in the map body.
- **Child ticket**: create the issue without adding it to the recorded child source. For `task-list`, put `Part of #<map>` immediately after the fallback type marker when present, otherwise at the top of the body. Apply the matching `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`) only when the label check permits it. Once published and claimed, the ticket is assigned to the driving dev.
- **Publish a child**: create all new ticket issues first so every blocker has an id. Then publish each ticket in map order under the relationship-write leases below. With the leases held, write and verify the ticket's complete blocker set before linking it as a native sub-issue or adding it to the fallback task list. Re-read the selected child source and complete blocker set before releasing the leases. An unpublished ticket is never a frontier candidate.
- **Blocking**: follow the recorded blocker backend. For `native`, add GitHub issue dependencies with `gh api --method POST repos/{owner}/{repo}/issues/<child>/dependencies/blocked_by -F issue_id=<blocker-db-id>`, where `<blocker-db-id>` is the blocker's numeric database `id` from `gh api repos/{owner}/{repo}/issues/<n> --jq .id`. For `body`, put `Blocked by: #<n>, #<n>` after any Wayfinder marker and `Part of` line, before the question. A ticket is unblocked when every recorded blocker is closed.
- **Child source**: read only the recorded child backend. For `native`, page through `gh api --paginate --slurp 'repos/{owner}/{repo}/issues/<map>/sub_issues?per_page=100' --jq 'add | .[].number'`; the returned order is map order. For `task-list`, parse that map's task-list children from `gh issue view <map> --json body --jq .body` in body order. Never substitute repository-wide issue order.
- **Blocker source**: read only the recorded blocker backend, independently of the child backend. For `native`, enumerate every edge with `gh api --paginate --slurp 'repos/{owner}/{repo}/issues/<child>/dependencies/blocked_by?per_page=100' --jq 'add | .[] | [.number, .state] | @tsv'`; retain open and closed relationships, applying the state filter only when calculating the frontier. For `body`, parse every `Blocked by:` reference and fetch each blocker's live state with a direct `gh issue view`.
- **Backend integrity**: a missing or duplicate marker, malformed fallback representation, unavailable selected backend, nonzero tracker command, or invalid/null output fails closed: do not choose, claim, or resolve a ticket. Use direct `gh ... --jq` commands as above and check their status before interpreting output; do not let a downstream filter mask a `gh` failure. Never switch representations because another backend becomes available or unavailable. A backend migration requires explicit maintainer authorization and the relationship-write leases below for every affected child. Proceed only while every lease remains active and no affected open child is assigned; copy and verify every child and blocker relationship, including closed dependencies, in the new representation before changing its marker.
- **Frontier**: preserve child-source order and scan at the lowest available resolution. Fetch only `number,state,assignees` for each child with `gh issue view <child> --json number,state,assignees`; use the selected blocker source, reading only `body` when the body backend must expose its references and only `number,state` for referenced blockers. Discard closed, assigned, or blocked children. The first remaining child is the frontier. After its exclusive claim succeeds, fetch that one ticket with the complete-ticket command before work; never load every child's comments or unrelated body during a frontier scan.
- **Claim**: use the exclusive ticket-state lease below. Claims and blocker rewrites share this per-child lease; assignment alone is not an exclusive claim.
- **Resolve**: use the safe resolution sequence below.

### Exclusive comment leases

Use issue comments as ordered, expiring lease attempts. Give each attempt a collision-resistant token and one kind: `ticket-state` on a child or `map-edit` on the map. Post `<!-- wayfinder-lease:<kind>:<token> -->` with `gh issue comment <issue> --body-file -` and a quoted heredoc.

Read all comments with `gh api --paginate --slurp 'repos/{owner}/{repo}/issues/<issue>/comments?per_page=100' --jq 'add | sort_by(.id)'`. Count markers only from comments whose REST `author_association` is `OWNER`, `MEMBER`, or `COLLABORATOR`; a release matches only the same kind, token, and author login as its lease. A lease attempt is active for 10 minutes from its REST `created_at` when no later `<!-- wayfinder-release:<kind>:<token> -->` comment exists. Among active attempts of the same kind, the lowest numeric REST comment `id` is the sole winner. Wait until the posted token appears in the read; only the winner may proceed. A loser posts its matching release marker and selects another ticket or retries after the winner releases. An expired attempt never wins. A failed or malformed comment read fails the lease attempt closed.

A **relationship write** is a child publication or removal, blocker-set change, or backend migration. Acquire the map's `map-edit` lease first, then every affected child's `ticket-state` lease in ascending issue-number order. Hold the complete set through all writes and read-back verification. If every lease cannot be held concurrently before expiry, release them and stop before writing; if one expires, reacquire the complete set and fresh snapshots before another write. This lock order serializes migrations with every relationship writer. Claims acquire only their child's `ticket-state` lease.

For a claim, the `ticket-state` winner re-reads the ticket and its blockers, confirms it is still open, unassigned, and unblocked, then runs `gh issue edit <n> --add-assignee @me`. Re-read once more and start work only when the ticket is assigned to the claimant and the claimant's token remains the winning active lease. Then release the lease; the assignment remains the durable claim. Otherwise remove the assignment if added, post the release marker, and stop. When abandoning work, remove the assignment.

Every blocker addition, removal, or replacement on an existing child is a relationship write and must use both the map's `map-edit` lease and that child's `ticket-state` lease. The winner re-reads the ticket and proceeds only while it is open and unassigned; never rewire an assigned child. Apply the change through the map's recorded blocker backend, then re-read and verify the complete blocker set before releasing the leases. If the ticket is assigned or another attempt wins, release any losing attempt and stop without changing blockers. Holding the child lease from the claimant's final blocker check through assignment prevents a blocker rewrite from interleaving with a claim.

When an operation changes the map body, the `map-edit` winner saves the exact latest body after acquiring the lease, derives the intended transformation from that snapshot, and writes the complete merged body with `gh issue edit <map> --body-file -`. Re-read the map and require its body to equal the proposed merged body. Compare the pre-edit snapshot with the verified result and account for every changed line: all content not intentionally edited must remain byte-for-byte unchanged, including Destination, Notes, every existing decision pointer, Not yet specified, Out of scope, backend markers, fallback task lists, and unrecognized sections. Any unexplained change fails verification; repair from the snapshot and verify again before posting the release marker. Finish within the 10-minute lease; after expiry, reacquire and re-read before any further write.

### Safe resolution sequence

Choose a quoted heredoc delimiter that does not occur on a line in the answer, then pass the answer verbatim through stdin:

```sh
gh issue comment <n> --body-file - <<'WAYFINDER_RESOLUTION'
<answer>
WAYFINDER_RESOLUTION
```

After the resolution comment succeeds, acquire the map's `map-edit` lease, update and verify the map, and release that lease. Then run `gh issue close <n>`. Resolution is complete only when the verified map update and issue close both succeed.
