package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"bubblecal/internal/model"
	"sort"
	"time"
)

// GetCalendarDir returns the base directory for calendar data
func GetCalendarDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".bubblecal")
}

// GetDaysDir returns the directory where day directories are stored
func GetDaysDir() string {
	return filepath.Join(GetCalendarDir(), "days")
}

// GetDayDirPath returns the directory path for a specific date
func GetDayDirPath(date time.Time) string {
	dayDir := date.Format("2006-01-02")
	return filepath.Join(GetDaysDir(), dayDir)
}

// LoadDayEvents loads events from a day directory
func LoadDayEvents(date time.Time) ([]*model.Event, error) {
	dirPath := GetDayDirPath(date)
	
	// If directory doesn't exist, return empty list (no events)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return []*model.Event{}, nil
	}
	
	// Read all files in the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read day directory: %w", err)
	}
	
	var events []*model.Event
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		filename := entry.Name()
		filePath := filepath.Join(dirPath, filename)
		
		// Read file content
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to read %s: %v\n", filename, err)
			continue
		}
		
		// Parse event from filename and content
		event, err := model.ParseEventFromFilename(filename, string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", filename, err)
			continue
		}
		
		events = append(events, event)
	}
	
	// Sort events
	sortEvents(events)
	
	return events, nil
}

// SaveDayEvents saves all events for a day (compatibility layer)
// This clears existing events and saves all provided events
func SaveDayEvents(date time.Time, events []*model.Event) error {
	// First, clear existing events for this day
	dirPath := GetDayDirPath(date)
	if _, err := os.Stat(dirPath); err == nil {
		// Directory exists, remove it to clear all events
		if err := os.RemoveAll(dirPath); err != nil {
			return fmt.Errorf("failed to clear existing events: %w", err)
		}
	}
	
	// Save each event
	for _, event := range events {
		if err := SaveEvent(date, event); err != nil {
			return fmt.Errorf("failed to save event: %w", err)
		}
	}
	
	return nil
}

// SaveEvent saves a single event to its own file
func SaveEvent(date time.Time, event *model.Event) error {
	// Ensure directories exist
	dirPath := GetDayDirPath(date)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create day directory: %w", err)
	}
	
	// Generate filename
	filename := event.GenerateFilename()
	filePath := filepath.Join(dirPath, filename)
	
	// Check for duplicate filename (same time and title)
	if _, err := os.Stat(filePath); err == nil {
		// File exists, add a suffix
		for i := 2; i < 100; i++ {
			altFilePath := filepath.Join(dirPath, fmt.Sprintf("%s_%d", filename, i))
			if _, err := os.Stat(altFilePath); os.IsNotExist(err) {
				filePath = altFilePath
				break
			}
		}
	}
	
	// Write event content
	content := event.FormatFileContent()
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write event file: %w", err)
	}
	
	return nil
}

// DeleteEvent deletes a single event file
func DeleteEvent(date time.Time, eventToDelete *model.Event) error {
	dirPath := GetDayDirPath(date)
	
	// Find the matching file
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read day directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		filename := entry.Name()
		filePath := filepath.Join(dirPath, filename)
		
		// Read and parse to check if it matches
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		
		event, err := model.ParseEventFromFilename(filename, string(content))
		if err != nil {
			continue
		}
		
		// Check if this is the event to delete
		if event.StartTime == eventToDelete.StartTime &&
		   event.EndTime == eventToDelete.EndTime &&
		   event.Title == eventToDelete.Title {
			// Delete the file
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete event file: %w", err)
			}
			
			// Clean up empty directory
			if entries, _ := os.ReadDir(dirPath); len(entries) == 0 {
				os.Remove(dirPath)
			}
			
			return nil
		}
	}
	
	return fmt.Errorf("event not found")
}

// UpdateEvent updates an existing event (might need to rename file)
func UpdateEvent(date time.Time, oldEvent, newEvent *model.Event) error {
	// First delete the old event
	if err := DeleteEvent(date, oldEvent); err != nil {
		return fmt.Errorf("failed to delete old event: %w", err)
	}
	
	// Then save the new event
	if err := SaveEvent(date, newEvent); err != nil {
		// Try to restore old event
		SaveEvent(date, oldEvent)
		return fmt.Errorf("failed to save updated event: %w", err)
	}
	
	return nil
}

// sortEvents sorts events by time (all-day events go first)
func sortEvents(events []*model.Event) {
	sort.Slice(events, func(i, j int) bool {
		// All-day events go first
		if events[i].IsAllDay() && events[j].IsAllDay() {
			return events[i].Title < events[j].Title // sort all-day by title
		}
		if events[i].IsAllDay() {
			return true // i goes before j
		}
		if events[j].IsAllDay() {
			return false // i goes after j
		}
		
		// Both are timed events, sort by start time
		ti, erri := events[i].GetStartTime()
		tj, errj := events[j].GetStartTime()
		
		// If parsing fails, treat as end of day
		if erri != nil {
			return false
		}
		if errj != nil {
			return true
		}
		
		return ti.Before(tj)
	})
}