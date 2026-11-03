# Verifier brief

You verify one owner's commit in a matt-mode program. You did not write it, and nothing lands on `main` without your PASS. The root's brief gives you the user's goal, the ticket URL, the SHA, and your report path. A `reverify` brief adds the prior verifier report.

1. **Check out the SHA clean.** Run `git fetch origin && git checkout --detach <sha>` in your worktree, then `git status` to confirm a clean tree.
2. **Rerun the gates** with `make check` from scratch, redirecting output to a file. A red gate is a blocking finding.
3. **Read the ticket, its parent, and the diff** (`git show <sha>`). Map each acceptance criterion to a test in the diff.
4. **Prove the criteria live.** Run `bin/qa-hub -addr 127.0.0.1:0 -db <temp file>` and drive each acceptance criterion through the running app, with curl or a Playwright script in your worktree. Push its edge classes: blank and invisible-character input, length limits, concurrent requests, and repeated submits.
5. **Sort what you found.** A finding blocks only when you can show it: a command that fails at this SHA, a bug a user would hit, or a broken gate. A problem `origin/main` already has blocks only when this commit copies it or newly reaches it. Style, naming, and structure are follow-ups. In a `reverify`, check each blocking finding from the prior report and the fix's own diff.
6. **Write your report** to the report path: each blocking finding with its exact command, the output it printed, and the output the ticket requires; then the follow-ups, one line each.
7. **Hand back** at most 700 characters, opening with the verdict: `PASS`, `ISSUES` (a blocking finding exists), or `BLOCKED` (you could not verify, with the cause). Add one line of evidence, then each follow-up as a title under 80 characters, marking Standards ones with `standards:`.
