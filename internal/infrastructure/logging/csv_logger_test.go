package logging_test

import (
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"flip7_strategy/internal/infrastructure/logging"
)

func TestNewCSVLogger_CreatesFileWithHeader(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.csv")

	// Create logger
	logger, err := logging.NewCSVLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Verify file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created")
	}

	// Verify header was written
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		t.Fatalf("Failed to read header: %v", err)
	}

	expectedHeader := []string{"Timestamp", "GameID", "RoundID", "PlayerID", "EventType", "Details"}
	if len(header) != len(expectedHeader) {
		t.Errorf("Header length mismatch: got %d, want %d", len(header), len(expectedHeader))
	}
	for i, col := range expectedHeader {
		if header[i] != col {
			t.Errorf("Header column %d mismatch: got %s, want %s", i, header[i], col)
		}
	}
}

func TestNewCSVLogger_AppendsToExistingFile(t *testing.T) {
	// Create temporary file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.csv")

	// Create first logger and write a record
	logger1, err := logging.NewCSVLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create first logger: %v", err)
	}
	logger1.Log("game1", "1", "player1", "TestEvent", map[string]interface{}{"key": "value"})
	logger1.Close()

	// Create second logger (should append, not recreate header)
	logger2, err := logging.NewCSVLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create second logger: %v", err)
	}
	logger2.Log("game2", "2", "player2", "TestEvent2", map[string]interface{}{"key2": "value2"})
	logger2.Close()

	// Verify file has header + 2 records
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read all records: %v", err)
	}

	// Should have header + 2 data rows
	if len(records) != 3 {
		t.Errorf("Expected 3 rows (header + 2 data), got %d", len(records))
	}
}

func TestCSVLogger_Log_WritesCorrectFormat(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.csv")

	logger, err := logging.NewCSVLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log an event
	details := map[string]interface{}{
		"score": 42,
		"card":  "5",
	}
	logger.Log("game123", "round1", "playerA", "CardPlayed", details)
	logger.Close()

	// Read and verify
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read records: %v", err)
	}

	if len(records) != 2 { // header + 1 data row
		t.Fatalf("Expected 2 rows, got %d", len(records))
	}

	dataRow := records[1]
	if len(dataRow) != 6 {
		t.Fatalf("Expected 6 columns, got %d", len(dataRow))
	}

	// Verify timestamp format (should be RFC3339)
	_, err = time.Parse(time.RFC3339, dataRow[0])
	if err != nil {
		t.Errorf("Timestamp not in RFC3339 format: %v", err)
	}

	// Verify other fields
	if dataRow[1] != "game123" {
		t.Errorf("GameID mismatch: got %s, want game123", dataRow[1])
	}
	if dataRow[2] != "round1" {
		t.Errorf("RoundID mismatch: got %s, want round1", dataRow[2])
	}
	if dataRow[3] != "playerA" {
		t.Errorf("PlayerID mismatch: got %s, want playerA", dataRow[3])
	}
	if dataRow[4] != "CardPlayed" {
		t.Errorf("EventType mismatch: got %s, want CardPlayed", dataRow[4])
	}

	// Verify details JSON
	var parsedDetails map[string]interface{}
	err = json.Unmarshal([]byte(dataRow[5]), &parsedDetails)
	if err != nil {
		t.Errorf("Failed to parse details JSON: %v", err)
	}
	if parsedDetails["score"] != float64(42) { // JSON numbers are float64
		t.Errorf("Details score mismatch: got %v, want 42", parsedDetails["score"])
	}
	if parsedDetails["card"] != "5" {
		t.Errorf("Details card mismatch: got %v, want 5", parsedDetails["card"])
	}
}

func TestCSVLogger_Log_HandlesEmptyDetails(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.csv")

	logger, err := logging.NewCSVLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Log with nil details
	logger.Log("game1", "1", "player1", "Event1", nil)
	// Log with empty details
	logger.Log("game2", "2", "player2", "Event2", map[string]interface{}{})
	logger.Close()

	// Read and verify both records exist
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read records: %v", err)
	}

	if len(records) != 3 { // header + 2 data rows
		t.Fatalf("Expected 3 rows, got %d", len(records))
	}

	// Both should have valid JSON (null or {})
	for i := 1; i <= 2; i++ {
		jsonStr := records[i][5]
		if jsonStr != "null" && jsonStr != "{}" {
			t.Errorf("Row %d: unexpected details JSON: %s", i, jsonStr)
		}
	}
}

func TestCSVLogger_ConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.csv")

	logger, err := logging.NewCSVLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Write concurrently from multiple goroutines
	const numGoroutines = 10
	const logsPerGoroutine = 10
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Log("game1", "1", "player1", "ConcurrentEvent", map[string]interface{}{
					"goroutine": id,
					"iteration": j,
				})
			}
		}(i)
	}

	wg.Wait()
	logger.Close()

	// Verify all logs were written
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read records: %v", err)
	}

	expectedRows := 1 + (numGoroutines * logsPerGoroutine) // header + data rows
	if len(records) != expectedRows {
		t.Errorf("Expected %d rows, got %d", expectedRows, len(records))
	}
}

func TestCSVLogger_Close_HandlesError(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.csv")

	logger, err := logging.NewCSVLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Close once (should work)
	logger.Close()

	// Close again (should not panic, error is logged to stderr)
	// We can't easily test the error output, but we can ensure no panic
	logger.Close()
}

func TestNewCSVLogger_InvalidPath(t *testing.T) {
	// Try to create logger with invalid path
	_, err := logging.NewCSVLogger("/invalid/nonexistent/path/test.csv")
	if err == nil {
		t.Errorf("Expected error for invalid path, got nil")
	}
	if !strings.Contains(err.Error(), "failed to open log file") {
		t.Errorf("Expected 'failed to open log file' error, got: %v", err)
	}
}

func TestCSVLogger_Log_InvalidDetailsJSON(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.csv")

	logger, err := logging.NewCSVLogger(logPath)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create details with a value that cannot be marshaled to JSON
	// (channels are not JSON-serializable)
	invalidDetails := map[string]interface{}{
		"channel": make(chan int),
	}

	// This should not panic; it should fall back to "{}"
	logger.Log("game1", "1", "player1", "Event", invalidDetails)
	logger.Close()

	// Verify fallback was used
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Failed to read records: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("Expected 2 rows, got %d", len(records))
	}

	// Should have "{}" as fallback
	if records[1][5] != "{}" {
		t.Errorf("Expected fallback '{}', got: %s", records[1][5])
	}
}
