package transcript

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

type Usage struct {
	Calls int
	Peak  int
	Final int
}

type Spawn struct {
	Description string
	Label       string
	RootContext int
}

type Phase struct {
	Label  string
	Spawns int
	Growth int
}

type Agent struct {
	Description string `json:"description"`
	Model       string `json:"model"`
	Calls       int
	Peak        int
}

type Report struct {
	Root            Usage
	Agents          []Agent
	Spawns          []Spawn
	Phases          []Phase
	Tickets         int
	SpawnsPerTicket float64
	Base            int
	PerTicket       int
	PerGroup        int
	Finish          int
}

const groupSize = 5

func (r Report) Projection(tickets int) int {
	groups := (tickets + groupSize - 1) / groupSize
	return r.Base + tickets*r.PerTicket + groups*r.PerGroup + r.Finish
}

func (r Report) TicketsWithin(limit int) int {
	if r.PerTicket <= 0 {
		return 0
	}
	tickets := 0
	for r.Projection(tickets+1) <= limit {
		tickets++
	}
	return tickets
}

type entry struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message"`
}

type assistantMessage struct {
	ID      string `json:"id"`
	Content []struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Name  string `json:"name"`
		Input struct {
			Description string `json:"description"`
		} `json:"input"`
	} `json:"content"`
	Usage struct {
		Input   int `json:"input_tokens"`
		Read    int `json:"cache_read_input_tokens"`
		Created int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
}

func (m assistantMessage) size() int {
	return m.Usage.Input + m.Usage.Read + m.Usage.Created
}

type call struct {
	label  string
	growth int
}

type userMessage struct {
	Content json.RawMessage `json:"content"`
}

type toolResult struct {
	Type      string          `json:"type"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
}

var (
	spawnLabel   = regexp.MustCompile(`#\d+|g\d+-\d+|group \d+`)
	findingLabel = regexp.MustCompile(`^g(\d+)-\d+$`)
	launchedID   = regexp.MustCompile(`agentId: (\w+)`)
	handBackID   = regexp.MustCompile(`<agent-message from="(\w+)"`)
)

func labelOf(description string) string {
	if match := spawnLabel.FindString(description); match != "" {
		return match
	}
	return description
}

func Analyze(sessionPath string) (Report, error) {
	var report Report
	var calls []call
	seen := map[string]bool{}
	labels := map[string]string{}
	agentLabels := map[string]string{}
	current, lastHandBack, previous := "setup", -1, -1
	err := eachEntry(sessionPath, func(e entry) {
		switch e.Type {
		case "assistant":
			var m assistantMessage
			if json.Unmarshal(e.Message, &m) != nil {
				return
			}
			if !seen[m.ID] {
				seen[m.ID] = true
				report.Root.Calls++
				report.Root.Final = m.size()
				report.Root.Peak = max(report.Root.Peak, m.size())
				if previous >= 0 {
					calls = append(calls, call{label: current, growth: m.size() - previous})
				}
				previous = m.size()
			}
			for _, block := range m.Content {
				if block.Type != "tool_use" || (block.Name != "Agent" && block.Name != "Task") {
					continue
				}
				spawn := Spawn{Description: block.Input.Description, Label: labelOf(block.Input.Description), RootContext: m.size()}
				report.Spawns = append(report.Spawns, spawn)
				labels[block.ID] = spawn.Label
				current = spawn.Label
			}
		case "user":
			var m userMessage
			if json.Unmarshal(e.Message, &m) != nil {
				return
			}
			var results []toolResult
			if json.Unmarshal(m.Content, &results) == nil {
				for _, result := range results {
					if match := launchedID.FindSubmatch(result.Content); match != nil && labels[result.ToolUseID] != "" {
						agentLabels[string(match[1])] = labels[result.ToolUseID]
					}
				}
				return
			}
			var text string
			if json.Unmarshal(m.Content, &text) != nil {
				return
			}
			if match := handBackID.FindStringSubmatch(text); match != nil {
				if label := agentLabels[match[1]]; label != "" {
					current = label
					lastHandBack = len(calls)
				}
			}
		}
	})
	if err != nil {
		return Report{}, err
	}

	report.Phases = phases(calls, lastHandBack, report.Spawns)
	if len(report.Spawns) > 0 {
		report.Base = report.Spawns[0].RootContext
	}
	ticketSpawns, ticketGrowth, groups, groupGrowth := 0, 0, map[string]bool{}, 0
	for _, phase := range report.Phases {
		switch {
		case strings.HasPrefix(phase.Label, "#"):
			report.Tickets++
			ticketSpawns += phase.Spawns
			ticketGrowth += phase.Growth
		case strings.HasPrefix(phase.Label, "group "):
			groups[strings.TrimPrefix(phase.Label, "group ")] = true
			groupGrowth += phase.Growth
		case findingLabel.MatchString(phase.Label):
			groups[findingLabel.FindStringSubmatch(phase.Label)[1]] = true
			groupGrowth += phase.Growth
		case phase.Label == "finish":
			report.Finish = phase.Growth
		}
	}
	if report.Tickets > 0 {
		report.SpawnsPerTicket = float64(ticketSpawns) / float64(report.Tickets)
		report.PerTicket = ticketGrowth / report.Tickets
	}
	if len(groups) > 0 {
		report.PerGroup = groupGrowth / len(groups)
	}

	report.Agents, err = readAgents(filepath.Join(strings.TrimSuffix(sessionPath, ".jsonl"), "subagents"))
	if err != nil {
		return Report{}, err
	}
	return report, nil
}

func readAgents(dir string) ([]Agent, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "agent-*.jsonl"))
	if err != nil {
		return nil, err
	}
	var agents []Agent
	for _, path := range paths {
		var agent Agent
		meta, err := os.ReadFile(strings.TrimSuffix(path, ".jsonl") + ".meta.json")
		if err == nil {
			err = json.Unmarshal(meta, &agent)
		}
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		err = eachCall(path, func(m assistantMessage, newMessage bool) {
			if newMessage {
				agent.Calls++
				agent.Peak = max(agent.Peak, m.size())
			}
		})
		if err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}
	slices.SortStableFunc(agents, func(a, b Agent) int { return strings.Compare(a.Description, b.Description) })
	return agents, nil
}

func eachCall(path string, visit func(m assistantMessage, newMessage bool)) error {
	seen := map[string]bool{}
	return eachEntry(path, func(e entry) {
		var m assistantMessage
		if e.Type != "assistant" || json.Unmarshal(e.Message, &m) != nil {
			return
		}
		visit(m, !seen[m.ID])
		seen[m.ID] = true
	})
}

func eachEntry(path string, visit func(e entry)) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for scanner.Scan() {
		var e entry
		if json.Unmarshal(scanner.Bytes(), &e) != nil {
			continue
		}
		visit(e)
	}
	return scanner.Err()
}

func phases(calls []call, lastHandBack int, spawns []Spawn) []Phase {
	var result []Phase
	index := map[string]int{}
	add := func(label string) int {
		i, ok := index[label]
		if !ok {
			i = len(result)
			index[label] = i
			result = append(result, Phase{Label: label})
		}
		return i
	}
	add("setup")
	for _, spawn := range spawns {
		result[add(spawn.Label)].Spawns++
	}
	for i, call := range calls {
		label := call.label
		if lastHandBack >= 0 && i >= lastHandBack {
			label = "finish"
		}
		result[add(label)].Growth += call.growth
	}
	return result
}
