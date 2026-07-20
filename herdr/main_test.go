package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatAgentsTable_SinglePane(t *testing.T) {
	// Set NO_COLOR=1 to ensure colorizeStatus does not output ANSI escape sequences
	t.Setenv("NO_COLOR", "1")

	agents := []AgentInfo{
		{
			Name:        "agent1",
			AgentStatus: "working",
			TabID:       "tab1",
			PaneID:      "w1:t1:p1",
			Focused:     false,
		},
	}

	tabLabels := map[string]string{
		"tab1": "MiTab",
	}

	var buf bytes.Buffer
	formatAgentsTable(&buf, agents, tabLabels)

	output := buf.String()
	expectedHeader := "AGENTE"
	if !strings.Contains(output, expectedHeader) {
		t.Errorf("Expected header to contain %q, but got:\n%s", expectedHeader, output)
	}

	// We replace tabs and multiple spaces to simplify comparison
	normalizedOutput := strings.Join(strings.Fields(output), " ")
	normalizedExpected := "AGENTE ESTADO FOCO ------ ------ ---- MiTab-agent1 working -"
	if !strings.Contains(normalizedOutput, normalizedExpected) {
		t.Errorf("Expected output to match %q, but normalized was %q", normalizedExpected, normalizedOutput)
	}
}

func TestFormatAgentsTable_MultiPane(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	agents := []AgentInfo{
		{
			Name:        "agentA",
			AgentStatus: "idle",
			TabID:       "tab1",
			PaneID:      "w1:t1:p1",
			Focused:     true,
		},
		{
			Name:        "agentB",
			AgentStatus: "done",
			TabID:       "tab1",
			PaneID:      "w1:t1:p2",
			Focused:     false,
		},
	}

	tabLabels := map[string]string{
		"tab1": "MiTab",
	}

	var buf bytes.Buffer
	formatAgentsTable(&buf, agents, tabLabels)

	normalizedOutput := strings.Join(strings.Fields(buf.String()), " ")
	expectedRowA := "MiTab-p1-agentA idle S"
	expectedRowB := "MiTab-p2-agentB done -"

	if !strings.Contains(normalizedOutput, expectedRowA) {
		t.Errorf("Expected output to contain multi-pane entry A %q, but got:\n%s", expectedRowA, normalizedOutput)
	}
	if !strings.Contains(normalizedOutput, expectedRowB) {
		t.Errorf("Expected output to contain multi-pane entry B %q, but got:\n%s", expectedRowB, normalizedOutput)
	}
}

func TestAgentsEqual_Focused(t *testing.T) {
	a := []AgentInfo{
		{
			PaneID:      "p1",
			AgentStatus: "idle",
			TabID:       "t1",
			WorkspaceID: "w1",
			Name:        "agent1",
			Focused:     false,
		},
	}

	b := []AgentInfo{
		{
			PaneID:      "p1",
			AgentStatus: "idle",
			TabID:       "t1",
			WorkspaceID: "w1",
			Name:        "agent1",
			Focused:     true, // Different focus
		},
	}

	if agentsEqual(a, b) {
		t.Error("Expected agentsEqual to return false for different Focused status, but got true")
	}

	b[0].Focused = false
	if !agentsEqual(a, b) {
		t.Error("Expected agentsEqual to return true for identical Focused status, but got false")
	}
}
