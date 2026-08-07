<!-- urkon:managed — installed by `urkon init`; edits are overwritten on the next run.
     Remove this marker line to take ownership of the file and urkon will leave it alone. -->

# Urkon workflow — core loop

This repository runs urkon's **core** loop: three phases (`spec`, `tests`, `code`) driven by
`phase:*` labels. This file describes that loop — its phases, its labels, and its board
columns.

Read it with the two files installed alongside it:

- `.urkon/mechanics.md` — the mode-independent machinery: sandboxing, retries,
  multi-instance ownership, identities, PR conventions, and the hard limits.
- `.urkon/agent-policy.md` — the rules the dispatched agent follows.

The issue tracker and its project board are the single source of truth for planning and
status.

## Three loops, three workflows

There are three **loops** — `spec`, `tests`, `code` — each a user⇄agent iteration.
The **spec** loop runs at the **issue** boundary; the **tests** and **code** loops run
at the **PR** boundary (one PR per issue). The user composes them into one of three
**workflows** by choosing the next phase manually at each hand-off:

- `spec → tests → code`
- `spec → code → tests`
- `spec → code`

There is no upfront declaration of the path; it emerges from which `phase:*` label the
user sets next.

## Roles & gates

- **User (`URKON_TRACKER_USER` on the tracker / `URKON_FORGE_USER` on the forge)** owns
  phase transitions (`phase:*`) and the `agent:review` trigger, and **merges PRs**. The agent
  never merges and never advances a phase on its own.
- **Agent (`URKON_TRACKER_AGENT` on the tracker / `URKON_FORGE_AGENT` on the
  forge)** plans, writes specs/tests/code, opens & updates the
  PR, and reviews it. It stamps `phase:*` **only at creation** (issue create, PR open).
  It does **not** move the Projects board — that is the watcher's (below).
- The **Watcher** (`urkon watch`) is deterministic: it polls, runs the label
  choreography, **moves the Projects board Status through the tracker provider at each
  step**, dispatches the agent headless, and streams the trace. It does **no git and no PR
  ops** — those are the agent's.

## Labels (state machine)

Three families, plus nothing else that this workflow reads. Unknown labels are always
preserved.

