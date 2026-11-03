#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
ledger.sh <ledger.md> set <row> <Column>=<value> [<Column>=<value> ...]
ledger.sh <ledger.md> get <row> <Column>
ledger.sh <ledger.md> event <text>
<row> is a row's first cell: a ticket, a finding row such as g1-2, or a group number. Values cannot contain | or newlines.
USAGE
  exit 2
}
[ $# -ge 2 ] || usage
ledger=$1 command=$2
shift 2
[ -f "$ledger" ] || { echo "ledger.sh: no ledger at $ledger" >&2; exit 1; }

case $command in
set)
  [ $# -ge 2 ] || usage
  row=$1
  shift
  for pair in "$@"; do
    case $pair in *=*) ;; *) echo "ledger.sh: '$pair' is not Column=value" >&2; exit 2 ;; esac
    case $pair in *'|'*) echo "ledger.sh: '$pair' contains |" >&2; exit 2 ;; esac
  done
  updated=$(mktemp)
  trap 'rm -f "$updated" "$updated.row"' EXIT
  awk -v row="$row" -v rowfile="$updated.row" -v pairs="$(printf '%s\n' "$@")" '
    function trim(s) { gsub(/^[ \t]+|[ \t]+$/, "", s); return s }
    BEGIN { n = split(pairs, list, "\n"); for (i = 1; i <= n; i++) if (list[i] != "") { k = index(list[i], "="); want[substr(list[i], 1, k - 1)] = substr(list[i], k + 1) } }
    /^\|/ {
      cells = split($0, cell, "|")
      if (!in_table) { in_table = 1; delete header; for (i = 2; i < cells; i++) header[trim(cell[i])] = i; print; next }
      if (!done && trim(cell[2]) == row) {
        missing = ""
        for (c in want) if (!(c in header)) missing = c
        if (missing == "") {
          for (c in want) cell[header[c]] = " " want[c] " "
          line = ""; for (i = 1; i < cells; i++) line = line cell[i] "|"
          print line; print line > rowfile; done = 1; next
        }
      }
      print; next
    }
    { in_table = 0; print }
    END {
      if (done) exit 0
      if (missing != "") print "ledger.sh: row " row " has no column " missing > "/dev/stderr"
      else print "ledger.sh: no row " row > "/dev/stderr"
      exit 1
    }' "$ledger" > "$updated"
  mv "$updated" "$ledger"
  cat "$updated.row"
  ;;
get)
  [ $# -eq 2 ] || usage
  awk -v row="$1" -v column="$2" '
    function trim(s) { gsub(/^[ \t]+|[ \t]+$/, "", s); return s }
    /^\|/ {
      split($0, cell, "|")
      if (!in_table) { in_table = 1; delete header; for (i in cell) header[trim(cell[i])] = i; next }
      if (trim(cell[2]) == row && (column in header)) { print trim(cell[header[column]]); found = 1; exit }
      next
    }
    { in_table = 0 }
    END { if (!found) { print "ledger.sh: no " column " for row " row > "/dev/stderr"; exit 1 } }' "$ledger"
  ;;
event)
  [ $# -ge 1 ] || usage
  printf '%s %s\n' "$(date -u +%H:%M:%SZ)" "$*" >> "$ledger"
  ;;
*) usage ;;
esac
