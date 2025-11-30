package logger

// GameLogger defines the interface for logging game events.
type GameLogger interface {
	// Log records a game event.
	// gameID: Unique identifier for the game session.
	// roundID: Round number or count (not a globally unique identifier).
	// playerID: Identifier of the player involved.
	// eventType: Type of event (e.g., "TurnStart", "Hit", "Stay", "Bust").
	// details: Additional context as a map (will be serialized to JSON).
	Log(gameID, roundID, playerID, eventType string, details map[string]interface{})

	// Close cleans up any resources (e.g., closes file handles).
	Close()
}
