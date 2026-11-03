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
			"id":    "tu-" + description,
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

func user(content any) string {
	out, err := json.Marshal(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": content}})
	if err != nil {
		panic(err)
	}
	return string(out)
}

func launched(description, agentID string) string {
	return user([]map[string]any{{
		"type":        "tool_result",
		"tool_use_id": "tu-" + description,
		"content":     []map[string]any{{"type": "text", "text": "Async agent launched successfully.\nagentId: " + agentID + " (internal ID)"}},
	}})
}

func handBack(agentID string) string {
	return user("Another Claude session sent a message:\n<agent-message from=\"" + agentID + "\">\n[Subagent hand-back] PASS\n</agent-message>")
}

func writeProgramSession(t *testing.T) string {
	t.Helper()
	return writeSession(t, t.TempDir(),
		line(call{id: "m1", input: 1, created: 20000}),
		line(call{id: "m2", read: 20001, created: 9999, spawns: []string{"#12 owner"}}),
		launched("#12 owner", "a1"),
		line(call{id: "m3", read: 30000, created: 2000}),
		handBack("a1"),
		line(call{id: "m4", read: 32000, created: 3000, spawns: []string{"#12 verifier"}}),
		launched("#12 verifier", "a2"),
		line(call{id: "m5", read: 35000, created: 1000}),
		handBack("a2"),
		line(call{id: "m6", read: 36000, created: 4000, spawns: []string{"#13 owner"}}),
		launched("#13 owner", "a3"),
		line(call{id: "m7", read: 40000, created: 1000}),
		handBack("a3"),
		line(call{id: "m8", read: 41000, created: 3000, spawns: []string{"group 1 interrogate"}}),
		launched("group 1 interrogate", "a4"),
		line(call{id: "m9", read: 44000, created: 1000}),
		handBack("a4"),
		line(call{id: "m10", read: 45000, created: 2000}),
		line(call{id: "m11", read: 47000, created: 1000}),
	)
}

func TestSpawnsAreLabeledByTheirTicketFindingOrGroup(t *testing.T) {
	report, err := transcript.Analyze(writeProgramSession(t))
	if err != nil {
		t.Fatal(err)
	}

	want := []transcript.Spawn{
		{Description: "#12 owner", Label: "#12", RootContext: 30000},
		{Description: "#12 verifier", Label: "#12", RootContext: 35000},
		{Description: "#13 owner", Label: "#13", RootContext: 40000},
		{Description: "group 1 interrogate", Label: "group 1", RootContext: 44000},
	}
	if len(report.Spawns) != len(want) {
		t.Fatalf("spawns = %+v, want %+v", report.Spawns, want)
	}
	for i := range want {
		if report.Spawns[i] != want[i] {
			t.Errorf("spawn %d = %+v, want %+v", i, report.Spawns[i], want[i])
		}
	}
}

func TestRootGrowthGoesToTheAgentWhoseBriefOrHandBackCausedIt(t *testing.T) {
	report, err := transcript.Analyze(writeProgramSession(t))
	if err != nil {
		t.Fatal(err)
	}

	want := []transcript.Phase{
		{Label: "setup", Spawns: 0, Growth: 9999},
		{Label: "#12", Spawns: 2, Growth: 10000},
		{Label: "#13", Spawns: 1, Growth: 4000},
		{Label: "group 1", Spawns: 1, Growth: 1000},
		{Label: "finish", Spawns: 0, Growth: 3000},
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

func TestProjectionAddsMarginalTicketsAnInterrogatePerFiveTicketsAndFinish(t *testing.T) {
	report, err := transcript.Analyze(writeProgramSession(t))
	if err != nil {
		t.Fatal(err)
	}

	if report.Base != 30000 || report.PerTicket != 7000 || report.PerGroup != 1000 || report.Finish != 3000 {
		t.Errorf("base, per ticket, per group, finish = %d, %d, %d, %d, want 30000, 7000, 1000, 3000",
			report.Base, report.PerTicket, report.PerGroup, report.Finish)
	}
	if report.Tickets != 2 || report.SpawnsPerTicket != 1.5 {
		t.Errorf("tickets, spawns per ticket = %d, %v, want 2, 1.5", report.Tickets, report.SpawnsPerTicket)
	}
	if got := report.Projection(20); got != 177000 {
		t.Errorf("projection at 20 tickets = %d, want 177000", got)
	}
	if got := report.TicketsWithin(200000); got != 23 {
		t.Errorf("tickets within 200000 = %d, want 23", got)
	}
}

func TestFindingRowsCountTowardTheirGroup(t *testing.T) {
	path := writeSession(t, t.TempDir(),
		line(call{id: "m1", input: 1, created: 9999, spawns: []string{"#5 owner"}}),
		launched("#5 owner", "a1"),
		handBack("a1"),
		line(call{id: "m2", read: 10000, created: 2000, spawns: []string{"group 1 interrogate"}}),
		launched("group 1 interrogate", "a2"),
		handBack("a2"),
		line(call{id: "m3", read: 12000, created: 1000, spawns: []string{"g1-1 owner"}}),
		launched("g1-1 owner", "a3"),
		handBack("a3"),
		line(call{id: "m4", read: 13000, created: 3000}),
	)

	report, err := transcript.Analyze(path)
	if err != nil {
		t.Fatal(err)
	}
	if report.Tickets != 1 || report.PerTicket != 2000 || report.PerGroup != 1000 || report.Finish != 3000 {
		t.Errorf("tickets, per ticket, per group, finish = %d, %d, %d, %d, want 1, 2000, 1000, 3000",
			report.Tickets, report.PerTicket, report.PerGroup, report.Finish)
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
