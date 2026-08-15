# Issue tracker

Issues and specs live in GitHub Issues.

- Repository: `BenDManning/jetkvm-mcp`
- CLI: `gh`
- Pull requests as a triage request surface: no

Read tickets, including their comments, with `gh issue view <number> --comments`. Create or change issues only when the current task explicitly authorizes that tracker write.

Preserve the existing label vocabulary. Do not create, rename, or delete GitHub labels without maintainer approval.

## Wayfinding operations

Used by `/wayfinder`. The **map** is a single issue with **child** issues as tickets.

- **Labels**: before creating issues, list existing names with `gh label list --limit 1000 --json name --jq '.[].name'`. Apply `wayfinder:map` or `wayfinder:<type>` only when that exact label exists. When it does not and label creation is not explicitly authorized, omit `--label`; prefix the map body with `<!-- wayfinder:map -->` and each child body with `<!-- wayfinder:type=<type> -->`. Never create labels implicitly.
- **Map**: a single issue holding the Notes / Decisions-so-far / Fog body. Create it with `gh issue create`, adding `--label wayfinder:map` only when the label check permits it.
- **Child ticket**: an issue linked to the map as a GitHub sub-issue (`gh api` on the sub-issues endpoint). Where sub-issues aren't enabled, add the child to a task list in the map body and put `Part of #<map>` at the top of the child body. Apply the matching `wayfinder:<type>` (`research`/`prototype`/`grilling`/`task`) only when the label check permits it. Once claimed, the ticket is assigned to the driving dev.
- **Blocking**: GitHub's **native issue dependencies** — the canonical, UI-visible representation. Add an edge with `gh api --method POST repos/<owner>/<repo>/issues/<child>/dependencies/blocked_by -F issue_id=<blocker-db-id>`, where `<blocker-db-id>` is the blocker's numeric **database id** (`gh api repos/<owner>/<repo>/issues/<n> --jq .id`, _not_ the `#number` or `node_id`). GitHub reports `issue_dependencies_summary.blocked_by` (open blockers only — the live gate). Where dependencies aren't available, fall back to a `Blocked by: #<n>, #<n>` line at the top of the child body. A ticket is unblocked when every blocker is closed.
- **Frontier query**: for the named map, run `gh api --paginate --slurp 'repos/{owner}/{repo}/issues/<map>/sub_issues?per_page=100' | jq -r 'add | map(select(.state == "open" and (.assignees | length == 0))) | .[].number'`; this preserves child order and yields open, unassigned candidates. For each candidate in order, run `gh api --paginate --slurp 'repos/{owner}/{repo}/issues/<child>/dependencies/blocked_by?per_page=100' | jq -r 'add | any(.[]; .state == "open")'`; the first candidate returning `false` is the frontier. Where sub-issues aren't available, read only the named map's body with `gh issue view <map> --json body --jq .body`, parse its task-list children in body order, and discard assigned children or children whose `Blocked by:` references include an open issue. Never use repository-wide issue order.
- **Claim**: `gh issue edit <n> --add-assignee @me` — the session's first write.
- **Resolve**: `gh issue comment <n> --body "<answer>"`, then `gh issue close <n>`, then append a context pointer (gist + link) to the map's Decisions-so-far.
