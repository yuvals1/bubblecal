package main

import (
	"log"
	"simple-tui-cal/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model := tui.NewModel()
	program := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := program.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}