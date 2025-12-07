package application_test

import (
	"bufio"
	"strings"
	"testing"

	"flip7_strategy/internal/application"
)

// MockLogger for testing prevents nil pointer dereferences if logging is attempted
type MockLogger struct{}

func (m *MockLogger) Log(gameID, round, playerID, eventType string, details map[string]interface{}) {
}

func (m *MockLogger) Close() {
}

func TestManualModeUndoRedo(t *testing.T) {
	// Scenario:
	// 1. Start Game (2 players, Me starts)
	// 2. Play 5
	// 3. Play 6 -> Score 11
	// 4. Undo -> Back to Score 5 (Hand: [5])
	// 5. Play 7 -> Score 12 (Hand: [5, 7])
	// 6. Stay
	// 7. Player 2 stays (auto/placeholder logic doesn't run input, but setupPlayers asks for 2 players.
	//    Wait, Player 2 is AUTO strategy?
	//    In setupPlayers, I set Player 2 to ProbabilisticStrategy.
	//    But ManualGameService playRound iterates through all players.
	//    If Player 2 is AI, does Manual Mode ask for input for AI?
	//    Let's check code.
	//    playRound gets next player.
	//    "Input (0-12...)"
	//    It asks input for EVERYONE because Manual Mode assumes User controls everything (hotseat).
	//    The logic `if p.Strategy == nil` is only used for "User Controlled IDs" in save state meta.
	//    But game loop asks input for EVERYONE.
	//    So I need to provide input for Player 2 as well.

	// Input sequence:
	// "" (Save code - empty)
	// "2" (Num players)
	// "MyName" (Player 2 name - waiting for input? "Enter name for Player 2: ")
	// "1" (Start player - Me)
	// --- Round Starts ---
	// Me Turn:
	// "5"
	// "6"
	// "U" (Undo 6)
	// "7"
	// "S" (Stay)
	// Player 2 Turn:
	// "S" (Stay - Wait, must hit on first turn)
	// "10"
	// "S"
	// --- Round Ends ---
	// Check results.

	input := `
2
Bot
1
5
0
6
U
7
S
0
S
`
	// Clean up input string to ensure newlines are correct
	// Clean up input string to ensure newlines are correct
	input = "\n" + strings.TrimSpace(input) + "\n"
	// Note: We prepend a newline to satisfy the "Save Code" prompt, as TrimSpace removes the backtick's leading newline.

	reader := bufio.NewReader(strings.NewReader(input))
	service := application.NewManualGameService(reader, &MockLogger{})

	// We can't easily inspect internal state while Run() is blocking.
	// But Run() returns when game is "Completed".
	// My input sequence finishes ONE round.
	// Game completes when someone reaches 200.
	// I won't reach 200 in one round.
	// So Run() will likely block waiting for more input for Round 2.
	// AND since reader will EOF, Run() might panic or print error and exit?
	// "Error reading input. Exiting game." -> Returns.
	// This is perfect!

	service.Run()

	// Verify state after exit
	// Me: 5 + 7 = 12.
	me := service.Game.Players[0]
	if me.Name != "Me" {
		t.Errorf("Expected Player 0 to be Me, got %s", me.Name)
	}
	if me.TotalScore != 12 {
		t.Errorf("Expected Me to have 12 points (5+7 after undoing 6), got %d", me.TotalScore)
	}
}

func TestManualModeRedo(t *testing.T) {
	// Scenario:
	// 1. Me Play 5
	// 2. Bot Play 0
	// 3. Me Play 6
	// 4. Undo (Back to Me Play 6) -> Actually Undo undoes ONE step.
	//    History:
	//    - Start
	//    - Me 5
	//    - Bot 0
	//    - Me 6
	//    Undo -> Bot 0. (Me Hand [5]). Current Player Me.
	//    Redo -> Me 6. (Me Hand [5, 6]).
	// 5. Me Stay.
	// 6. Bot Stay.

	input := `
2
Bot
1
5
0
6
U
R
S
0
S
`
	// Clean up input string to ensure newlines are correct
	input = "\n" + strings.TrimSpace(input) + "\n"
	// Note: We prepend a newline to satisfy the "Save Code" prompt, as TrimSpace removes the backtick's leading newline.

	reader := bufio.NewReader(strings.NewReader(input))
	service := application.NewManualGameService(reader, &MockLogger{})
	service.Run()

	me := service.Game.Players[0]
	if me.TotalScore != 11 { // 5 + 6
		t.Errorf("Expected Me to have 11 points (Redo restored 6), got %d", me.TotalScore)
	}
}
