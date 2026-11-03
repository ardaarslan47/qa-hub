package main

import (
	"errors"
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
	if len(args) != 1 {
		return errors.New("usage: transcript-stats <session.jsonl>")
	}
	report, err := transcript.Analyze(args[0])
	if err != nil {
		return err
	}
	writeReport(out, report)
	return nil
}

func writeReport(out io.Writer, report transcript.Report) {
	fmt.Fprintf(out, "root: %d calls, peak %d, final %d\n", report.Root.Calls, report.Root.Peak, report.Root.Final)
	fmt.Fprintf(out, "tickets: %d, spawns: %d, spawns per ticket: %.1f\n", report.Tickets, len(report.Spawns), report.SpawnsPerTicket)
	fmt.Fprintf(out, "root growth per ticket: %d, projected at 20 tickets: %d\n", report.GrowthPerTicket, report.ProjectedAt20Tickets)

	fmt.Fprintln(out, "\n| ticket | spawns | root growth |")
	fmt.Fprintln(out, "|---|---|---|")
	for _, phase := range report.Phases {
		fmt.Fprintf(out, "| %d | %d | %d |\n", phase.Ticket, phase.Spawns, phase.Growth)
	}

	fmt.Fprintln(out, "\n| spawn | ticket | root context | growth since previous spawn |")
	fmt.Fprintln(out, "|---|---|---|---|")
	previous := 0
	for i, spawn := range report.Spawns {
		growth := "-"
		if i > 0 {
			growth = fmt.Sprint(spawn.RootContext - previous)
		}
		fmt.Fprintf(out, "| %s | %d | %d | %s |\n", spawn.Description, spawn.Ticket, spawn.RootContext, growth)
		previous = spawn.RootContext
	}

	fmt.Fprintln(out, "\n| agent | model | calls | peak context |")
	fmt.Fprintln(out, "|---|---|---|---|")
	for _, agent := range report.Agents {
		fmt.Fprintf(out, "| %s | %s | %d | %d |\n", agent.Description, agent.Model, agent.Calls, agent.Peak)
	}
}
