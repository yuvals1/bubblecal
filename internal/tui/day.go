package tui

import (
	"fmt"
	"bubblecal/internal/storage"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// DayViewModel represents the day view
type DayViewModel struct {
	selectedDate *time.Time
	selectedHour *int
	styles       *Styles
	width        int
	height       int
}

// NewDayViewModel creates a new day view model
func NewDayViewModel(selectedDate *time.Time, selectedHour *int, styles *Styles) *DayViewModel {
	return &DayViewModel{
		selectedDate: selectedDate,
		selectedHour: selectedHour,
		styles:       styles,
	}
}

func (d *DayViewModel) SetSize(width, height int) {
	d.width = width
	d.height = height
}

func (d *DayViewModel) View() string {
	if d.width == 0 || d.height == 0 {
		return ""
	}
	
	date := *d.selectedDate
	var lines []string
	
	// Date header
	headerText := date.Format("Monday, January 2, 2006")
	dateHeader := lipgloss.NewStyle().
		Width(d.width - 4).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Render(headerText)
	lines = append(lines, dateHeader)
	lines = append(lines, strings.Repeat("─", d.width-4))
	
	// Load events
	events, _ := storage.LoadDayEvents(date)
	
	// Create hour map
	hourEvents := make(map[int][]string)
	var allDayEvents []string
	
	for _, evt := range events {
		if evt.IsAllDay() {
			allDayEvents = append(allDayEvents, evt.Title)
		} else {
			var hour int
			if _, err := fmt.Sscanf(evt.StartTime, "%d:", &hour); err == nil {
				eventText := evt.Title
				if evt.EndTime != "" {
					eventText = fmt.Sprintf("%s-%s %s", evt.StartTime, evt.EndTime, evt.Title)
				} else {
					eventText = fmt.Sprintf("%s %s", evt.StartTime, evt.Title)
				}
				if evt.Category != "" {
					eventText += fmt.Sprintf(" [%s]", evt.Category)
				}
				hourEvents[hour] = append(hourEvents[hour], eventText)
			}
		}
	}
	
	// All-day events
	if len(allDayEvents) > 0 {
		// Calculate height for all-day section
		allDayHeight := len(allDayEvents)
		if allDayHeight < 1 {
			allDayHeight = 1
		}
		
		allDayLabel := lipgloss.NewStyle().
			Width(10).
			Height(allDayHeight).
			Align(lipgloss.Right).
			Foreground(lipgloss.Color("240")).
			Render("All Day ")
		
		allDayContent := lipgloss.NewStyle().
			Foreground(d.styles.EventBadge.GetForeground()).
			Width(d.width - 14).
			Height(allDayHeight).
			Render(strings.Join(allDayEvents, "\n"))
		
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, allDayLabel, allDayContent))
		lines = append(lines, strings.Repeat("─", d.width-4))
	}
	
	// Calculate dynamic hour range based on events
	startHour := 6  // Default minimum
	endHour := 22   // Default maximum
	
	// Check if any events fall outside the default range
	for hour := range hourEvents {
		if hour < startHour {
			startHour = hour
		}
		if hour > endHour {
			endHour = hour
		}
	}
	
	// Ensure we don't go beyond reasonable bounds
	if startHour < 0 {
		startHour = 0
	}
	if endHour > 23 {
		endHour = 23
	}
	
	currentHour := time.Now().Hour()
	isToday := sameDay(date, time.Now())
	
	// Just show all hours - let the content flow naturally
	for h := startHour; h <= endHour; h++ {
		// Calculate row height based on number of events
		rowHeight := 1
		eventsText := ""
		if events, ok := hourEvents[h]; ok {
			// Display each event on its own line
			eventsText = strings.Join(events, "\n")
			rowHeight = len(events)
			if rowHeight < 1 {
				rowHeight = 1
			}
		}
		
		hourStyle := lipgloss.NewStyle().
			Width(10).
			Height(rowHeight).
			Align(lipgloss.Right).
			Foreground(lipgloss.Color("240"))
		
		eventStyle := lipgloss.NewStyle().
			Height(rowHeight)
		
		// Highlight selected hour
		if h == *d.selectedHour {
			hourStyle = hourStyle.
				Background(d.styles.SelectedDate.GetBackground()).
				Foreground(d.styles.SelectedDate.GetForeground()).
				Bold(true)
			eventStyle = eventStyle.
				Background(d.styles.SelectedDate.GetBackground()).
				Foreground(d.styles.SelectedDate.GetForeground())
		} else if isToday && h == currentHour {
			// Show current hour if today (but not selected)
			hourStyle = hourStyle.
				Background(d.styles.TodayDate.GetBackground()).
				Foreground(d.styles.TodayDate.GetForeground())
			eventStyle = eventStyle.
				Background(d.styles.TodayDate.GetBackground()).
				Foreground(d.styles.TodayDate.GetForeground())
		}
		
		hourLabel := hourStyle.Render(fmt.Sprintf("%02d:00 ", h))
		
		eventContent := eventStyle.
			Width(d.width - 14).
			Render(eventsText)
		
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, hourLabel, eventContent))
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// GetSelectedHour returns the currently selected hour as a formatted string
func (d *DayViewModel) GetSelectedHour() string {
	if *d.selectedHour >= 6 && *d.selectedHour <= 22 {
		return fmt.Sprintf("%02d:00", *d.selectedHour)
	}
	return ""
}