package tui

import (
	"fmt"
	"simple-tui-cal/internal/storage"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// MonthViewModel represents the month view
type MonthViewModel struct {
	selectedDate *time.Time
	styles       *Styles
	width        int
	height       int
}

// NewMonthViewModel creates a new month view model
func NewMonthViewModel(selectedDate *time.Time, styles *Styles) *MonthViewModel {
	return &MonthViewModel{
		selectedDate: selectedDate,
		styles:       styles,
	}
}

func (m *MonthViewModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *MonthViewModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	
	now := *m.selectedDate
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startWeekday := int(firstOfMonth.Weekday())
	
	// Calculate days in month
	firstOfNext := firstOfMonth.AddDate(0, 1, 0)
	daysInMonth := int(firstOfNext.Sub(firstOfMonth).Hours()/24 + 0.5)
	
	// Build the calendar
	var lines []string
	
	// Calculate cell width (account for 7 days and spacing)
	cellWidth := (m.width - 6) / 7
	if cellWidth < 8 {
		cellWidth = 8
	}
	
	// Header with weekday names
	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	var headerCells []string
	for _, day := range weekdays {
		cell := lipgloss.NewStyle().
			Width(cellWidth).
			Align(lipgloss.Center).
			Bold(true).
			Render(day)
		headerCells = append(headerCells, cell)
	}
	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	lines = append(lines, strings.Repeat("─", len(lipgloss.JoinHorizontal(lipgloss.Top, headerCells...)))) // Separator
	
	// Previous month's trailing days
	prevLast := firstOfMonth.AddDate(0, 0, -1)
	
	// Build calendar grid
	var currentWeek []string
	var currentWeekDates []time.Time
	
	// Fill leading days from previous month
	for i := 0; i < startWeekday; i++ {
		day := prevLast.Day() - (startWeekday - 1 - i)
		date := time.Date(prevLast.Year(), prevLast.Month(), day, 0, 0, 0, 0, now.Location())
		currentWeekDates = append(currentWeekDates, date)
	}
	
	// Fill current month
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location())
		currentWeekDates = append(currentWeekDates, date)
		
		// End of week
		if len(currentWeekDates) == 7 {
			// Calculate max height for this week
			maxHeight := m.getMaxHeightForWeek(currentWeekDates)
			
			// Render all cells with consistent height
			for _, d := range currentWeekDates {
				isOtherMonth := d.Month() != now.Month()
				currentWeek = append(currentWeek, m.renderDayCellWithHeight(d, isOtherMonth, cellWidth, maxHeight))
			}
			
			lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, currentWeek...))
			currentWeek = []string{}
			currentWeekDates = []time.Time{}
		}
	}
	
	// Fill trailing days from next month
	if len(currentWeekDates) > 0 {
		nextDay := 1
		for len(currentWeekDates) < 7 {
			date := firstOfNext.AddDate(0, 0, nextDay-1)
			currentWeekDates = append(currentWeekDates, date)
			nextDay++
		}
		
		// Calculate max height for this week
		maxHeight := m.getMaxHeightForWeek(currentWeekDates)
		
		// Render all cells with consistent height
		for _, d := range currentWeekDates {
			isOtherMonth := d.Month() != now.Month()
			currentWeek = append(currentWeek, m.renderDayCellWithHeight(d, isOtherMonth, cellWidth, maxHeight))
		}
		
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, currentWeek...))
	}
	
	// Ensure we have 6 weeks for consistent height
	for len(lines) < 8 { // 1 header + 1 separator + 6 weeks
		emptyWeek := make([]string, 7)
		for i := 0; i < 7; i++ {
			emptyWeek[i] = lipgloss.NewStyle().Width(cellWidth).Height(2).Render("")
		}
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, emptyWeek...))
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m *MonthViewModel) getMaxHeightForWeek(dates []time.Time) int {
	maxHeight := 2 // Minimum height
	
	for _, date := range dates {
		events, _ := storage.LoadDayEvents(date)
		allDayCount := 0
		hasTimedEvents := false
		
		for _, evt := range events {
			if evt.IsAllDay() {
				allDayCount++
			} else {
				hasTimedEvents = true
			}
		}
		
		// Calculate needed height for this cell
		cellHeight := 2
		if allDayCount > 0 {
			cellHeight = 1 + allDayCount
			if hasTimedEvents {
				cellHeight++
			}
		}
		
		if cellHeight > maxHeight {
			maxHeight = cellHeight
		}
	}
	
	return maxHeight
}

