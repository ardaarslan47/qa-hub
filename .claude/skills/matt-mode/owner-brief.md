# Owner brief

You own one ticket of a matt-mode program. The root's brief gives you the user's goal, the ticket URL, the gate, and your report's name, and in a fix round the prior report paths and branch. You write the code; the root never reads it, so your report and commit are the whole record.

This brief is your playbook. Past poteto-mode's SKILL.md, read only the `tdd` skill and the domain docs. Skip poteto-mode's playbook routing and other skills.

1. **Start from current `main`.** Run `git fetch origin && git reset --hard origin/main` in your worktree. In a fix round, check out the branch from the brief instead and rebase it onto `origin/main`. In a rebase round, resolve the conflict and change nothing else.
2. **Read the ticket and its parent** with `gh issue view`. The ticket's acceptance criteria are the scope. A ticket that is wrong or unbuildable stops here: write why to your report and hand back `UNBUILDABLE`.
3. **Read the domain docs** per `docs/agents/domain.md`.
4. **Build test-first** with the `tdd` skill. Its seams are the ticket's acceptance criteria. In a fix round, start with a test that goes red on each blocking finding from the prior reports.
5. **Close with the gate.** Run it on the committed tree with output to `.matt-mode-gate.log` at your worktree root. Green is the bar.
6. **Commit once.** One Conventional Commit on top of `origin/main`, with `Refs #N` in the body for a ticket and `Refs #<parent>` for an interrogate finding. Stage by path: `.matt-mode-report.md` and `.matt-mode-gate.log` never go in the commit. Squash fix rounds into that one commit. The commit stays local; the root lands it.
7. **Write your report** to `.matt-mode-report.md` at your worktree root: what you built, the tests that cover each acceptance criterion, decisions the ticket left open and how you made them, and anything you saw outside scope.
8. **Hand back** at most 700 characters. The first line is the report's absolute path. Then the branch, the SHA, the gate result, and one line per open decision.
