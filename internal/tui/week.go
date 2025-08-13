package tui

import (
	"fmt"
	"simple-tui-cal/internal/storage"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// WeekViewModel represents the week view
type WeekViewModel struct {
	selectedDate *time.Time
	selectedHour *int
	styles       *Styles
	width        int
	height       int
}

// NewWeekViewModel creates a new week view model
func NewWeekViewModel(selectedDate *time.Time, selectedHour *int, styles *Styles) *WeekViewModel {
	return &WeekViewModel{
		selectedDate: selectedDate,
		selectedHour: selectedHour,
		styles:       styles,
	}
}

func (w *WeekViewModel) SetSize(width, height int) {
	w.width = width
	w.height = height
}

func (w *WeekViewModel) View() string {
	if w.width == 0 || w.height == 0 {
		return ""
	}
	
	sel := *w.selectedDate
	weekStart := sel.AddDate(0, 0, -int(sel.Weekday()))
	
	var lines []string
	
	// Calculate column width
	colWidth := (w.width - 10) / 7 // -10 for hour column
	if colWidth < 10 {
		colWidth = 10
	}
	
	// Header with days
	var headerCells []string
	headerCells = append(headerCells, lipgloss.NewStyle().Width(8).Align(lipgloss.Right).Render(""))
	
	for d := 0; d < 7; d++ {
		date := weekStart.AddDate(0, 0, d)
		label := fmt.Sprintf("%s %d", date.Weekday().String()[:3], date.Day())
		
		style := lipgloss.NewStyle().
			Width(colWidth).
			Align(lipgloss.Center).
			Bold(true)
		
		if sameDay(date, time.Now()) {
			style = style.
				Background(w.styles.TodayDate.GetBackground()).
				Foreground(w.styles.TodayDate.GetForeground())
		}
		
		headerCells = append(headerCells, style.Render(label))
	}
	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	lines = append(lines, strings.Repeat("─", w.width-4))
	
	// All-day events row
	var allDayRow []string
	
	// First, calculate the maximum height needed for all-day events
	maxAllDayHeight := 1
	var allDayContents []string
	for d := 0; d < 7; d++ {
		date := weekStart.AddDate(0, 0, d)
		content := w.getAllDayEvents(date)
		allDayContents = append(allDayContents, content)
		if content != "" {
			height := strings.Count(content, "\n") + 1
			if height > maxAllDayHeight {
				maxAllDayHeight = height
			}
		}
	}
	
	allDayLabel := lipgloss.NewStyle().
		Width(8).
		Height(maxAllDayHeight).
		Align(lipgloss.Right).
		Foreground(lipgloss.Color("240")).
		Render("All Day ")
	allDayRow = append(allDayRow, allDayLabel)
	
	for d := 0; d < 7; d++ {
		date := weekStart.AddDate(0, 0, d)
		allDayEvents := allDayContents[d]
		
		cellStyle := lipgloss.NewStyle().
			Width(colWidth).
			Height(maxAllDayHeight).
			Padding(0, 1)
		
		// Subtle background for today
		if sameDay(date, time.Now()) {
			cellStyle = cellStyle.Background(lipgloss.Color("234"))
		}
		
		// Highlight selected date's all-day events
		if sameDay(date, *w.selectedDate) {
			cellStyle = cellStyle.Foreground(w.styles.EventBadge.GetForeground())
		}
		
		allDayRow = append(allDayRow, cellStyle.Render(allDayEvents))
	}
	
	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, allDayRow...))
	lines = append(lines, strings.Repeat("─", w.width-4))
	
	// Hour rows (8:00 - 20:00)
	startHour := 8
	endHour := 20
	
	for h := startHour; h <= endHour; h++ {
		var rowCells []string
		
		// First, determine the maximum height needed for this hour row
		maxHeight := 1
		var cellContents []string
		for d := 0; d < 7; d++ {
			date := weekStart.AddDate(0, 0, d)
			content := w.getHourEvents(date, h)
			cellContents = append(cellContents, content)
			if content != "" {
				height := strings.Count(content, "\n") + 1
				if height > maxHeight {
					maxHeight = height
				}
			}
		}
		
		// Hour label - match the row height
		hourLabel := lipgloss.NewStyle().
			Width(8).
			Height(maxHeight).
			Align(lipgloss.Right).
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("%02d:00 ", h))
		rowCells = append(rowCells, hourLabel)
		
		// Day cells with consistent height
		for d := 0; d < 7; d++ {
			date := weekStart.AddDate(0, 0, d)
			cellContent := cellContents[d]
			
			cellStyle := lipgloss.NewStyle().
				Width(colWidth).
				Height(maxHeight).
				Padding(0, 1)
			
			// Subtle background for today's column
			if sameDay(date, time.Now()) {
				cellStyle = cellStyle.Background(lipgloss.Color("234"))
			}
			
			// Highlight selected cell (selected date + selected hour)
			if sameDay(date, *w.selectedDate) && h == *w.selectedHour {
				cellStyle = cellStyle.
					Background(w.styles.SelectedDate.GetBackground()).
					Foreground(w.styles.SelectedDate.GetForeground()).
					Bold(true)
			}
			
			rowCells = append(rowCells, cellStyle.Render(cellContent))
		}
		
		lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, rowCells...))
	}
	
	// Mini month at bottom
	lines = append(lines, "")
	lines = append(lines, w.renderMiniMonth())
	
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// GetSelectedHour returns the currently selected hour as a formatted string
func (w *WeekViewModel) GetSelectedHour() string {
	if *w.selectedHour >= 8 && *w.selectedHour <= 20 {
		return fmt.Sprintf("%02d:00", *w.selectedHour)
	}
	return ""  
}

