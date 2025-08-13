package main

import (
	"fmt"
	"log"
	"simple-tui-cal/internal/model"
	"simple-tui-cal/internal/storage"
	"time"
)

func main() {
	fmt.Println("Testing CRUD Operations")
	fmt.Println("=======================\n")

	testDate := time.Date(2025, 8, 20, 0, 0, 0, 0, time.Local)
	fmt.Printf("Test date: %s\n\n", testDate.Format("2006-01-02"))

	// 1. CREATE - Add test events
	fmt.Println("1. CREATE - Adding test events...")
	testEvents := []*model.Event{
		{StartTime: "09:00", EndTime: "10:00", Title: "Original Meeting", Categories: []string{"work"}},
		{StartTime: "14:00", EndTime: "", Title: "Appointment", Categories: []string{"personal"}},
		{StartTime: "all-day", Title: "Holiday", Categories: []string{"personal"}},
	}
	
	if err := storage.SaveDayEvents(testDate, testEvents); err != nil {
		log.Fatalf("Failed to save events: %v", err)
	}
	fmt.Printf("   Created %d events\n\n", len(testEvents))

	// 2. READ - Load and display
	fmt.Println("2. READ - Loading events...")
	events, err := storage.LoadDayEvents(testDate)
	if err != nil {
		log.Fatalf("Failed to load events: %v", err)
	}
	for i, evt := range events {
		fmt.Printf("   [%d] %s\n", i, evt.FormatEventLine())
	}
	fmt.Println()

	// 3. UPDATE - Edit the first event
	fmt.Println("3. UPDATE - Editing first event...")
	if len(events) > 0 {
		events[0].Title = "Updated Meeting"
		events[0].EndTime = "11:00"
		events[0].Categories = append(events[0].Categories, "important")
		
		if err := storage.SaveDayEvents(testDate, events); err != nil {
			log.Fatalf("Failed to update events: %v", err)
		}
		fmt.Printf("   Updated: %s\n\n", events[0].FormatEventLine())
	}

	// 4. DELETE - Remove middle event
	fmt.Println("4. DELETE - Removing second event...")
	if len(events) > 1 {
		deletedEvent := events[1].FormatEventLine()
		events = append(events[:1], events[2:]...)
		
		if err := storage.SaveDayEvents(testDate, events); err != nil {
			log.Fatalf("Failed to delete event: %v", err)
		}
		fmt.Printf("   Deleted: %s\n\n", deletedEvent)
	}

	// 5. VERIFY - Reload and display final state
	fmt.Println("5. VERIFY - Final state...")
	finalEvents, err := storage.LoadDayEvents(testDate)
	if err != nil {
		log.Fatalf("Failed to reload events: %v", err)
	}
	for i, evt := range finalEvents {
		fmt.Printf("   [%d] %s\n", i, evt.FormatEventLine())
	}

	// Cleanup
	fmt.Println("\n6. CLEANUP - Removing test file...")
	if err := storage.SaveDayEvents(testDate, []*model.Event{}); err != nil {
		log.Fatalf("Failed to cleanup: %v", err)
	}
	fmt.Println("   Test file removed")

	fmt.Println("\n✓ All CRUD operations successful!")
}