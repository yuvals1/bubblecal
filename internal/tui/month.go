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
	
	// Fill leading days from previous month
	for i := 0; i < startWeekday; i++ {
		day := prevLast.Day() - (startWeekday - 1 - i)
		date := time.Date(prevLast.Year(), prevLast.Month(), day, 0, 0, 0, 0, now.Location())
		currentWeek = append(currentWeek, m.renderDayCell(date, true, cellWidth))
	}
	
	// Fill current month
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location())
		currentWeek = append(currentWeek, m.renderDayCell(date, false, cellWidth))
		
		// End of week
		if len(currentWeek) == 7 {
			lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, currentWeek...))
			currentWeek = []string{}
		}
	}
	
	// Fill trailing days from next month
	if len(currentWeek) > 0 {
		nextDay := 1
		for len(currentWeek) < 7 {
			date := firstOfNext.AddDate(0, 0, nextDay-1)
			currentWeek = append(currentWeek, m.renderDayCell(date, true, cellWidth))
			nextDay++
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

func (m *MonthViewModel) renderDayCell(date time.Time, otherMonth bool, width int) string {
	dayNum := fmt.Sprintf("%2d", date.Day())
	
	// Base style
	style := lipgloss.NewStyle().
		Width(width).
		Height(2).
		Padding(0, 1)
	
	// Today marker
	todayMarker := ""
	today := time.Now()
	if sameDay(date, today) {
		todayMarker = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(" T")
	}
	
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
	
	// Load events for badge
	events, _ := storage.LoadDayEvents(date)
	eventInfo := ""
	if len(events) > 0 {
		// Check for all-day events first
		for _, evt := range events {
			if evt.IsAllDay() {
				title := evt.Title
				maxLen := width - 6
				if len(title) > maxLen {
					title = title[:maxLen-1] + "…"
				}
				eventInfo = "\n" + m.styles.EventBadge.Render(title)
				break
			}
		}
		// If no all-day event, show count
		if eventInfo == "" {
			eventInfo = " " + m.styles.EventBadge.Render(fmt.Sprintf("●%d", len(events)))
		}
	}
	
	// Compose the cell content
	if todayMarker != "" {
		// Put today marker inline with day number
		return style.Render(dayNum + todayMarker + eventInfo)
	}
	return style.Render(dayNum + eventInfo)
}