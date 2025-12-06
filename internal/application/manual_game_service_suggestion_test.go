package application_test

import (
	"flip7_strategy/internal/application"
	"flip7_strategy/internal/domain"
	"strings"
	"testing"
)

func TestFormatCandidateOption(t *testing.T) {
	service := &application.ManualGameService{}

	candidate := domain.NewPlayer("Candidate", nil)
	candidate.TotalScore = 150
	candidate.CurrentHand = &domain.PlayerHand{
		RawNumberCards: []domain.NumberValue{5, 8},
	}

	t.Run("Suggested", func(t *testing.T) {
		suggested := candidate // Same pointer
		output := service.FormatCandidateOption(candidate, suggested)

		if !strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output to contain '[Suggested]', got: %s", output)
		}
		if !strings.Contains(output, "Candidate") {
			t.Errorf("Expected output to contain name, got: %s", output)
		}
	})

	t.Run("NotSuggested", func(t *testing.T) {
		other := domain.NewPlayer("Other", nil)
		output := service.FormatCandidateOption(candidate, other)

		if strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output NOT to contain '[Suggested]', got: %s", output)
		}
	})

	t.Run("NilSuggested", func(t *testing.T) {
		output := service.FormatCandidateOption(candidate, nil)

		if strings.Contains(output, "[Suggested]") {
			t.Errorf("Expected output NOT to contain '[Suggested]' for nil suggestion, got: %s", output)
		}
		if !strings.Contains(output, "Candidate") {
			t.Errorf("Expected output to contain name, got: %s", output)
		}
	})

	t.Run("NilCurrentHand", func(t *testing.T) {
		candidateWithoutHand := domain.NewPlayer("NoHand", nil)
		candidateWithoutHand.TotalScore = 100
		candidateWithoutHand.CurrentHand = nil

		output := service.FormatCandidateOption(candidateWithoutHand, nil)

		if !strings.Contains(output, "NoHand") {
			t.Errorf("Expected output to contain name, got: %s", output)
		}
		if !strings.Contains(output, "[]") {
			t.Errorf("Expected output to contain empty hand '[]', got: %s", output)
		}
	})
}
