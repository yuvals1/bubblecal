package model

import (
	"fmt"
	"strings"
	"time"
)

type Event struct {
	StartTime  string   // "09:00", "all-day"
	EndTime    string   // "10:00", "" for single time or all-day
	Title      string
	Categories []string
}

// ParseEventLine parses a line from a day file into an Event
// Format: "09:00-10:00 Team standup [work,important]"
// Format: "14:00 Quick check [health]"
// Format: "all-day Vacation [personal]"
func ParseEventLine(line string) (*Event, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("empty line")
	}

	event := &Event{}
	
	// Check for all-day events first
	if strings.HasPrefix(line, "all-day ") {
		event.StartTime = "all-day"
		event.EndTime = ""
		remainder := strings.TrimPrefix(line, "all-day ")
		event.Title, event.Categories = extractTitleAndCategories(remainder)
		return event, nil
	}

	// Parse time (either "HH:MM-HH:MM" or "HH:MM")
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid format: no space after time")
	}

	timePart := parts[0]
	remainder := parts[1]

	// Check if it's a time range or single time
	if strings.Contains(timePart, "-") {
		times := strings.Split(timePart, "-")
		if len(times) != 2 {
			return nil, fmt.Errorf("invalid time range format")
		}
		event.StartTime = strings.TrimSpace(times[0])
		event.EndTime = strings.TrimSpace(times[1])
	} else {
		event.StartTime = timePart
		event.EndTime = ""
	}

	// Extract title and categories from remainder
	event.Title, event.Categories = extractTitleAndCategories(remainder)

	return event, nil
}

func extractTitleAndCategories(text string) (string, []string) {
	// Look for categories in brackets at the end
	if idx := strings.LastIndex(text, "["); idx != -1 {
		title := strings.TrimSpace(text[:idx])
		catPart := text[idx:]
		
		// Remove brackets
		catPart = strings.TrimPrefix(catPart, "[")
		catPart = strings.TrimSuffix(catPart, "]")
		
		// Split categories
		var categories []string
		if catPart != "" {
			for _, cat := range strings.Split(catPart, ",") {
				if trimmed := strings.TrimSpace(cat); trimmed != "" {
					categories = append(categories, trimmed)
				}
			}
		}
		
		return title, categories
	}
	
	// No categories found
	return strings.TrimSpace(text), nil
}

// FormatEventLine formats an Event back into a line for saving
func (e *Event) FormatEventLine() string {
	var timePart string
	if e.StartTime == "all-day" {
		timePart = "all-day"
	} else if e.EndTime != "" {
		timePart = fmt.Sprintf("%s-%s", e.StartTime, e.EndTime)
	} else {
		timePart = e.StartTime
	}

	result := fmt.Sprintf("%s %s", timePart, e.Title)
	
	if len(e.Categories) > 0 {
		result += fmt.Sprintf(" [%s]", strings.Join(e.Categories, ","))
	}
	
	return result
}

// IsAllDay returns true if this is an all-day event
func (e *Event) IsAllDay() bool {
	return e.StartTime == "all-day"
}

// GetStartTime parses the start time as a time.Time (for sorting)
// Returns a zero time for all-day events
func (e *Event) GetStartTime() (time.Time, error) {
	if e.IsAllDay() {
		return time.Time{}, nil
	}
	
	// Parse as HH:MM
	return time.Parse("15:04", e.StartTime)
}