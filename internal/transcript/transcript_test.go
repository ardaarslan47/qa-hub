package transcript_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ardaarslan47/qa-hub/internal/transcript"
)

type call struct {
	id      string
	input   int
	read    int
	created int
	spawns  []string
}

func line(c call) string {
	content := []map[string]any{{"type": "text", "text": "x"}}
	for _, description := range c.spawns {
		content = append(content, map[string]any{
			"type":  "tool_use",
			"name":  "Agent",
			"input": map[string]any{"description": description, "prompt": "p"},
		})
	}
	entry := map[string]any{
		"type": "assistant",
		"message": map[string]any{
			"id":      c.id,
			"content": content,
			"usage": map[string]any{
				"input_tokens":                c.input,
				"cache_read_input_tokens":     c.read,
				"cache_creation_input_tokens": c.created,
				"output_tokens":               10,
			},
		},
	}
	out, err := json.Marshal(entry)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func writeSession(t *testing.T, dir string, lines ...string) string {
	t.Helper()
	path := filepath.Join(dir, "session.jsonl")
	body := `{"type":"user","message":{"role":"user","content":"go"}}` + "\n" + strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRootContextCountsInputCacheReadAndCacheCreationOncePerMessage(t *testing.T) {
	dir := t.TempDir()
	path := writeSession(t, dir,
		line(call{id: "m1", input: 2, read: 0, created: 10000}),
		line(call{id: "m1", input: 2, read: 0, created: 10000}),
		line(call{id: "m2", input: 3, read: 10000, created: 30000}),
		line(call{id: "m3", input: 1, read: 20000, created: 4000}),
	)

	report, err := transcript.Analyze(path)
	if err != nil {
		t.Fatal(err)
	}

	if report.Root.Calls != 3 {
		t.Errorf("calls = %d, want 3", report.Root.Calls)
	}
	if report.Root.Peak != 40003 {
		t.Errorf("peak = %d, want 40003", report.Root.Peak)
	}
	if report.Root.Final != 24001 {
		t.Errorf("final = %d, want 24001", report.Root.Final)
	}
}

func writeProgramSession(t *testing.T) string {
	t.Helper()
	return writeSession(t, t.TempDir(),
		line(call{id: "m1", input: 1, created: 20000}),
		line(call{id: "m2", input: 1, read: 20000, created: 9999}),
		line(call{id: "m2", input: 1, read: 20000, created: 9999, spawns: []string{"#12 owner"}}),
		line(call{id: "m3", input: 1, read: 30000, created: 3999, spawns: []string{"#12 verifier", "#13 owner"}}),
		line(call{id: "m4", input: 1, read: 34000, created: 1999, spawns: []string{"group 1 interrogate"}}),
		line(call{id: "m5", input: 0, read: 36000, created: 4000}),
	)
}

func TestSpawnsAreAttributedToTicketsAndRootGrowthIsMeasuredFromTheFirstSpawn(t *testing.T) {
	path := writeProgramSession(t)

	report, err := transcript.Analyze(path)
	if err != nil {
		t.Fatal(err)
	}

	want := []transcript.Spawn{
		{Description: "#12 owner", Ticket: 12, RootContext: 30000},
		{Description: "#12 verifier", Ticket: 12, RootContext: 34000},
		{Description: "#13 owner", Ticket: 13, RootContext: 34000},
		{Description: "group 1 interrogate", Ticket: 0, RootContext: 36000},
	}
	if len(report.Spawns) != len(want) {
		t.Fatalf("spawns = %+v, want %+v", report.Spawns, want)
	}
	for i := range want {
		if report.Spawns[i] != want[i] {
			t.Errorf("spawn %d = %+v, want %+v", i, report.Spawns[i], want[i])
		}
	}
	if report.Tickets != 2 {
		t.Errorf("tickets = %d, want 2", report.Tickets)
	}
	if report.SpawnsPerTicket != 2 {
		t.Errorf("spawns per ticket = %v, want 2", report.SpawnsPerTicket)
	}
	if report.GrowthPerTicket != 5000 {
		t.Errorf("growth per ticket = %d, want 5000", report.GrowthPerTicket)
	}
	if report.ProjectedAt20Tickets != 130000 {
		t.Errorf("projection at 20 tickets = %d, want 130000", report.ProjectedAt20Tickets)
	}
}

func writeAgent(t *testing.T, dir, id, meta string, lines ...string) {
	t.Helper()
	subagents := filepath.Join(dir, "session", "subagents")
	if err := os.MkdirAll(subagents, 0o700); err != nil {
		t.Fatal(err)
	}
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(subagents, "agent-"+id+".jsonl"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subagents, "agent-"+id+".meta.json"), []byte(meta), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestEachSubagentReportsItsPeakContextWithDescriptionAndModel(t *testing.T) {
	dir := t.TempDir()
	path := writeSession(t, dir, line(call{id: "m1", input: 1, created: 999}))
	writeAgent(t, dir, "b",
		`{"agentType":"pstack:poteto-agent","description":"#12 verifier","model":"fable","effort":"xhigh"}`,
		line(call{id: "v1", input: 1, created: 50000}),
		line(call{id: "v2", input: 1, read: 50000, created: 1000}),
		line(call{id: "v3", input: 1, read: 30000}),
	)
	writeAgent(t, dir, "a",
		`{"agentType":"pstack:poteto-agent","description":"#12 owner","model":"sonnet","effort":"xhigh"}`,
		line(call{id: "o1", input: 1, created: 70000}),
	)

	report, err := transcript.Analyze(path)
	if err != nil {
		t.Fatal(err)
	}

	want := []transcript.Agent{
		{Description: "#12 owner", Model: "sonnet", Calls: 1, Peak: 70001},
		{Description: "#12 verifier", Model: "fable", Calls: 3, Peak: 51001},
	}
	if len(report.Agents) != len(want) {
		t.Fatalf("agents = %+v, want %+v", report.Agents, want)
	}
	for i := range want {
		if report.Agents[i] != want[i] {
			t.Errorf("agent %d = %+v, want %+v", i, report.Agents[i], want[i])
		}
	}
}

func TestSessionWithoutSubagentsHasNoAgents(t *testing.T) {
	dir := t.TempDir()
	path := writeSession(t, dir, line(call{id: "m1", input: 1}))

	report, err := transcript.Analyze(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Agents) != 0 {
		t.Errorf("agents = %+v, want none", report.Agents)
	}
}

func TestRootGrowthBetweenSpawnsIsAttributedToTheTicketOfTheLaterSpawn(t *testing.T) {
	report, err := transcript.Analyze(writeProgramSession(t))
	if err != nil {
		t.Fatal(err)
	}

	want := []transcript.Phase{
		{Ticket: 0, Spawns: 1, Growth: 2000},
		{Ticket: 12, Spawns: 2, Growth: 4000},
		{Ticket: 13, Spawns: 1, Growth: 0},
	}
	if len(report.Phases) != len(want) {
		t.Fatalf("phases = %+v, want %+v", report.Phases, want)
	}
	for i := range want {
		if report.Phases[i] != want[i] {
			t.Errorf("phase %d = %+v, want %+v", i, report.Phases[i], want[i])
		}
	}
}
