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
	styles       *Styles
	width        int
	height       int
}

// NewWeekViewModel creates a new week view model
func NewWeekViewModel(selectedDate *time.Time, styles *Styles) *WeekViewModel {
	return &WeekViewModel{
		selectedDate: selectedDate,
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
	
	// Header with days
	var headerCells []string
	headerCells = append(headerCells, lipgloss.NewStyle().Width(8).Render(""))
	
	for d := 0; d < 7; d++ {
		date := weekStart.AddDate(0, 0, d)
		label := fmt.Sprintf("%s %d", date.Weekday().String()[:3], date.Day())
		
		style := lipgloss.NewStyle().
			Width(12).
			Align(lipgloss.Center).
			Bold(true)
		
		if sameDay(date, time.Now()) {
			style = style.
				Background(w.styles.TodayDate.GetBackground()).
				Foreground(w.styles.TodayDate.GetForeground())
			label += " T"
		}
		
		headerCells = append(headerCells, style.Render(label))
	}
	lines = append(lines, lipgloss.JoinHorizontal(lipgloss.Top, headerCells...))
	lines = append(lines, strings.Repeat("─", w.width-4))
	
	// Hour rows (8:00 - 20:00)
	startHour := 8
	endHour := 20
	
	for h := startHour; h <= endHour; h++ {
		var rowCells []string
		
		// Hour label
		hourLabel := lipgloss.NewStyle().
			Width(8).
			Align(lipgloss.Right).
			Foreground(lipgloss.Color("240")).
			Render(fmt.Sprintf("%02d:00 ", h))
		rowCells = append(rowCells, hourLabel)
		
		// Day cells
		for d := 0; d < 7; d++ {
			date := weekStart.AddDate(0, 0, d)
			cellContent := w.getHourEvents(date, h)
			
			cellStyle := lipgloss.NewStyle().
				Width(12).
				Height(1).
				Padding(0, 1)
			
			if sameDay(date, time.Now()) {
				cellStyle = cellStyle.Background(lipgloss.Color("234"))
			}
			
			if sameDay(date, *w.selectedDate) && h == 12 {
				cellStyle = cellStyle.
					Background(w.styles.SelectedDate.GetBackground()).
					Foreground(w.styles.SelectedDate.GetForeground())
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
			if len(title) > 10 {
				title = title[:9] + "…"
			}
			hourEvents = append(hourEvents, w.styles.EventBadge.Render("●")+" "+title)
		}
	}
	
	if len(hourEvents) > 0 {
		return hourEvents[0] // Show first event only due to space
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