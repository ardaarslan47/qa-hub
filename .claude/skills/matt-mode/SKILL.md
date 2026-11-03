---
name: matt-mode
description: "Run a parent issue's sub-issues as a program through pstack's Orchestrate playbook: a fresh owner and an independent verifier per ticket, a ledger on disk, an interrogate per group. Use for /matt-mode or a parent issue URL from to-spec and to-tickets."
disable-model-invocation: true
---

# Matt mode

Read the **poteto-mode** skill in full, then its `playbooks/orchestrate.md` and `playbooks/pause-safely.md`. The skill's path is the last line of `ls ~/.claude/plugins/cache/cursor-plugins/pstack/*/skills/poteto-mode/SKILL.md`. Route to Orchestrate whatever the ticket count. This file lists only where this program differs. Everything not named here, including model choice per role, comes from poteto-mode, Orchestrate, and `~/.claude/rules/pstack-models.md`.

## Input

The argument is a parent issue URL. Its body is the spec from `/to-spec`. Its sub-issues are the tickets from `/to-tickets`, each with acceptance criteria and blocking edges. Fetch both with `gh` per `docs/agents/issue-tracker.md`.

The spec and tickets are the user's design, already grilled. Keep every ticket's scope as written: no re-splitting, merging, or re-scoping. A ticket that is wrong or unbuildable goes to the user as a report.

## Root

You are the root, and you run the program. You hold the ledger, pick ready tickets, spawn owners and verifiers, push, and watch CI. Owners write the code and verifiers read it, so you work from their verdicts and report files. This replaces poteto-mode's "Review the diff" and Orchestrate's "Read the diff or the artifact" for the root.

Guard your context: it must hold a 30-ticket program. Read hand-backs, the ledger, and command exit codes. Send long command output to a log file under the program directory and read its tail only on failure. Read a report file only when a decision needs more than its hand-back.

## Translation from Orchestrate

- **PR** means one ticket's single commit that lands on `main`. This repo has no PRs (ADR 0001, `CLAUDE.md`).
- **Only the root pushes.** This replaces "the owner merges". `main` is the one shared write target, so one writer removes push races.
- **Workers run locally,** each in its own worktree from current `origin/main` (Agent `isolation: "worktree"`).
- **No sub-coordinators.** When your context nears its limit, stop per Pause safely. A new session resumes from the ledger.

## Roles

| Role | Role line | Spawn | Description |
|---|---|---|---|
| Owner | `feature, refactoring`; `hardest tasks` for a ticket with cross-cutting design or subtle logic | `pstack:poteto-agent`, worktree | `#N owner`, `#N fix K` |
| Verifier | `swarm workers` | `pstack:poteto-agent`, worktree | `#N verifier`, `#N reverify K` |
| Adversary | `interrogate reviewers`, first entry | `general-purpose`, worktree | `group G interrogate` |

Each description starts with the ticket reference exactly as shown, because `cmd/transcript-stats` attributes spawns to tickets by the first `#N` in it.

## Program directory

Everything the program writes lives in `.matt-mode/<parent>/` at the root of the main checkout, which git ignores:

- `ledger.md`: the program's state, from [ledger-template.md](ledger-template.md).
- `reports/`: one file per agent run, named `<ticket>-<role>-<round>.md`, such as `12-owner-0.md`, `12-verifier-0.md`, `12-owner-1.md`, `g1-interrogate.md`.
- `landing/`: your landing worktree, detached at `origin/main`.
- `logs/`: output of your own `make check` and CI watch runs.

The ledger is the only source for resuming. Update it on every state change before you act on the next one, and take each decision from the ledger, not from memory. Start by reading it: when `ledger.md` exists, resume from it.

## Setup

1. Read the parent and its sub-issues once. Write the ledger: one row per ticket with its blocking edges, every row `blocked` or `ready`.
2. Split tickets into groups of 4-6 along the blocking order. A ticket shares a group with its blockers or follows them. Record the groups in the ledger.
3. Create the landing worktree: `git worktree add --detach .matt-mode/<parent>/landing origin/main`.

## Ticket states

A ticket moves only along these transitions.

| From | To | When |
|---|---|---|
| `blocked` | `ready` | every blocker is `landed` |
| `ready` | `building` | owner spawned |
| `building` | `verifying` | owner handed back a SHA with `make check` green |
| `verifying` | `landing` | verdict PASS |
| `landing` | `landed` | pushed, CI green, issue closed |
| `building`, `verifying`, `landing` | `building` | a failure while `retry < 2`; `retry += 1` |
| `building`, `verifying`, `landing` | `dropped` | a failure while `retry = 2` |
| `building` | `dropped` | owner handed back `UNBUILDABLE` |

