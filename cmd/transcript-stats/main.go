package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ardaarslan47/qa-hub/internal/transcript"
)

func main() {
	if err := run(os.Stdout, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "transcript-stats:", err)
		os.Exit(1)
	}
}

func run(out io.Writer, args []string) error {
	flags := flag.NewFlagSet("transcript-stats", flag.ContinueOnError)
	tickets := flags.Int("tickets", 20, "ticket count to project the root's context at")
	limit := flags.Int("limit", 200000, "root context limit")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 1 {
		return errors.New("usage: transcript-stats [-tickets N] [-limit TOKENS] <session.jsonl>")
	}
	report, err := transcript.Analyze(flags.Arg(0))
	if err != nil {
		return err
	}
	writeReport(out, report, *tickets, *limit)
	return nil
}

func writeReport(out io.Writer, report transcript.Report, tickets, limit int) {
	fmt.Fprintf(out, "root: %d calls, peak %d, final %d\n", report.Root.Calls, report.Root.Peak, report.Root.Final)
	fmt.Fprintf(out, "tickets: %d, spawns: %d, spawns per ticket: %.1f\n", report.Tickets, len(report.Spawns), report.SpawnsPerTicket)
	fmt.Fprintf(out, "base %d, per ticket %d, per group %d, finish %d\n", report.Base, report.PerTicket, report.PerGroup, report.Finish)
	fmt.Fprintf(out, "projected at %d tickets: %d, tickets within %d: %d\n", tickets, report.Projection(tickets), limit, report.TicketsWithin(limit))

	fmt.Fprintln(out, "\n| phase | spawns | root growth |")
	fmt.Fprintln(out, "|---|---|---|")
	for _, phase := range report.Phases {
		fmt.Fprintf(out, "| %s | %d | %d |\n", phase.Label, phase.Spawns, phase.Growth)
	}

	fmt.Fprintln(out, "\n| spawn | phase | root context |")
	fmt.Fprintln(out, "|---|---|---|")
	for _, spawn := range report.Spawns {
		fmt.Fprintf(out, "| %s | %s | %d |\n", spawn.Description, spawn.Label, spawn.RootContext)
	}

	fmt.Fprintln(out, "\n| agent | model | calls | peak context |")
	fmt.Fprintln(out, "|---|---|---|---|")
	for _, agent := range report.Agents {
		fmt.Fprintf(out, "| %s | %s | %d | %d |\n", agent.Description, agent.Model, agent.Calls, agent.Peak)
	}
}
