package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// SocketMessage represents a combined payload for requests, responses, and events.
type SocketMessage struct {
	ID     string                       `json:"id,omitempty"`
	Method string                       `json:"method,omitempty"`
	Params interface{}                  `json:"params,omitempty"`
	Result *AgentListResult             `json:"result,omitempty"`
	Error  interface{}                  `json:"error,omitempty"`
	Event  string                       `json:"event,omitempty"`
	Data   *PaneAgentStatusChangedEvent `json:"data,omitempty"`
}

// AgentListResult is the result payload returned by agent.list.
type AgentListResult struct {
	Type   string      `json:"type"`
	Agents []AgentInfo `json:"agents"`
}

// AgentInfo holds detailed information for a single Herdr agent.
type AgentInfo struct {
	TerminalID    string            `json:"terminal_id"`
	Name          string            `json:"name"`
	Agent         *string           `json:"agent,omitempty"`
	AgentStatus   string            `json:"agent_status"`
	WorkspaceID   string            `json:"workspace_id"`
	TabID         string            `json:"tab_id"`
	PaneID        string            `json:"pane_id"`
	Focused       bool              `json:"focused"`
	CWD           string            `json:"cwd"`
	ForegroundCWD string            `json:"foreground_cwd"`
	Revision      uint64            `json:"revision"`
	CustomStatus  *string           `json:"custom_status,omitempty"`
	DisplayAgent  *string           `json:"display_agent,omitempty"`
	StateLabels   map[string]string `json:"state_labels,omitempty"`
}

// PaneAgentStatusChangedEvent contains status details sent in stream events.
type PaneAgentStatusChangedEvent struct {
	Agent        *string           `json:"agent,omitempty"`
	AgentStatus  string            `json:"agent_status"`
	CustomStatus *string           `json:"custom_status,omitempty"`
	DisplayAgent *string           `json:"display_agent,omitempty"`
	PaneID       string            `json:"pane_id"`
	StateLabels  map[string]string `json:"state_labels,omitempty"`
	Title        *string           `json:"title,omitempty"`
	WorkspaceID  string            `json:"workspace_id"`
}

// Subscription defines the subscription type.
type Subscription struct {
	Type   string `json:"type"`
	PaneID string `json:"pane_id,omitempty"`
}

// EventsSubscribeParams represents subscription command arguments.
type EventsSubscribeParams struct {
	Subscriptions []Subscription `json:"subscriptions"`
}

