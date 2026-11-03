package transcript

import (
	"bufio"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Usage struct {
	Calls int
	Peak  int
	Final int
}

type Spawn struct {
	Description string
	Ticket      int
	RootContext int
}

type Phase struct {
	Ticket int
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
	Root                 Usage
	Agents               []Agent
	Spawns               []Spawn
	Phases               []Phase
	Tickets              int
	SpawnsPerTicket      float64
	GrowthPerTicket      int
	ProjectedAt20Tickets int
}

type entry struct {
	Type    string `json:"type"`
	Message struct {
		ID      string `json:"id"`
		Content []struct {
			Type  string `json:"type"`
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
	} `json:"message"`
}

func (e entry) size() int {
	usage := e.Message.Usage
	return usage.Input + usage.Read + usage.Created
}

var ticketReference = regexp.MustCompile(`#(\d+)`)

func Analyze(sessionPath string) (Report, error) {
	var report Report
	tickets := map[int]bool{}
	err := eachCall(sessionPath, func(e entry, newMessage bool) {
		if newMessage {
			report.Root.Calls++
			report.Root.Final = e.size()
			report.Root.Peak = max(report.Root.Peak, e.size())
		}
		for _, block := range e.Message.Content {
			if block.Type != "tool_use" || (block.Name != "Agent" && block.Name != "Task") {
				continue
			}
			spawn := Spawn{Description: block.Input.Description, RootContext: e.size()}
			if match := ticketReference.FindStringSubmatch(spawn.Description); match != nil {
				spawn.Ticket, _ = strconv.Atoi(match[1])
				tickets[spawn.Ticket] = true
			}
			report.Spawns = append(report.Spawns, spawn)
		}
	})
	if err != nil {
		return Report{}, err
	}

	report.Tickets = len(tickets)
	if report.Tickets > 0 {
		base := report.Spawns[0].RootContext
		report.SpawnsPerTicket = float64(len(report.Spawns)) / float64(report.Tickets)
		report.GrowthPerTicket = (report.Root.Final - base) / report.Tickets
		report.ProjectedAt20Tickets = base + 20*report.GrowthPerTicket
	}
	report.Phases = phases(report.Spawns)

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
		err = eachCall(path, func(e entry, newMessage bool) {
			if newMessage {
				agent.Calls++
				agent.Peak = max(agent.Peak, e.size())
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

func eachCall(path string, visit func(e entry, newMessage bool)) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	seen := map[string]bool{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for scanner.Scan() {
		var e entry
		if json.Unmarshal(scanner.Bytes(), &e) != nil || e.Type != "assistant" {
			continue
		}
		visit(e, !seen[e.Message.ID])
		seen[e.Message.ID] = true
	}
	return scanner.Err()
}

func phases(spawns []Spawn) []Phase {
	byTicket := map[int]Phase{}
	for i, spawn := range spawns {
		phase := byTicket[spawn.Ticket]
		phase.Ticket = spawn.Ticket
		phase.Spawns++
		if i > 0 {
			phase.Growth += spawn.RootContext - spawns[i-1].RootContext
		}
		byTicket[spawn.Ticket] = phase
	}
	result := slices.Collect(maps.Values(byTicket))
	slices.SortFunc(result, func(a, b Phase) int { return a.Ticket - b.Ticket })
	return result
}