func (w *WeekViewModel) getAllDayEvents(date time.Time) string {
	events, _ := storage.LoadDayEvents(date)
	
	var allDayTitles []string
	for _, evt := range events {
		if evt.IsAllDay() {
			title := evt.Title
			// Calculate max length based on column width
			colWidth := (w.width - 10) / 7
			maxLen := colWidth - 4 // Account for padding
			if len(title) > maxLen && maxLen > 3 {
				title = title[:maxLen-1] + "…"
			}
			allDayTitles = append(allDayTitles, title)
		}
	}
	
	if len(allDayTitles) > 0 {
		// Return each all-day event on its own line
		return strings.Join(allDayTitles, "\n")
	}
	return ""
}

func (w *WeekViewModel) getHourEvents(date time.Time, hour int) string {
	events, _ := storage.LoadDayEvents(date)
	
	var hourEvents []string
	for _, evt := range events {
		if evt.IsAllDay() {
			continue
		}
		
		// Check if event starts at this hour
		if strings.HasPrefix(evt.StartTime, fmt.Sprintf("%02d:", hour)) {
			title := evt.Title
			// Calculate max length based on column width
			colWidth := (w.width - 10) / 7
			maxLen := colWidth - 4 // Account for padding and bullet
			if len(title) > maxLen && maxLen > 3 {
				title = title[:maxLen-1] + "…"
			}
			hourEvents = append(hourEvents, w.styles.EventBadge.Render("●")+" "+title)
		}
	}
	
	if len(hourEvents) > 0 {
		// Return all events, each on its own line
		return strings.Join(hourEvents, "\n")
	}
	return ""
}

func (w *WeekViewModel) renderMiniMonth() string {
	sel := *w.selectedDate
	firstOfMonth := time.Date(sel.Year(), sel.Month(), 1, 0, 0, 0, 0, sel.Location())
	startWeekday := int(firstOfMonth.Weekday())
	
	// Calculate days in month
	firstOfNext := firstOfMonth.AddDate(0, 1, 0)
	daysInMonth := int(firstOfNext.Sub(firstOfMonth).Hours()/24 + 0.5)
	
	// Week bounds
	weekStart := sel.AddDate(0, 0, -int(sel.Weekday()))
	weekEnd := weekStart.AddDate(0, 0, 6)
	
	var lines []string
	
	// Header
	headers := "  S  M  T  W  T  F  S"
	lines = append(lines, lipgloss.NewStyle().Bold(true).Render(headers))
	
	// Build calendar
	var currentLine string
	dayCount := 1
	
	// Leading spaces for first week
	for i := 0; i < startWeekday; i++ {
		currentLine += "   "
	}
	
	// Days of month
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(sel.Year(), sel.Month(), day, 0, 0, 0, 0, sel.Location())
		
		style := lipgloss.NewStyle()
		
		// Highlight current week
		if !date.Before(weekStart) && !date.After(weekEnd) {
			style = style.Foreground(lipgloss.Color("196"))
		}
		
		// Highlight today
		if sameDay(date, time.Now()) {
			style = style.
				Background(w.styles.TodayDate.GetBackground()).
				Bold(true)
		}
		
		currentLine += style.Render(fmt.Sprintf("%3d", day))
		
		if (startWeekday+dayCount)%7 == 0 {
			lines = append(lines, currentLine)
			currentLine = ""
		}
		dayCount++
	}
	
	// Last line if needed
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}