A **failure** is any of: an owner hand-back without a green `make check`, an ISSUES verdict, a BLOCKED verdict, a red `make check` in landing, a red CI run, or a cherry-pick conflict after the ticket's first one. `retry` starts at 0 and counts failures, so a ticket gets at most two fix rounds. Retry each failure by cause per Orchestrate's Supervision: a BLOCKED verdict reruns a verifier on the same SHA, everything else spawns a fresh owner. A ticket's first cherry-pick conflict is free and goes to a fresh owner as a rebase round.

`UNBUILDABLE` means the ticket is wrong or unbuildable. Comment on its issue as for any `dropped` ticket, and name it to the user in the next report.

## Loop

1. **Start with one ticket.** Take the first `ready` ticket alone to `landed` before you spawn another owner. After that, spawn every `ready` ticket in the current group in parallel, about ten agents in flight. A group starts only after the previous group's interrogate is done and its fixes have landed, so later tickets never build on an unfound bug.
2. **Owner.** Spawn per the Roles table. The brief is four lines: the user's goal, the ticket URL, the report path, and "Follow `.claude/skills/matt-mode/owner-brief.md`". A fix round adds the prior report paths, the failure, and the branch from the ledger. The issue body stays out of the brief, because the owner reads it with `gh`. Record the owner's branch and SHA from its hand-back.
3. **Verifier.** Spawn on the owner's SHA with the same four lines pointing at `verifier-brief.md`, plus the SHA. The first code-ready SHA gets the full verification. After a fix round, one verifier checks the new SHA against the prior verifier report, with `reverify` in the brief.
4. **Landing.** On PASS, in the landing worktree: `git fetch origin && git checkout --detach origin/main && git cherry-pick <sha>`. A conflict aborts the cherry-pick and goes to a fresh owner as a rebase round on current `main`, then to a single verifier. On a clean pick, run `make check > logs/<N>-landing.log 2>&1`. On green, push with `git push origin HEAD:main`, record the pushed SHA, and watch each workflow run for that SHA with `gh run watch <id> --exit-status > logs/<N>-ci-<id>.log` in the background. Land one ticket at a time, so each push is checked on the `main` it lands on.
5. **CI.** On green, close the issue with `gh issue close <N> --comment "Landed in <sha>, CI <run url>"` and mark the ticket `landed`. On red, rerun once with `gh run rerun --failed`, since a flake is not a failure. Still red: revert it at once with `git revert --no-edit <pushed sha>`, push the revert, and record a failure with the CI log path. `main` stays green, and the fix round rebuilds the ticket's one commit from its branch.
6. **Follow-ups.** Verifier observations that don't block go to the ticket's ledger row and the final report, never to a fix round. When the same Standards observation arrives a second time, encode it with `/encode-rule` as a git hook, or record it as a follow-up when a hook can't check it.

## Group interrogate

When a group's last ticket lands, spawn one adversary per the Roles table. Its brief names the user's goal, the parent URL, the range `<base>..HEAD` where `<base>` is the commit before the group's first landed commit, the report path `reports/g<G>-interrogate.md`, and: "Read `<pstack>/skills/interrogate/SKILL.md` and act as its only adversary on this range, with the parent issue as the intent. Rerun each finding's command once before you report it. Return at most 700 characters: the count of verified findings and one line per finding." The last group's interrogate replaces a final interrogate.

Each verified finding becomes a ledger row `g<G>-<k>` in the same group and goes through owner, single verifier, and landing like a ticket, with no issue to close.

A group is **done** when its interrogate has run and every row in it is `landed` or `dropped`, or `blocked` behind a `dropped` row. Its interrogate starts when every row is in one of those states.

## Dropped tickets and stopping

A `dropped` ticket gets a comment on its issue: what each round tried, the last verdict, and a summary of its last report. Every ticket that depends on it, directly or through another, stays `blocked`. Keep going while any ticket is `ready`. When no ticket is `ready`, `building`, `verifying`, or `landing`, and no interrogate is running or due, the program stops: run Finish and report what waits on whom.

Only these reach the user: a wrong or unbuildable ticket, the program stopping, and actions on poteto-mode's Always-pause list. Retries, CI flakes, and verdict triage stay with you.

## Finish

1. Comment on the parent with a table, one row per ticket: commit, verdict, retries, CI run, and what was left open. List the parked follow-ups below it. Close the parent only when every ticket landed.
2. Measure the run: `go run ./cmd/transcript-stats ~/.claude/projects/<project>/<session>.jsonl > .matt-mode/<parent>/stats.md`. When the program spanned sessions, run it once per session.
3. Remove the landing worktree.

## Reply

Per Orchestrate's Reply. Add the parent comment's table and, from `stats.md`, root growth per ticket, spawns per ticket, and the projection at 20 tickets.
