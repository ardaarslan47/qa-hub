# Ledger: #<parent> <title>

- Goal: <the user's goal, one line>
- Gate: <the command that runs every check CI runs>
- Sessions: <session id per root session, for cmd/transcript-stats>

## Groups

| Group | Tickets | Base | Interrogate |
|---|---|---|---|
| 1 | #12 #13 #14 #15 | <commit before the group's first landed commit> | <pending, running, findings count> |

## Tickets

| Ticket | Group | Blocked by | State | Branch | Owner SHA | Verdict | Retry | Pushed SHA | CI run | Follow-ups |
|---|---|---|---|---|---|---|---|---|---|---|
| #12 | 1 | - | ready | | | | 0 | | | |
| #13 | 1 | #12 | blocked | | | | 0 | | | |

## Events

Appended by `ledger.sh event`, newest last: `<time> <row> <from> -> <to>: <why>`, plus `<time> <label> worktree <path>` for each copied report.
