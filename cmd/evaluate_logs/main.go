package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type LogRecord struct {
	Timestamp string
	GameID    string
	RoundID   string
	PlayerID  string
	EventType string
	Details   map[string]interface{}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: evaluate_logs <log_file>")
		return
	}

	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open file: %v\n", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header
	_, err = reader.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read header: %v\n", err)
		return
	}

	var records []LogRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading row: %v\n", err)
			continue
		}

		// Validate row has at least 5 fields before accessing
		if len(row) < 5 {
			fmt.Fprintf(os.Stderr, "Skipping malformed row (expected at least 5 fields, got %d): %v\n", len(row), row)
			continue
		}

		var details map[string]interface{}
		if len(row) > 5 {
			if err := json.Unmarshal([]byte(row[5]), &details); err != nil {
				fmt.Fprintf(os.Stderr, "Error unmarshalling details for row: %v\n\tJSON: %s\n\tError: %v\n", row, row[5], err)
				details = make(map[string]interface{})
			}
		}

		records = append(records, LogRecord{
			Timestamp: row[0],
			GameID:    row[1],
			RoundID:   row[2],
			PlayerID:  row[3],
			EventType: row[4],
			Details:   details,
		})
	}

	analyze(records)
}

func analyze(records []LogRecord) {
	fmt.Printf("Total Records: %d\n", len(records))

	games := make(map[string]bool)
	playerWins := make(map[string]int)
	busts := 0
	flips := 0

	for _, r := range records {
		games[r.GameID] = true

		if r.EventType == "GameEnd" {
			if winners, ok := r.Details["winners"].([]interface{}); ok {
				for _, w := range winners {
					if name, ok := w.(string); ok {
						playerWins[name]++
					}
				}
			}
		}

		if r.EventType == "Bust" {
			busts++
		}

		if r.EventType == "Flip7" {
			flips++
		}
	}

	fmt.Printf("Total Games: %d\n", len(games))
	fmt.Printf("Total Busts: %d\n", busts)
	fmt.Printf("Total Flip7s: %d\n", flips)

	fmt.Println("\nWins by Player:")
	for p, w := range playerWins {
		fmt.Printf("- %s: %d\n", p, w)
	}
}
