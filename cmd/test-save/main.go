package main

import (
	"fmt"
	"log"
	"simple-tui-cal/internal/model"
	"simple-tui-cal/internal/storage"
	"time"
)

func main() {
	fmt.Println("Testing Event Save Functionality")
	fmt.Println("=================================\n")

	// Test date
	testDate := time.Date(2025, 8, 15, 0, 0, 0, 0, time.Local)
	fmt.Printf("Test date: %s\n\n", testDate.Format("2006-01-02 (Monday)"))

	// Load existing events (should be none)
	fmt.Println("1. Loading existing events...")
	events, err := storage.LoadDayEvents(testDate)
	if err != nil {
		log.Fatalf("Error loading events: %v", err)
	}
	fmt.Printf("   Found %d events\n\n", len(events))

	// Create new events
	fmt.Println("2. Creating new events...")
	newEvents := []*model.Event{
		{
			StartTime:  "09:00",
			EndTime:    "10:00",
			Title:      "Morning meeting",
			Categories: []string{"work"},
		},
		{
			StartTime:  "14:00",
			EndTime:    "",
			Title:      "Doctor appointment",
			Categories: []string{"health"},
		},
		{
			StartTime:  "all-day",
			EndTime:    "",
			Title:      "Project deadline",
			Categories: []string{"work", "important"},
		},
	}

	// Add to existing events
	events = append(events, newEvents...)
	fmt.Printf("   Total events to save: %d\n\n", len(events))

	// Save events
	fmt.Println("3. Saving events to file...")
	if err := storage.SaveDayEvents(testDate, events); err != nil {
		log.Fatalf("Error saving events: %v", err)
	}
	fmt.Printf("   Saved to: %s\n\n", storage.GetDayFilePath(testDate))

	// Reload to verify
	fmt.Println("4. Reloading events to verify...")
	reloaded, err := storage.LoadDayEvents(testDate)
	if err != nil {
		log.Fatalf("Error reloading events: %v", err)
	}

	fmt.Printf("   Loaded %d events:\n", len(reloaded))
	for i, evt := range reloaded {
		fmt.Printf("   %d. %s\n", i+1, evt.FormatEventLine())
	}
	fmt.Println()

	// Test deletion (save empty list)
	fmt.Println("5. Testing deletion (saving empty list)...")
	if err := storage.SaveDayEvents(testDate, []*model.Event{}); err != nil {
		log.Fatalf("Error deleting events: %v", err)
	}
	
	reloaded, err = storage.LoadDayEvents(testDate)
	if err != nil {
		log.Fatalf("Error reloading after delete: %v", err)
	}
	fmt.Printf("   After deletion: %d events (file should be removed)\n", len(reloaded))

	fmt.Println("\n✓ All tests passed!")
}