| Label               | Who sets                     | Meaning / next action |
|---------------------|------------------------------|-----------------------|
| `phase:spec`        | User; agent at issue creation | Spec loop active on the issue. Also the prompt selector. |
| `phase:tests`       | User; agent at PR open        | Tests loop active on the PR. Prompt selector. |
| `phase:code`        | User; agent at PR open        | Code loop active on the PR. Prompt selector. |
| `agent:review`      | User only                   | **The review-loop trigger.** "Agent, your turn." Consumed by the watcher at pickup. |
| `agent:debug`       | User (issue or PR)          | **The debugging trigger** (independent of `agent:review`). Runs the debugging workflow; consumed by the watcher at pickup. See [Debugging workflow](#debugging-workflow-agentdebug). |
| `agent:in-progress` | Watcher (agent)              | A headless run is actively working this item now. Added at pickup, removed at cleanup. |
| `agent:errored`     | Watcher (agent)              | A headless run hit an error. Durable auto-resume marker; kept on hand-back. See [Auto-resume](#auto-resume-on-transient-failure). |
| `agent:retrying`    | Watcher (agent)              | Auto-resume in progress (waiting to retry a failed run). Held across attempts, removed on recovery or hand-back. See [Auto-resume](#auto-resume-on-transient-failure). |
| `agent:<name>`      | User only                   | **Agent selector** (e.g. `agent:claude`, `agent:amp`). Routes the item to that agent; absent ⇒ `URKON_AGENT_CLI_NAME`. Persists across phases (never stripped). One per registered agent, created by `urkon init`. Label Guard reverts an agent-added selector. |
| `human:review`      | Agent; user may also set/remove | The agent finished a turn **or** is blocked on a user decision (see below). Best-effort removed by the watcher at pickup. |

`phase:*` is both the phase marker and the prompt selector; it is **never** a dedup key.
Dedup rides entirely on the trigger being consumed at pickup — the watcher fires only when a
trigger (`agent:review` or `agent:debug`) is present and the item is not already
`agent:in-progress`. When both triggers are present, `agent:debug` wins and the debug turn
consumes **both** triggers, so the review loop does not auto-fire afterward (the user re-adds
a trigger to continue).

**Blocked on a decision.** The no-assumptions hard rule holds in headless runs: the agent
never guesses. When it hits any decision it (1) posts a comment listing concrete options,
(2) sets `human:review`, (3) stops. There is **no separate blocked label** — a blocked
item is just `human:review` with no `agent:review`, so it sits idle until the user
answers and re-adds `agent:review`. The user disambiguates "done" from "blocked" by
reading the comment.

## Uniform watcher choreography

Every dispatch — regardless of phase — follows the same steps:

1. **Trigger.** Item (issue or PR) has a trigger (`agent:review` or `agent:debug`), is not
   `agent:in-progress`, and the in-flight guard is free.
2. **Pickup.** Consume the trigger (abort if it can't be removed), best-effort remove
   `human:review`, add `agent:in-progress`, then move the board Status to the phase's
   **working** state (`phase:spec`→Spec, `phase:tests`→Tests, `phase:code`→Code; a debug run
   makes no board move).
3. **Run.** Dispatch the selected prompt — the item's `phase:*` label picks the spec,
   tests, or code prompt; `agent:debug` picks the debug prompt (see
   [Debugging workflow](#debugging-workflow-agentdebug)).
4. **Cleanup.** Remove `agent:in-progress` (cancellation-immune, pass or fail), add
   `human:review`, then move the board Status to the phase's **review** state
   (`phase:tests`→Tests review, `phase:code`→Code review; Spec has no separate review
   column, so it stays Spec; a debug run makes no board move).
5. **Merge trigger.** A merged PR that still carries workflow labels (and not `bookkept`)
   dispatches the bookkeeping prompt; its cleanup adds the permanent, human-visible
   `bookkept` marker, **removes all workflow labels**, and moves the board Status to
   **Done**. GitHub's search index lags label writes, so the watcher also keeps an
   in-process at-most-once set of bookkept PRs — the durable dedup is the `bookkept`
   label, the in-process set closes the search-lag window so bookkeeping never re-runs.

All board moves go through the tracker provider; a provider with no board
configured returns `ErrUnsupported` and the step is a clean no-op, so the loop still runs
without a board.

Only one urkon instance runs per repo, and an in-process per-item in-flight guard means
the same issue/PR is never worked twice at once. Distinct items run independently, and
each dispatch executes in its **own git worktree** rather than a single shared tree — see
[Sandboxed runs & worktree isolation](#sandboxed-runs--worktree-isolation-urkon_sandbox).

Each loop keeps its own agent session keyed by the **feature id** (the issue number `N`):
the spec loop resumes `spec/N`, and the implementation loops (tests + code) share one
feature-stable `code/N` session that spans the issue→PR boundary. The **first** run of a
session is *cold* — the full initial prompt; every later run *resumes* it with a terse
**continuation** prompt. Bookkeeping resumes the feature's last session (`code/N`, else
`spec/N`). A resume id the agent rejects as expired falls back to a cold run. See
the resume behaviour described above.

Sessions key by **feature id** while worktrees fan out by **`(feature, agent)`** (see
[Sandboxed runs](#sandboxed-runs--worktree-isolation-urkon_sandbox)): switching the
`agent:<name>` selector mid-issue creates a second worktree while the session store still
keys on the feature — this interaction is intended.

## The loops

### Spec loop (issue is the live item)

1. The user or the agent creates an issue tagged `phase:spec` (the agent stamps it at creation).
   The issue body captures the spec directly.
2. When the issue is ready for the agent, the user sets `agent:review`.
3. The watcher picks it up (consume `agent:review`, drop `human:review`, add
   `agent:in-progress`) and runs the spec prompt. The agent refines the spec on the issue and
   captures its current state as a single **`## Full specification`** comment — the
   finalized spec the implementation loops build from.
4. On finish, cleanup sets `human:review`. User reviews.
5. **Iterate:** user re-adds `agent:review` → back to step 3.
6. **Advance:** user removes `phase:spec` and sets `phase:tests` **or** `phase:code`
   (plus `agent:review`) — this both starts the next loop and picks the workflow variant.

### Tests loop and Code loop (PR is the live item)

The two PR-bound loops are identical in choreography; they differ only by which prompt
runs and what the agent produces.

1. **First non-spec run.** Triggered by `agent:review` on the issue carrying `phase:tests`
   (or `phase:code`). The watcher reads the issue's **`## Full specification`** comment and
   **injects it by value** into the cold prompt, so the agent works **from that spec** without
   re-reading the whole issue thread; this is the spec→next-phase hand-off. If no
   `## Full specification` comment exists yet (spec phase skipped), the prompt's guard body
   has the agent write one from the issue first. When the work is green the agent
   **opens the PR itself** (`Closes #<issue>` in the body, or `Closes KAN-42` under JIRA) — this
   is the agent's git/PR job, not
   the watcher's — and stamps the new PR with the matching `phase:*` **and** `human:review`. The
   agent must set `human:review` on the PR itself because the watcher's cleanup labels only the
   item it picked up (the issue), not the PR the agent created out-of-band. From here the **PR is
   the live item**; the issue waits to be closed on merge — GitHub auto-closes a linked GitHub
   issue, but under JIRA GitHub cannot auto-close the issue (the `Closes KAN-42` line is a human
   breadcrumb), so urkon itself drives the JIRA issue to Done + closes it on merge. There is
   **no** issue↔PR label sync.
2. **Review.** User reviews the PR (code comments and/or PR comments) and sets
   `agent:review` on the PR.
3. **Iterate.** Watcher picks up the PR, the agent addresses the feedback, pushes to the same
   branch, replies per thread, and finishes with `human:review`. Back to step 2.
4. **Advance (tests → code).** User swaps the `phase:*` label on the PR (remove
   `phase:tests`, add `phase:code`) plus `agent:review`. Same PR, same branch — **one PR
   per issue**. The `spec → code` and `spec → code → tests` variants are just different
   orders of setting these labels.
5. **Merge.** When satisfied, the user **merges** the PR. The merge trigger dispatches the
   bookkeeping prompt, which strips the workflow labels and stamps `bookkept`. On a GitHub
   tracker the `Closes #<issue>` reference auto-closes the issue; under JIRA GitHub cannot,
   so urkon drives the linked JIRA issue to the mapped Done status and closes it via acli.

**Code review is this loop, not a separate one.** There is no review-only label or prompt:
the user's PR review is addressed by the same `agent:review` dispatch (steps 2–3 above). The
tests and code prompts instruct the agent to address the user's PR review feedback per thread on
each PR turn.

## Projects board (`URKON_TRACKER_GITHUB_PROJECT_NUMBER`)

Status field options: **Backlog → Spec → Tests → Tests review → Code → Code review → Done**.

The **watcher** keeps the board in sync through the tracker provider at each
choreography step (pickup, cleanup, merge bookkeeping) — the GitHub provider resolves the
board's Status field + option IDs once, adds the item to the board if absent, then sets the
single-select value. GitHub's built-in Projects automations only cover item-added /
item-closed / PR-merged — they cannot map labels → Status. A board move never blocks the
loop: a failure is logged, and a provider with no board configured no-ops via
`ErrUnsupported`.

State → Projects Status mapping:

| State                                   | Projects Status |
|-----------------------------------------|-----------------|
| (created)                               | Backlog         |
| `phase:spec`                            | Spec            |
| `phase:tests`, agent working / PR opened | Tests           |
| `phase:tests` + `human:review`          | Tests review    |
| `phase:code`, agent working             | Code            |
| `phase:code` + `human:review`           | Code review     |
| (PR merged)                             | Done            |
