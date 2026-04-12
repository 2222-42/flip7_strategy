package main

import (
	"flip7_strategy/internal/application"
	"fmt"
	"time"
)

func main() {
	fmt.Println("--- EEV Strategy Evaluation ---")
	startTime := time.Now()

	svc := application.NewSimulationService()

	// Fast test: 1500 games per risk tolerance
	// The problem in Incan Gold was that conservative play loses to risky play.
	svc.RunEEVOptimization(1500)

	elapsed := time.Since(startTime)
	fmt.Printf("\nTotal evaluation time: %s\n", elapsed)
}
