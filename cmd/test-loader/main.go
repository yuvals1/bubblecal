package main

import (
	"fmt"
	"log"
	"simple-tui-cal/internal/storage"
	"time"
)

func main() {
	fmt.Println("Testing Event File Loading")
	fmt.Println("==========================\n")

	// Test dates
	dates := []time.Time{
		time.Date(2025, 1, 13, 0, 0, 0, 0, time.Local), // Today (has events)
		time.Date(2025, 1, 14, 0, 0, 0, 0, time.Local), // Tomorrow (has events)
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local), // Day after (no file)
	}

	for _, date := range dates {
		fmt.Printf("Loading events for %s:\n", date.Format("2006-01-02 (Monday)"))
		fmt.Printf("File path: %s\n", storage.GetDayFilePath(date))
		
		events, err := storage.LoadDayEvents(date)
		if err != nil {
			log.Printf("Error loading events: %v\n", err)
			continue
		}

		if len(events) == 0 {
			fmt.Println("  No events for this day\n")
			continue
		}

		fmt.Printf("  Found %d events:\n", len(events))
		for i, event := range events {
			fmt.Printf("  %d. ", i+1)
			if event.IsAllDay() {
				fmt.Printf("[ALL DAY] ")
			} else if event.EndTime != "" {
				fmt.Printf("[%s-%s] ", event.StartTime, event.EndTime)
			} else {
				fmt.Printf("[%s] ", event.StartTime)
			}
			fmt.Printf("%s", event.Title)
			if len(event.Categories) > 0 {
				fmt.Printf(" (%v)", event.Categories)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Test the directories
	fmt.Println("Directory Information:")
	fmt.Println("======================")
	fmt.Printf("Calendar dir: %s\n", storage.GetCalendarDir())
	fmt.Printf("Days dir:     %s\n", storage.GetDaysDir())
	
	// Test creating directories (should be idempotent)
	if err := storage.EnsureDirectories(); err != nil {
		log.Printf("Error ensuring directories: %v\n", err)
	} else {
		fmt.Println("Directories verified ✓")
	}
}