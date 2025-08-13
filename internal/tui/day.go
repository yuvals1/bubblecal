package tui

import (
	"fmt"
	"simple-tui-cal/internal/storage"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// DayViewModel represents the day view
type DayViewModel struct {
	selectedDate *time.Time
	styles       *Styles
	width        int
	height       int
}

// NewDayViewModel creates a new day view model
func NewDayViewModel(selectedDate *time.Time, styles *Styles) *DayViewModel {
	return &DayViewModel{
		selectedDate: selectedDate,
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
	dateHeader := lipgloss.NewStyle().
		Width(d.width - 4).
		Align(lipgloss.Center).
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Render(date.Format("Monday, January 2, 2006"))
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
		allDayLabel := lipgloss.NewStyle().
			Width(10).
			Align(lipgloss.Right).
			Foreground(lipgloss.Color("240")).
			Render("All Day ")
		
		allDayContent := lipgloss.NewStyle().
			Foreground(d.styles.EventBadge.GetForeground()).
			Render(strings.Join(allDayEvents, " | "))
		
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, allDayLabel, allDayContent))
		lines = append(lines, "")
	}
	
	// Hour rows (6:00 - 22:00)
	startHour := 6
	endHour := 22
	currentHour := time.Now().Hour()
	isToday := sameDay(date, time.Now())
	
	for h := startHour; h <= endHour; h++ {
		hourStyle := lipgloss.NewStyle().
			Width(10).
			Align(lipgloss.Right).
			Foreground(lipgloss.Color("240"))
		
		eventStyle := lipgloss.NewStyle()
		
		// Highlight current hour if today
		if isToday && h == currentHour {
			hourStyle = hourStyle.
				Background(d.styles.TodayDate.GetBackground()).
				Foreground(d.styles.TodayDate.GetForeground()).
				Bold(true)
			eventStyle = eventStyle.
				Background(d.styles.TodayDate.GetBackground()).
				Foreground(d.styles.TodayDate.GetForeground())
		}
		
		hourLabel := hourStyle.Render(fmt.Sprintf("%02d:00 ", h))
		
		eventsText := ""
		if events, ok := hourEvents[h]; ok {
			eventsText = strings.Join(events, " | ")
		}
		
		eventContent := eventStyle.
			Width(d.width - 14).
			Render(eventsText)
		
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, hourLabel, eventContent))
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}