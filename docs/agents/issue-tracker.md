# Issue tracker

Issues and specs live in GitHub Issues.

- Repository: `BenDManning/jetkvm-mcp`
- CLI: `gh`
- Pull requests as a triage request surface: no

Read complete tickets in non-interactive sessions with `gh issue view <number> --json number,title,body,state,stateReason,author,labels,assignees,milestone,createdAt,updatedAt,closedAt,url,parent,blockedBy,blocking,comments`. Create or change issues only when the current task explicitly authorizes that tracker write.

Preserve the existing label vocabulary. Do not create, rename, or delete GitHub labels without maintainer approval.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a single issue with **child** issues as tickets.

- **Labels**: before creating issues, list existing names with `gh label list --limit 1000 --json name --jq '.[].name'`. Apply `wayfinder:map` or `wayfinder:<type>` only when that exact label exists. When it does not and label creation is not explicitly authorized, omit `--label`; prefix the map body with `<!-- wayfinder:map -->` and each child body with `<!-- wayfinder:type=<type> -->`. Never create labels implicitly.
- **Map**: a single issue holding the Notes / Decisions-so-far / Fog body. Create it with `gh issue create`, adding `--label wayfinder:map` only when the label check permits it.
- **Child ticket**: an issue linked to the map as a GitHub sub-issue (`gh api` on the sub-issues endpoint). Where sub-issues aren't enabled, add the child to a task list in the map body and put `Part of #<map>` at the top of the child body. Apply the matching `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`) only when the label check permits it. Once claimed, the ticket is assigned to the driving dev.
- **Blocking**: GitHub's **native issue dependencies** — the canonical, UI-visible representation. Add an edge with `gh api --method POST repos/<owner>/<repo>/issues/<child>/dependencies/blocked_by -F issue_id=<blocker-db-id>`, where `<blocker-db-id>` is the blocker's numeric **database id** (`gh api repos/<owner>/<repo>/issues/<n> --jq .id`, _not_ the `#number` or `node_id`). GitHub reports `issue_dependencies_summary.blocked_by` (open blockers only — the live gate). Where dependencies aren't available, fall back to a `Blocked by: #<n>, #<n>` line at the top of the child body. A ticket is unblocked when every blocker is closed.
- **Child source**: for the named map, page through `gh api --paginate --slurp 'repos/{owner}/{repo}/issues/<map>/sub_issues?per_page=100' | jq -r 'add | .[].number'`; the returned order is map order. Only when that endpoint is unavailable, parse that map's task-list children from `gh issue view <map> --json body --jq .body` in body order. Never substitute repository-wide issue order.
- **Blocker source**: choose independently of the child source. For each child, page through `gh api --paginate --slurp 'repos/{owner}/{repo}/issues/<child>/dependencies/blocked_by?per_page=100' | jq -r 'add | map(select(.state == "open")) | .[].number'` when native dependencies are available. Also honor every `Blocked by:` body reference after fetching its live state; when the dependency endpoint is unavailable, those body references are the fallback source. A child is blocked when either available source has an open issue.
- **Frontier**: read each child with the complete-ticket command, preserve child-source order, and discard closed, assigned, or blocked children. The first remaining child is the frontier.
- **Claim**: use the exclusive ticket-claim lease below. Assignment alone is not an exclusive claim.
- **Resolve**: use the safe resolution sequence below.

### Exclusive comment leases

Use issue comments as ordered, expiring lease attempts. Give each attempt a collision-resistant token and one kind: `ticket-claim` on a child or `map-edit` on the map. Post `<!-- wayfinder-lease:<kind>:<token> -->` with `gh issue comment <issue> --body-file -` and a quoted heredoc.

Read all comments with `gh api --paginate --slurp 'repos/{owner}/{repo}/issues/<issue>/comments?per_page=100' | jq 'add | sort_by(.id)'`. Count markers only from comments whose REST `author_association` is `OWNER`, `MEMBER`, or `COLLABORATOR`; a release matches only the same kind, token, and author login as its lease. A lease attempt is active for 10 minutes from its REST `created_at` when no later `<!-- wayfinder-release:<kind>:<token> -->` comment exists. Among active attempts of the same kind, the lowest numeric REST comment `id` is the sole winner. Wait until the posted token appears in the read; only the winner may proceed. A loser posts its matching release marker and selects another ticket or retries after the winner releases. An expired attempt never wins.

For a `ticket-claim`, the winner re-reads the ticket and its blockers, confirms it is still open, unassigned, and unblocked, then runs `gh issue edit <n> --add-assignee @me`. Re-read once more and start work only when the ticket is assigned to the claimant and the claimant's token remains the winning active lease. Otherwise remove the assignment if added, post the release marker, and stop. When abandoning work, remove the assignment and release the lease.

For a `map-edit`, the winner re-reads the latest map body after acquiring the lease, merges its decision pointer and fog changes into that body, and writes the complete merged body with `gh issue edit <map> --body-file -`. Re-read the map and verify the intended edits and every pre-existing Decisions-so-far pointer remain. Repair and verify again while holding the lease if necessary, then post the release marker. Finish within the 10-minute lease; after expiry, discard the draft and reacquire before writing.

### Safe resolution sequence

Choose a quoted heredoc delimiter that does not occur on a line in the answer, then pass the answer verbatim through stdin:

```sh
gh issue comment <n> --body-file - <<'WAYFINDER_RESOLUTION'
<answer>
WAYFINDER_RESOLUTION
```

After the resolution comment succeeds, acquire the map's `map-edit` lease, update and verify the map, and release that lease. Then run `gh issue close <n>`. Resolution is complete only when the verified map update and issue close both succeed.
