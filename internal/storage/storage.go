package storage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"simple-tui-cal/internal/model"
	"sort"
	"strings"
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

// LoadDayEvents loads events from a day file
func LoadDayEvents(date time.Time) ([]*model.Event, error) {
	filePath := GetDayFilePath(date)
	
	// If file doesn't exist, return empty list (no events)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []*model.Event{}, nil
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open day file: %w", err)
	}
	defer file.Close()
	
	var events []*model.Event
	scanner := bufio.NewScanner(file)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines
		if line == "" {
			continue
		}
		
		// Skip comment lines (starting with #)
		if strings.HasPrefix(line, "#") {
			continue
		}
		
		event, err := model.ParseEventLine(line)
		if err != nil {
			// Log error but continue loading other events
			fmt.Fprintf(os.Stderr, "Warning: line %d in %s: %v\n", lineNum, filePath, err)
			continue
		}
		
		events = append(events, event)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading day file: %w", err)
	}
	
	// Events should already be sorted in the file, but we can verify
	sortEvents(events)
	
	return events, nil
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

// SaveDayEvents saves events to a day file (for future use)
func SaveDayEvents(date time.Time, events []*model.Event) error {
	// Ensure directories exist
	if err := EnsureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	
	filePath := GetDayFilePath(date)
	
	// Sort events before saving
	sortEvents(events)
	
	// If no events, remove the file
	if len(events) == 0 {
		os.Remove(filePath) // Ignore error if file doesn't exist
		return nil
	}
	
	// Write to temporary file first (atomic write)
	tmpPath := filePath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	
	writer := bufio.NewWriter(file)
	for _, event := range events {
		if _, err := writer.WriteString(event.FormatEventLine() + "\n"); err != nil {
			file.Close()
			os.Remove(tmpPath)
			return fmt.Errorf("failed to write event: %w", err)
		}
	}
	
	if err := writer.Flush(); err != nil {
		file.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("failed to flush writer: %w", err)
	}
	
	if err := file.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to close file: %w", err)
	}
	
	// Atomic rename
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	
	return nil
}