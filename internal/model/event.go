package model

import (
	"fmt"
	"strings"
	"time"
)

type Event struct {
	StartTime   string // "09:00", "all-day"
	EndTime     string // "10:00", "" for single time or all-day
	Title       string
	Category    string // Single category field (was []string)
	Description string // New field for event description
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
		event.Title, event.Category = extractTitleAndCategories(remainder)
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
	event.Title, event.Category = extractTitleAndCategories(remainder)

	return event, nil
}

func extractTitleAndCategories(text string) (string, string) {
	// Look for categories in brackets at the end
	if idx := strings.LastIndex(text, "["); idx != -1 {
		title := strings.TrimSpace(text[:idx])
		catPart := text[idx:]
		
		// Remove brackets
		catPart = strings.TrimPrefix(catPart, "[")
		catPart = strings.TrimSuffix(catPart, "]")
		
		// For now, join multiple categories with comma (legacy support)
		return title, strings.TrimSpace(catPart)
	}
	
	// No categories found
	return strings.TrimSpace(text), ""
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
	
	if e.Category != "" {
		result += fmt.Sprintf(" [%s]", e.Category)
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

// GenerateFilename creates a filename for this event
// Format: "HHMM-HHMM-title" or "allday-title"
func (e *Event) GenerateFilename() string {
	// Sanitize title for filename (replace spaces and special chars)
	safeTitle := strings.ReplaceAll(e.Title, " ", "_")
	safeTitle = strings.ReplaceAll(safeTitle, "/", "-")
	safeTitle = strings.ReplaceAll(safeTitle, ":", "-")
	safeTitle = strings.ReplaceAll(safeTitle, "\\", "-")
	safeTitle = strings.ReplaceAll(safeTitle, "?", "")
	safeTitle = strings.ReplaceAll(safeTitle, "*", "")
	safeTitle = strings.ReplaceAll(safeTitle, "\"", "")
	safeTitle = strings.ReplaceAll(safeTitle, "<", "")
	safeTitle = strings.ReplaceAll(safeTitle, ">", "")
	safeTitle = strings.ReplaceAll(safeTitle, "|", "")
	
	if e.IsAllDay() {
		return fmt.Sprintf("allday-%s", safeTitle)
	}
	
	// Format times without colons for filename
	start := strings.ReplaceAll(e.StartTime, ":", "")
	if e.EndTime != "" {
		end := strings.ReplaceAll(e.EndTime, ":", "")
		return fmt.Sprintf("%s-%s-%s", start, end, safeTitle)
	}
	return fmt.Sprintf("%s-%s", start, safeTitle)
}

// ParseEventFromFilename parses event details from a filename and file content
func ParseEventFromFilename(filename string, content string) (*Event, error) {
	event := &Event{}
	
	// Parse category and description from content
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "category:") {
			event.Category = strings.TrimSpace(strings.TrimPrefix(line, "category:"))
		} else if strings.HasPrefix(line, "description:") {
			event.Description = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	
	// Check if it's an all-day event
	if strings.HasPrefix(filename, "allday-") {
		event.StartTime = "all-day"
		event.EndTime = ""
		titlePart := strings.TrimPrefix(filename, "allday-")
		event.Title = strings.ReplaceAll(titlePart, "_", " ")
		return event, nil
	}
	
	// Parse timed event: HHMM-HHMM-title or HHMM-title
	parts := strings.SplitN(filename, "-", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid filename format: %s", filename)
	}
	
	// First part is start time
	if len(parts[0]) == 4 {
		event.StartTime = fmt.Sprintf("%s:%s", parts[0][:2], parts[0][2:])
	} else {
		return nil, fmt.Errorf("invalid start time in filename: %s", parts[0])
	}
	
	// Check if second part is end time or title
	if len(parts[1]) == 4 && len(parts) > 2 {
		// It's an end time
		event.EndTime = fmt.Sprintf("%s:%s", parts[1][:2], parts[1][2:])
		event.Title = strings.ReplaceAll(parts[2], "_", " ")
	} else {
		// It's part of the title
		event.EndTime = ""
		titleParts := parts[1:]
		event.Title = strings.ReplaceAll(strings.Join(titleParts, "-"), "_", " ")
	}
	
	return event, nil
}

// FormatFileContent formats the event's content for saving to file
func (e *Event) FormatFileContent() string {
	return fmt.Sprintf("category:%s\ndescription:%s\n", e.Category, e.Description)
}