func (m *MonthViewModel) renderDayCellWithHeight(date time.Time, otherMonth bool, width int, height int) string {
	dayNum := fmt.Sprintf("%2d", date.Day())
	
	// Base style with specified height
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(0, 1)
	
	today := time.Now()
	
	// Apply styling based on date properties
	if otherMonth {
		style = style.Foreground(m.styles.OtherMonth.GetForeground())
	} else if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		style = style.Foreground(m.styles.Weekend.GetForeground())
	}
	
	// Selected date
	if sameDay(date, *m.selectedDate) {
		style = style.
			Background(m.styles.SelectedDate.GetBackground()).
			Foreground(m.styles.SelectedDate.GetForeground()).
			Bold(true)
	}
	
	// Today background (only if not selected)
	if sameDay(date, today) && !sameDay(date, *m.selectedDate) {
		style = style.
			Background(m.styles.TodayDate.GetBackground()).
			Foreground(m.styles.TodayDate.GetForeground())
	}
	
	// Load events
	events, _ := storage.LoadDayEvents(date)
	eventInfo := ""
	if len(events) > 0 {
		// Collect all all-day events
		var allDayEvents []string
		timedEventCount := 0
		
		for _, evt := range events {
			if evt.IsAllDay() {
				title := evt.Title
				maxLen := width - 6
				if len(title) > maxLen && maxLen > 3 {
					title = title[:maxLen-1] + "…"
				}
				allDayEvents = append(allDayEvents, m.styles.EventBadge.Render(title))
			} else {
				timedEventCount++
			}
		}
		
		// Build event info
		if len(allDayEvents) > 0 {
			// Show all-day events on separate lines
			eventInfo = "\n" + strings.Join(allDayEvents, "\n")
			// Add timed event count if any
			if timedEventCount > 0 {
				eventInfo += "\n" + m.styles.EventBadge.Render(fmt.Sprintf("●%d", timedEventCount))
			}
		} else if timedEventCount > 0 {
			// Only timed events
			eventInfo = " " + m.styles.EventBadge.Render(fmt.Sprintf("●%d", timedEventCount))
		}
	}
	
	// Compose the cell content
	return style.Render(dayNum + eventInfo)
}

// Keep the old renderDayCell for compatibility (not used anymore)
func (m *MonthViewModel) renderDayCell(date time.Time, otherMonth bool, width int) string {
	dayNum := fmt.Sprintf("%2d", date.Day())
	
	// Load events once
	events, _ := storage.LoadDayEvents(date)
	
	// Calculate cell height based on events
	cellHeight := 2 // Minimum height
	allDayCount := 0
	hasTimedEvents := false
	
	for _, evt := range events {
		if evt.IsAllDay() {
			allDayCount++
		} else {
			hasTimedEvents = true
		}
	}
	
	// Calculate needed height: 1 for day number + 1 per all-day event + 1 for timed events indicator
	if allDayCount > 0 {
		cellHeight = 1 + allDayCount
		if hasTimedEvents {
			cellHeight++
		}
	}
	if cellHeight < 2 {
		cellHeight = 2
	}
	
	// Base style
	style := lipgloss.NewStyle().
		Width(width).
		Height(cellHeight).
		Padding(0, 1)
	
	today := time.Now()
	
	// Apply styling based on date properties
	if otherMonth {
		style = style.Foreground(m.styles.OtherMonth.GetForeground())
	} else if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		style = style.Foreground(m.styles.Weekend.GetForeground())
	}
	
	// Selected date
	if sameDay(date, *m.selectedDate) {
		style = style.
			Background(m.styles.SelectedDate.GetBackground()).
			Foreground(m.styles.SelectedDate.GetForeground()).
			Bold(true)
	}
	
	// Today background (only if not selected)
	if sameDay(date, today) && !sameDay(date, *m.selectedDate) {
		style = style.
			Background(m.styles.TodayDate.GetBackground()).
			Foreground(m.styles.TodayDate.GetForeground())
	}
	
	// Build events display
	eventInfo := ""
	if len(events) > 0 {
		// Collect all all-day events
		var allDayEvents []string
		timedEventCount := 0
		
		for _, evt := range events {
			if evt.IsAllDay() {
				title := evt.Title
				maxLen := width - 6
				if len(title) > maxLen && maxLen > 3 {
					title = title[:maxLen-1] + "…"
				}
				allDayEvents = append(allDayEvents, m.styles.EventBadge.Render(title))
			} else {
				timedEventCount++
			}
		}
		
		// Build event info
		if len(allDayEvents) > 0 {
			// Show all-day events on separate lines
			eventInfo = "\n" + strings.Join(allDayEvents, "\n")
			// Add timed event count if any
			if timedEventCount > 0 {
				eventInfo += "\n" + m.styles.EventBadge.Render(fmt.Sprintf("●%d", timedEventCount))
			}
		} else if timedEventCount > 0 {
			// Only timed events
			eventInfo = " " + m.styles.EventBadge.Render(fmt.Sprintf("●%d", timedEventCount))
		}
	}
	
	// Compose the cell content
	return style.Render(dayNum + eventInfo)
}