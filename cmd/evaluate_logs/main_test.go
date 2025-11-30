package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyze_EmptyRecords(t *testing.T) {
	var buf bytes.Buffer
	// Redirect stdout to buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	analyze([]LogRecord{})

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()
	if !strings.Contains(output, "Total Records: 0") {
		t.Errorf("Expected 'Total Records: 0', got: %s", output)
	}
	if !strings.Contains(output, "Total Games: 0") {
		t.Errorf("Expected 'Total Games: 0', got: %s", output)
	}
}

func TestAnalyze_MultipleRecords(t *testing.T) {
	records := []LogRecord{
		{
			GameID:    "game1",
			EventType: "GameStart",
		},
		{
			GameID:    "game1",
			EventType: "Bust",
		},
		{
			GameID:    "game1",
			EventType: "Flip7",
		},
		{
			GameID:    "game1",
			EventType: "GameEnd",
			Details: map[string]interface{}{
				"winners": []interface{}{"Alice", "Bob"},
			},
		},
		{
			GameID:    "game2",
			EventType: "GameStart",
		},
		{
			GameID:    "game2",
			EventType: "Bust",
		},
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	analyze(records)

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()

	// Check totals
	if !strings.Contains(output, "Total Records: 6") {
		t.Errorf("Expected 'Total Records: 6', got: %s", output)
	}
	if !strings.Contains(output, "Total Games: 2") {
		t.Errorf("Expected 'Total Games: 2', got: %s", output)
	}
	if !strings.Contains(output, "Total Busts: 2") {
		t.Errorf("Expected 'Total Busts: 2', got: %s", output)
	}
	if !strings.Contains(output, "Total Flip7s: 1") {
		t.Errorf("Expected 'Total Flip7s: 1', got: %s", output)
	}

	// Check wins
	if !strings.Contains(output, "- Alice: 1") {
		t.Errorf("Expected 'Alice: 1' in wins, got: %s", output)
	}
	if !strings.Contains(output, "- Bob: 1") {
		t.Errorf("Expected 'Bob: 1' in wins, got: %s", output)
	}
}

func TestMain_NoArguments(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to just program name
	os.Args = []string{"evaluate_logs"}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()
	if !strings.Contains(output, "Usage: evaluate_logs <log_file>") {
		t.Errorf("Expected usage message, got: %s", output)
	}
}

func TestMain_FileNotFound(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args with non-existent file
	os.Args = []string{"evaluate_logs", "/nonexistent/file.csv"}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()
	if !strings.Contains(output, "Failed to open file") {
		t.Errorf("Expected 'Failed to open file' error, got: %s", output)
	}
}

func TestMain_ValidCSVFile(t *testing.T) {
	// Create a temporary CSV file with test data
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")

	csvContent := `Timestamp,GameID,RoundID,PlayerID,EventType,Details
2024-01-01T12:00:00Z,game1,1,player1,GameStart,{}
2024-01-01T12:01:00Z,game1,1,player1,Bust,{}
2024-01-01T12:02:00Z,game1,2,player2,Flip7,{}
2024-01-01T12:03:00Z,game1,2,system,GameEnd,"{""winners"":[""Player1""]}"
`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args with test file
	os.Args = []string{"evaluate_logs", csvPath}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()

	// Verify output contains expected statistics
	if !strings.Contains(output, "Total Records: 4") {
		t.Errorf("Expected 'Total Records: 4', got: %s", output)
	}
	if !strings.Contains(output, "Total Games: 1") {
		t.Errorf("Expected 'Total Games: 1', got: %s", output)
	}
	if !strings.Contains(output, "Total Busts: 1") {
		t.Errorf("Expected 'Total Busts: 1', got: %s", output)
	}
	if !strings.Contains(output, "Total Flip7s: 1") {
		t.Errorf("Expected 'Total Flip7s: 1', got: %s", output)
	}
	if !strings.Contains(output, "- Player1: 1") {
		t.Errorf("Expected 'Player1: 1' in wins, got: %s", output)
	}
}

func TestMain_MalformedCSV(t *testing.T) {
	// Create a temporary CSV file with malformed data
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "malformed.csv")

	csvContent := `Timestamp,GameID,RoundID,PlayerID,EventType,Details
2024-01-01T12:00:00Z,game1,1,player1
incomplete,row
2024-01-01T12:02:00Z,game1,2,player2,Bust,{}
`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args with test file
	os.Args = []string{"evaluate_logs", csvPath}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()

	// Should show error reading malformed rows
	if !strings.Contains(output, "Error reading row") {
		t.Errorf("Expected 'Error reading row' warning, got: %s", output)
	}
	// Should still process the valid row
	if !strings.Contains(output, "Total Records: 1") {
		t.Errorf("Expected 'Total Records: 1' (only valid row), got: %s", output)
	}
}

func TestMain_InvalidJSON(t *testing.T) {
	// Create a temporary CSV file with invalid JSON in details
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "invalid_json.csv")

	csvContent := `Timestamp,GameID,RoundID,PlayerID,EventType,Details
2024-01-01T12:00:00Z,game1,1,player1,Event1,{invalid json}
2024-01-01T12:01:00Z,game1,1,player1,Event2,{}
`

	err := os.WriteFile(csvPath, []byte(csvContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test CSV: %v", err)
	}

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args with test file
	os.Args = []string{"evaluate_logs", csvPath}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()

	// Should show error for invalid JSON
	if !strings.Contains(output, "Error unmarshalling details") {
		t.Errorf("Expected JSON unmarshalling error, got: %s", output)
	}
	// Should still process both records (with empty details for invalid one)
	if !strings.Contains(output, "Total Records: 2") {
		t.Errorf("Expected 'Total Records: 2', got: %s", output)
	}
}

func TestMain_EmptyFile(t *testing.T) {
	// Create an empty CSV file
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")

	err := os.WriteFile(csvPath, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty CSV: %v", err)
	}

	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args with empty file
	os.Args = []string{"evaluate_logs", csvPath}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	main()

	w.Close()
	os.Stdout = oldStdout
	buf.ReadFrom(r)

	output := buf.String()

	// Should fail to read header
	if !strings.Contains(output, "Failed to read header") {
		t.Errorf("Expected 'Failed to read header' error, got: %s", output)
	}
}