// getAgents connects to the socket, retrieves the initial agent list, and closes the connection.
func getAgents(socketPath string) ([]AgentInfo, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	id := fmt.Sprintf("req_%d", time.Now().UnixNano())
	req := SocketMessage{
		ID:     id,
		Method: "agent.list",
		Params: map[string]interface{}{},
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(append(reqBytes, '\n'))
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(conn)
	for {
		var msg SocketMessage
		if err := dec.Decode(&msg); err != nil {
			return nil, err
		}

		if msg.ID == id {
			if msg.Error != nil {
				return nil, fmt.Errorf("API error: %v", msg.Error)
			}
			if msg.Result != nil {
				return msg.Result.Agents, nil
			}
			return nil, fmt.Errorf("empty result")
		}
	}
}

func getTabLabels(socketPath string, workspaceID string) (map[string]string, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	id := fmt.Sprintf("req_tabs_%d", time.Now().UnixNano())
	req := map[string]interface{}{
		"id":     id,
		"method": "tab.list",
		"params": map[string]interface{}{
			"workspace_id": workspaceID,
		},
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(append(reqBytes, '\n'))
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(conn)
	for {
		var msg struct {
			ID    string `json:"id"`
			Error interface{} `json:"error"`
			Result struct {
				Tabs []struct {
					TabID string `json:"tab_id"`
					Label string `json:"label"`
				} `json:"tabs"`
			} `json:"result"`
		}
		if err := dec.Decode(&msg); err != nil {
			return nil, err
		}

		if msg.ID == id {
			if msg.Error != nil {
				return nil, fmt.Errorf("API error: %v", msg.Error)
			}
			labels := make(map[string]string)
			for _, t := range msg.Result.Tabs {
				labels[t.TabID] = t.Label
			}
			return labels, nil
		}
	}
}

func buildTabLabelsMap(socketPath string, agents []AgentInfo) map[string]string {
	labels := make(map[string]string)
	visitedWorkspaces := make(map[string]bool)

	for _, a := range agents {
		if a.WorkspaceID == "" || visitedWorkspaces[a.WorkspaceID] {
			continue
		}
		visitedWorkspaces[a.WorkspaceID] = true
		workspaceLabels, err := getTabLabels(socketPath, a.WorkspaceID)
		if err == nil {
			for k, v := range workspaceLabels {
				labels[k] = v
			}
		}
	}
	return labels
}

func getPaneIDs(socketPath string) ([]string, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	id := fmt.Sprintf("req_panes_%d", time.Now().UnixNano())
	req := SocketMessage{
		ID:     id,
		Method: "pane.list",
		Params: map[string]interface{}{},
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(append(reqBytes, '\n'))
	if err != nil {
		return nil, err
	}

	dec := json.NewDecoder(conn)
	for {
		var msg struct {
			ID    string `json:"id"`
			Error interface{} `json:"error"`
			Result struct {
				Panes []struct {
					PaneID string `json:"pane_id"`
				} `json:"panes"`
			} `json:"result"`
		}
		if err := dec.Decode(&msg); err != nil {
			return nil, err
		}

		if msg.ID == id {
			if msg.Error != nil {
				return nil, fmt.Errorf("API error: %v", msg.Error)
			}
			var ids []string
			for _, p := range msg.Result.Panes {
				ids = append(ids, p.PaneID)
			}
			return ids, nil
		}
	}
}

func main() {
	socketFlag := flag.String("socket", "", "Path to the Herdr UNIX domain socket (default: resolved from HERDR_SOCKET_PATH or ~/.config/herdr/herdr.sock)")
	watchFlag := flag.Bool("watch", false, "Enable watch mode to stream real-time agent status changes")
	flag.BoolVar(watchFlag, "w", false, "Enable watch mode to stream real-time agent status changes (shorthand)")
	flag.Parse()

	socketPath := getSocketPath(*socketFlag)

	// Validate socket existence
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Herdr socket not found at %q.\nIs the Herdr server running?\n", socketPath)
		os.Exit(1)
	}

	// 1. Retrieve initial agent list on a temporary connection
	agents, err := getAgents(socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying agents: %v\n", err)
		os.Exit(1)
	}

	if !*watchFlag {
		printAgentsTable(agents, socketPath)
		return
	}

	// Watch Mode: Run persistent connection loop with automatic reconnect and dynamic subscriptions
	var lastAgents []AgentInfo
	for {
		// 1. Retrieve latest agents list
		currentAgents, err := getAgents(socketPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error querying agents: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Print updated table
		printUpdatedTable(currentAgents, &lastAgents, socketPath)

		// 2. Open a persistent connection for events streaming
		conn, err := net.Dial("unix", socketPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to connect to Herdr socket: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Channels to route asynchronous responses
		responseChans := make(map[string]chan *SocketMessage)
		var mu sync.Mutex

		// Goroutine to continuously read NDJSON from the socket
		dec := json.NewDecoder(conn)
		eventChan := make(chan *SocketMessage, 100)
		doneChan := make(chan bool)

		go func() {
			for {
				var msg SocketMessage
				if err := dec.Decode(&msg); err != nil {
					break
				}

				if msg.ID != "" {
					mu.Lock()
					ch, ok := responseChans[msg.ID]
					mu.Unlock()
					if ok {
						ch <- &msg
					}
				} else if msg.Event != "" {
					eventChan <- &msg
				}
			}
			close(eventChan)
			close(doneChan)
		}()

		sendRequest := func(method string, params interface{}) (*SocketMessage, error) {
			id := fmt.Sprintf("req_%d", time.Now().UnixNano())
			req := SocketMessage{
				ID:     id,
				Method: method,
				Params: params,
			}

			reqBytes, err := json.Marshal(req)
			if err != nil {
				return nil, err
			}

			ch := make(chan *SocketMessage, 1)
			mu.Lock()
			responseChans[id] = ch
			mu.Unlock()

			defer func() {
				mu.Lock()
				delete(responseChans, id)
				mu.Unlock()
			}()

			_, err = conn.Write(append(reqBytes, '\n'))
			if err != nil {
				return nil, err
			}

			select {
			case resp := <-ch:
				if resp.Error != nil {
					return nil, fmt.Errorf("API error: %v", resp.Error)
				}
				return resp, nil
			case <-time.After(5 * time.Second):
				return nil, fmt.Errorf("request timeout")
			}
		}

		// Build subscription list
		subscriptions := []Subscription{
			{Type: "pane.agent_detected"},
		}
		
		paneIDs, err := getPaneIDs(socketPath)
		if err == nil {
			for _, pid := range paneIDs {
				subscriptions = append(subscriptions, Subscription{
					Type:   "pane.agent_status_changed",
					PaneID: pid,
				})
			}
		}

		subParams := EventsSubscribeParams{
			Subscriptions: subscriptions,
		}

		_, err = sendRequest("events.subscribe", subParams)
		if err != nil {
			conn.Close()
			time.Sleep(1 * time.Second)
			continue
		}

		// Ticker to poll and force-update every 1 second
		ticker := time.NewTicker(1 * time.Second)

		// Read events or tick until we need to reconnect
		loop:
		for {
			select {
			case event, ok := <-eventChan:
				if !ok {
					break loop
				}

				switch event.Event {
				case "pane.agent_detected":
					break loop

				case "pane.agent_status_changed":
					latestAgents, err := getAgents(socketPath)
					if err == nil {
						printUpdatedTable(latestAgents, &lastAgents, socketPath)
					}
				}

			case <-ticker.C:
				latestAgents, err := getAgents(socketPath)
				if err == nil {
					printUpdatedTable(latestAgents, &lastAgents, socketPath)
				}
			}
		}
		ticker.Stop()

		conn.Close()
		<-doneChan // Wait for reader goroutine to exit

		// Small delay before looping to prevent hot loops
		time.Sleep(100 * time.Millisecond)
	}
}

func getSocketPath(override string) string {
	if override != "" {
		return override
	}
	if path := os.Getenv("HERDR_SOCKET_PATH"); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to temporary directory if user home cannot be resolved
		return filepath.Join(os.TempDir(), "herdr.sock")
	}
	return filepath.Join(home, ".config", "herdr", "herdr.sock")
}

func printAgentsTable(agents []AgentInfo, socketPath string) {
	if len(agents) == 0 {
		fmt.Println("No active Herdr agents found.")
		return
	}

	tabLabels := buildTabLabelsMap(socketPath, agents)
	formatAgentsTable(os.Stdout, agents, tabLabels)
}

func formatAgentsTable(w io.Writer, agents []AgentInfo, tabLabels map[string]string) {
	tabPaneCounts := make(map[string]int)
	for _, a := range agents {
		if a.TabID != "" {
			tabPaneCounts[a.TabID]++
		}
	}

	type tableRow struct {
		combinedName string
		cleanStatus  string
		focused      string
		rawAgent     AgentInfo
	}

	var rows []tableRow
	maxAgentWidth := len("AGENTE")
	maxStatusWidth := len("ESTADO")

	for _, a := range agents {
		// Determine the name
		agentName := a.Name
		if a.Agent != nil {
			agentName = *a.Agent
		} else if a.DisplayAgent != nil {
			agentName = *a.DisplayAgent
		}

		tabLabel, ok := tabLabels[a.TabID]
		if !ok || tabLabel == "" {
			tabLabel = a.TabID
			if tabLabel == "" {
				tabLabel = "unknown"
			}
		}

		var combinedName string
		if a.TabID != "" && tabPaneCounts[a.TabID] > 1 {
			shortPaneID := a.PaneID
			parts := strings.Split(a.PaneID, ":")
			if len(parts) > 1 {
				shortPaneID = parts[len(parts)-1]
			}
			combinedName = fmt.Sprintf("%s-%s-%s", tabLabel, shortPaneID, agentName)
		} else {
			combinedName = fmt.Sprintf("%s-%s", tabLabel, agentName)
		}

		cleanStatus := a.AgentStatus
		if a.CustomStatus != nil && *a.CustomStatus != "" {
			cleanStatus = fmt.Sprintf("%s (%s)", a.AgentStatus, *a.CustomStatus)
		}

		focusVal := "-"
		if a.Focused {
			focusVal = "S"
		}

		if len(combinedName) > maxAgentWidth {
			maxAgentWidth = len(combinedName)
		}
		if len(cleanStatus) > maxStatusWidth {
			maxStatusWidth = len(cleanStatus)
		}

		rows = append(rows, tableRow{
			combinedName: combinedName,
			cleanStatus:  cleanStatus,
			focused:      focusVal,
			rawAgent:     a,
		})
	}

	// Dynamic column widths with safety spacing
	agentColWidth := maxAgentWidth + 3
	statusColWidth := maxStatusWidth + 3

	// Print headers
	fmt.Fprintf(w, "%-*s%-*s%s\n", agentColWidth, "AGENTE", statusColWidth, "ESTADO", "FOCO")
	fmt.Fprintf(w, "%-*s%-*s%s\n", agentColWidth, "------", statusColWidth, "------", "----")

	// Print rows
	for _, r := range rows {
		colorizedStatus := colorizeStatus(r.rawAgent.AgentStatus, r.rawAgent.CustomStatus, false)
		
		paddingLen := statusColWidth - len(r.cleanStatus)
		if paddingLen < 0 {
			paddingLen = 0
		}
		paddingSpaces := strings.Repeat(" ", paddingLen)

		fmt.Fprintf(w, "%-*s%s%s%s\n", agentColWidth, r.combinedName, colorizedStatus, paddingSpaces, r.focused)
	}
}

func agentsEqual(a, b []AgentInfo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].PaneID != b[i].PaneID ||
			a[i].AgentStatus != b[i].AgentStatus ||
			a[i].TabID != b[i].TabID ||
			a[i].WorkspaceID != b[i].WorkspaceID ||
			a[i].Focused != b[i].Focused {
			return false
		}

		nameA := a[i].Name
		if a[i].Agent != nil {
			nameA = *a[i].Agent
		} else if a[i].DisplayAgent != nil {
			nameA = *a[i].DisplayAgent
		}

		nameB := b[i].Name
		if b[i].Agent != nil {
			nameB = *b[i].Agent
		} else if b[i].DisplayAgent != nil {
			nameB = *b[i].DisplayAgent
		}

		if nameA != nameB {
			return false
		}

		statusA := ""
		if a[i].CustomStatus != nil {
			statusA = *a[i].CustomStatus
		}
		statusB := ""
		if b[i].CustomStatus != nil {
			statusB = *b[i].CustomStatus
		}
		if statusA != statusB {
			return false
		}
	}
	return true
}

func copyAgents(agents []AgentInfo) []AgentInfo {
	dst := make([]AgentInfo, len(agents))
	for i, a := range agents {
		dst[i] = a
		if a.Agent != nil {
			val := *a.Agent
			dst[i].Agent = &val
		}
		if a.CustomStatus != nil {
			val := *a.CustomStatus
			dst[i].CustomStatus = &val
		}
		if a.DisplayAgent != nil {
			val := *a.DisplayAgent
			dst[i].DisplayAgent = &val
		}
	}
	return dst
}

func printUpdatedTable(agentsList []AgentInfo, lastAgents *[]AgentInfo, socketPath string) {
	sort.Slice(agentsList, func(i, j int) bool {
		if agentsList[i].WorkspaceID != agentsList[j].WorkspaceID {
			return agentsList[i].WorkspaceID < agentsList[j].WorkspaceID
		}
		if agentsList[i].TabID != agentsList[j].TabID {
			return agentsList[i].TabID < agentsList[j].TabID
		}
		return agentsList[i].PaneID < agentsList[j].PaneID
	})

	if lastAgents != nil && agentsEqual(agentsList, *lastAgents) {
		return // No changes, avoid redrawing
	}

	// Clear screen and redraw table
	fmt.Print("\033[H\033[2J")
	printAgentsTable(agentsList, socketPath)

	if lastAgents != nil {
		*lastAgents = copyAgents(agentsList)
	}
}

func isTTY() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

func colorizeStatus(agentStatus string, customStatus *string, uppercase bool) string {
	displayStatus := agentStatus
	if uppercase {
		displayStatus = strings.ToUpper(agentStatus)
	}
	status := displayStatus
	if customStatus != nil && *customStatus != "" {
		status = fmt.Sprintf("%s (%s)", displayStatus, *customStatus)
	}

	if !isTTY() {
		return status
	}

	var colorCode string
	switch strings.ToLower(agentStatus) {
	case "working":
		colorCode = "\033[1;36m" // Bright Cyan
	case "done":
		colorCode = "\033[1;32m" // Bright Green
	case "blocked":
		colorCode = "\033[1;31m" // Bright Red
	case "idle":
		colorCode = "\033[90m" // Dark Gray (Atenuado)
	default:
		colorCode = "\033[1;33m" // Bright Yellow (unknown)
	}

	return colorCode + status + "\033[0m"
}
