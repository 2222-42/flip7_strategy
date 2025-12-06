package application

import (
	"flip7_strategy/internal/domain"
	"strings"
	"testing"
)

func TestFormatCandidateOption(t *testing.T) {
	service := &ManualGameService{}

	candidate := domain.NewPlayer("Candidate", nil)
	candidate.TotalScore = 150
	candidate.CurrentHand = &domain.PlayerHand{
		RawNumberCards: []domain.NumberValue{5, 8},
	}

	t.Run("Suggested", func(t *testing.T) {
		suggested := candidate // Same pointer
		output := service.formatCandidateOption(candidate, suggested)

		if !strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output to contain '[Suggested]', got: %s", output)
		}
		if !strings.Contains(output, "Candidate") {
			t.Errorf("Expected output to contain name, got: %s", output)
		}
	})

	t.Run("NotSuggested", func(t *testing.T) {
		other := domain.NewPlayer("Other", nil)
		output := service.formatCandidateOption(candidate, other)

		if strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output NOT to contain '[Suggested]', got: %s", output)
		}
	})

	t.Run("NilSuggested", func(t *testing.T) {
		output := service.formatCandidateOption(candidate, nil)

		if strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output NOT to contain '[Suggested]' for nil suggestion, got: %s", output)
		}
		if !strings.Contains(output, "Candidate") {
			t.Errorf("Expected output to contain name, got: %s", output)
		}
	})
}
