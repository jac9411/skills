package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"
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
	TerminalID   string            `json:"terminal_id"`
	Name         string            `json:"name"`
	Agent        *string           `json:"agent,omitempty"`
	AgentStatus  string            `json:"agent_status"`
	WorkspaceID  string            `json:"workspace_id"`
	TabID        string            `json:"tab_id"`
	PaneID       string            `json:"pane_id"`
	Focused      bool              `json:"focused"`
	CWD          string            `json:"cwd"`
	ForegroundCWD string            `json:"foreground_cwd"`
	Revision     uint64            `json:"revision"`
	CustomStatus *string           `json:"custom_status,omitempty"`
	DisplayAgent *string           `json:"display_agent,omitempty"`
	StateLabels  map[string]string `json:"state_labels,omitempty"`
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

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var msg SocketMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
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
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("connection closed before response")
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
		printAgentsTable(agents)
		return
	}

	// Watch Mode:
	fmt.Printf("Watching Herdr agents on socket %s...\n", socketPath)
	fmt.Println("\n--- INITIAL SNAPSHOT ---")
	printAgentsTable(agents)
	fmt.Println("------------------------")
	fmt.Println("Listening for live agent status changes...")

	// 2. Open a persistent connection for events streaming
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to connect to Herdr socket: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Channels to route asynchronous responses
	responseChans := make(map[string]chan *SocketMessage)
	var mu sync.Mutex

	// Goroutine to continuously read NDJSON from the socket
	scanner := bufio.NewScanner(conn)
	eventChan := make(chan *SocketMessage, 100)

	go func() {
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var msg SocketMessage
			if err := json.Unmarshal(line, &msg); err != nil {
				continue
			}

			// Route by ID if it's a response
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
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Socket connection error: %v\n", err)
		}
		close(eventChan)
	}()

	// Helper to send a request and wait for the response
	sendRequest := func(method string, params interface{}) (*SocketMessage, error) {
		id := fmt.Sprintf("req_%d", time.Now().UnixNano())
		req := SocketMessage{
			ID:     id,
			Method: method,
			Params: params,
		}

		reqBytes, err := json.Marshal(req)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
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
			return nil, fmt.Errorf("failed to write request: %w", err)
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

	// Dynamically build subscription list:
	// - Subscribe to pane.agent_detected globally
	// - Subscribe to pane.agent_status_changed for each existing agent pane
	subscriptions := []Subscription{
		{Type: "pane.agent_detected"},
	}
	for _, a := range agents {
		subscriptions = append(subscriptions, Subscription{
			Type:   "pane.agent_status_changed",
			PaneID: a.PaneID,
		})
	}

	subParams := EventsSubscribeParams{
		Subscriptions: subscriptions,
	}

	_, err = sendRequest("events.subscribe", subParams)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to subscribe to events: %v. Live updates may not stream.\n", err)
	}

	// Handle streamed events
	for event := range eventChan {
		timeStr := time.Now().Format("15:04:05")
		if event.Data == nil {
			continue
		}

		agentName := "unknown"
		if event.Data.Agent != nil {
			agentName = *event.Data.Agent
		} else if event.Data.DisplayAgent != nil {
			agentName = *event.Data.DisplayAgent
		}

		status := strings.ToUpper(event.Data.AgentStatus)
		if event.Data.CustomStatus != nil && *event.Data.CustomStatus != "" {
			status = fmt.Sprintf("%s (%s)", status, *event.Data.CustomStatus)
		}

		switch event.Event {
		case "pane.agent_status_changed":
			fmt.Printf("[%s] CHANGE: Agent %q status is now %s in Pane %s (Workspace: %s)\n",
				timeStr, agentName, status, event.Data.PaneID, event.Data.WorkspaceID)
		case "pane.agent_detected":
			fmt.Printf("[%s] DETECT: Agent %q detected in Pane %s (Workspace: %s)\n",
				timeStr, agentName, event.Data.PaneID, event.Data.WorkspaceID)
		default:
			fmt.Printf("[%s] EVENT %q: Pane %s\n", timeStr, event.Event, event.Data.PaneID)
		}
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
		// Fallback to absolute local path if user home cannot be resolved
		return "/home/jalvarez/.config/herdr/herdr.sock"
	}
	return filepath.Join(home, ".config", "herdr", "herdr.sock")
}

func printAgentsTable(agents []AgentInfo) {
	if len(agents) == 0 {
		fmt.Println("No active Herdr agents found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "AGENT\tSTATUS\tWORKSPACE\tTAB\tPANE\tFOCUS\tCWD")
	fmt.Fprintln(w, "-----\t------\t---------\t---\t----\t-----\t---")

	for _, a := range agents {
		// Determine the name
		name := a.Name
		if a.Agent != nil {
			name = *a.Agent
		}

		// Status string
		status := a.AgentStatus
		if a.CustomStatus != nil && *a.CustomStatus != "" {
			status = fmt.Sprintf("%s (%s)", status, *a.CustomStatus)
		}

		focusStr := "no"
		if a.Focused {
			focusStr = "yes"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			name,
			status,
			a.WorkspaceID,
			a.TabID,
			a.PaneID,
			focusStr,
			a.CWD,
		)
	}
	w.Flush()
}
