package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"simple-tui-cal/internal/model"
	"sort"
	"time"
)

// GetCalendarDir returns the base directory for calendar data
func GetCalendarDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".simple-tui-cal")
}

// GetDaysDir returns the directory where day files are stored
func GetDaysDir() string {
	return filepath.Join(GetCalendarDir(), "days")
}

// GetDayFilePath returns the file path for a specific date
func GetDayFilePath(date time.Time) string {
	filename := date.Format("2006-01-02")
	return filepath.Join(GetDaysDir(), filename)
}

// EnsureDirectories creates the necessary directories if they don't exist
func EnsureDirectories() error {
	daysDir := GetDaysDir()
	return os.MkdirAll(daysDir, 0755)
}

// LoadDayEvents loads events from a day directory (new format)
func LoadDayEvents(date time.Time) ([]*model.Event, error) {
	return LoadDayEventsNew(date)
}

// sortEvents sorts events by time (all-day events go to the end)
func sortEvents(events []*model.Event) {
	sort.Slice(events, func(i, j int) bool {
		// All-day events go to the end
		if events[i].IsAllDay() && events[j].IsAllDay() {
			return events[i].Title < events[j].Title // sort all-day by title
		}
		if events[i].IsAllDay() {
			return false // i goes after j
		}
		if events[j].IsAllDay() {
			return true // i goes before j
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

// SaveDayEvents saves events to day directory (new format)
// This is a compatibility layer that saves all events at once
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
		if err := SaveEventNew(date, event); err != nil {
			return fmt.Errorf("failed to save event: %w", err)
		}
	}
	
	return nil
}