package main

import (
	"fmt"
	"log"
	"simple-tui-cal/internal/model"
)

func main() {
	testLines := []string{
		"09:00-10:00 Team standup [work]",
		"10:30-11:30 Client meeting [work,important]",
		"12:00-13:00 Lunch with Sarah [personal]",
		"14:00 Quick dentist check [health]",
		"18:00-19:00 Gym session [health]",
		"all-day Project deadline [work,important]",
		"15:30-16:00 Code review",  // No categories
		"all-day Birthday",          // All-day with no categories
		"",                          // Empty line (should error)
		"invalid format",            // Invalid (no time)
	}

	fmt.Println("Testing Event Parser")
	fmt.Println("====================\n")

	for i, line := range testLines {
		fmt.Printf("Test %d: %q\n", i+1, line)
		
		event, err := model.ParseEventLine(line)
		if err != nil {
			fmt.Printf("  ERROR: %v\n\n", err)
			continue
		}

		fmt.Printf("  Start:      %s\n", event.StartTime)
		fmt.Printf("  End:        %s\n", event.EndTime)
		fmt.Printf("  Title:      %s\n", event.Title)
		fmt.Printf("  Categories: %v\n", event.Categories)
		fmt.Printf("  IsAllDay:   %v\n", event.IsAllDay())
		
		// Test formatting back to string
		formatted := event.FormatEventLine()
		fmt.Printf("  Formatted:  %s\n", formatted)
		
		// Verify round-trip parsing
		reparsed, err := model.ParseEventLine(formatted)
		if err != nil {
			fmt.Printf("  ROUND-TRIP ERROR: %v\n", err)
		} else if reparsed.Title != event.Title {
			fmt.Printf("  ROUND-TRIP MISMATCH!\n")
		} else {
			fmt.Printf("  Round-trip: ✓\n")
		}
		
		fmt.Println()
	}

	// Test sorting by time
	fmt.Println("Testing Time Parsing for Sorting")
	fmt.Println("=================================\n")
	
	events := []*model.Event{
		{StartTime: "14:00", Title: "Afternoon"},
		{StartTime: "09:00", Title: "Morning"},
		{StartTime: "all-day", Title: "All day event"},
		{StartTime: "18:30", Title: "Evening"},
	}
	
	for _, e := range events {
		t, err := e.GetStartTime()
		if err != nil {
			log.Printf("Error parsing %s: %v", e.StartTime, err)
		} else if e.IsAllDay() {
			fmt.Printf("%s: (all-day - sorts last)\n", e.Title)
		} else {
			fmt.Printf("%s: %s\n", e.Title, t.Format("15:04"))
		}
	}
}