package logging

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// CsvLogger implements GameLogger using a CSV file.
type CsvLogger struct {
	file   *os.File
	writer *csv.Writer
	mu     sync.Mutex
}

// NewCsvLogger creates a new CsvLogger writing to the specified path.
func NewCsvLogger(path string) (*CsvLogger, error) {
	// Open file in append mode, create if not exists
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	writer := csv.NewWriter(file)

	// Check if file is empty to write header
	stat, err := file.Stat()
	if err == nil && stat.Size() == 0 {
		header := []string{"Timestamp", "GameID", "RoundID", "PlayerID", "EventType", "Details"}
		if err := writer.Write(header); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to write header: %w", err)
		}
		writer.Flush()
	}

	return &CsvLogger{
		file:   file,
		writer: writer,
	}, nil
}

// Log records a game event to the CSV file.
func (l *CsvLogger) Log(gameID, roundID, playerID, eventType string, details map[string]interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format(time.RFC3339)
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		detailsJSON = []byte("{}") // Fallback
	}

	record := []string{
		timestamp,
		gameID,
		roundID,
		playerID,
		eventType,
		string(detailsJSON),
	}

	if err := l.writer.Write(record); err != nil {
		fmt.Printf("Error writing log: %v\n", err)
	}
	l.writer.Flush()
}

// Close closes the underlying file.
func (l *CsvLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.writer.Flush()
	l.file.Close()
}
