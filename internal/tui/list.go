package tui

import (
	"fmt"
	"bubblecal/internal/config"
	"bubblecal/internal/model"
	"bubblecal/internal/storage"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ListViewModel represents the list/agenda view
type ListViewModel struct {
	selectedDate  *time.Time
	styles        *Styles
	width         int
	height        int
	config        *config.Config
	selectedIndex int
	scrollOffset  int
	daysToShow    int // Number of days to display
	events        map[string][]*model.Event // Events grouped by date
	dateOrder     []time.Time // Ordered list of dates
	flatEvents    []EventWithDate // Flattened list for navigation
	// Jump mode
	jumpMode      bool
	jumpKeys      []string
}

// EventWithDate wraps an event with its date for list navigation
type EventWithDate struct {
	Date  time.Time
	Event *model.Event
}

// NewListViewModel creates a new list view model
func NewListViewModel(selectedDate *time.Time, styles *Styles, config *config.Config) *ListViewModel {
	return &ListViewModel{
		selectedDate:  selectedDate,
		styles:        styles,
		config:        config,
		selectedIndex: 0,
		scrollOffset:  0,
		daysToShow:    30, // Show 30 days by default
		events:        make(map[string][]*model.Event),
	}
}

func (l *ListViewModel) SetSize(width, height int) {
	l.width = width
	l.height = height
}

func (l *ListViewModel) LoadEvents() {
	l.events = make(map[string][]*model.Event)
	l.dateOrder = []time.Time{}
	l.flatEvents = []EventWithDate{}
	
	// Start from a week ago to show recent past events too
	startDate := time.Now().AddDate(0, 0, -7)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	
	// Load events for past week + next N days
	totalDays := l.daysToShow + 7
	for i := 0; i < totalDays; i++ {
		date := startDate.AddDate(0, 0, i)
		dateKey := date.Format("2006-01-02")
		
		events, _ := storage.LoadDayEvents(date)
		// Only show dates with events
		if len(events) > 0 {
			l.events[dateKey] = events
			l.dateOrder = append(l.dateOrder, date)
			
			// Add to flat list for navigation
			for _, evt := range events {
				l.flatEvents = append(l.flatEvents, EventWithDate{
					Date:  date,
					Event: evt,
				})
			}
		}
	}
	
	// If no events found, show a message
	if len(l.dateOrder) == 0 {
		today := time.Now()
		l.dateOrder = append(l.dateOrder, today)
		l.events[today.Format("2006-01-02")] = []*model.Event{}
	}
}

func (l *ListViewModel) MoveUp() {
	if l.selectedIndex > 0 {
		l.selectedIndex--
		l.ensureVisible()
	}
}

func (l *ListViewModel) MoveDown() {
	if l.selectedIndex < len(l.flatEvents)-1 {
		l.selectedIndex++
		l.ensureVisible()
	}
}

func (l *ListViewModel) GoToTop() {
	l.selectedIndex = 0
	l.scrollOffset = 0
}

func (l *ListViewModel) GoToBottom() {
	if len(l.flatEvents) > 0 {
		l.selectedIndex = len(l.flatEvents) - 1
		l.ensureVisible()
	}
}

func (l *ListViewModel) SetJumpMode(jumpMode bool, jumpKeys []string) {
	l.jumpMode = jumpMode
	l.jumpKeys = jumpKeys
}

func (l *ListViewModel) JumpToIndex(index int) {
	if index >= 0 && index < len(l.flatEvents) {
		l.selectedIndex = index
		l.ensureVisible()
	}
}

func (l *ListViewModel) GetEventCount() int {
	return len(l.flatEvents)
}

func (l *ListViewModel) PageUp() {
	pageSize := l.height / 3
	if pageSize < 5 {
		pageSize = 5
	}
	l.selectedIndex -= pageSize
	if l.selectedIndex < 0 {
		l.selectedIndex = 0
	}
	l.ensureVisible()
}

func (l *ListViewModel) PageDown() {
	pageSize := l.height / 3
	if pageSize < 5 {
		pageSize = 5
	}
	l.selectedIndex += pageSize
	if l.selectedIndex >= len(l.flatEvents) {
		l.selectedIndex = len(l.flatEvents) - 1
	}
	if l.selectedIndex < 0 {
		l.selectedIndex = 0
	}
	l.ensureVisible()
}

func (l *ListViewModel) GetSelectedEvent() *EventWithDate {
	if l.selectedIndex >= 0 && l.selectedIndex < len(l.flatEvents) {
		return &l.flatEvents[l.selectedIndex]
	}
	return nil
}

func (l *ListViewModel) ensureVisible() {
	// Build the line index for each event
	currentLine := 0
	targetLine := -1
	eventIndex := 0
	
	for i, date := range l.dateOrder {
		dateKey := date.Format("2006-01-02")
		events := l.events[dateKey]
		
		// Date header takes one line
		currentLine++
		
		if len(events) == 0 {
			// "No upcoming events" message
			currentLine++
		} else {
			// Each event takes one line
			for range events {
				if eventIndex == l.selectedIndex {
					targetLine = currentLine
				}
				currentLine++
				eventIndex++
			}
		}
		
		// Spacing between days (except last)
		if i < len(l.dateOrder)-1 {
			currentLine++
		}
	}
	
	if targetLine == -1 {
		return // Selected index not found
	}
	
	visibleHeight := l.height - 4
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	
	// Adjust scroll to keep selected line visible with some context
	if targetLine < l.scrollOffset {
		l.scrollOffset = targetLine - 1 // Show header above if possible
	} else if targetLine >= l.scrollOffset + visibleHeight {
		l.scrollOffset = targetLine - visibleHeight + 2 // Keep some context below
	}
	
	if l.scrollOffset < 0 {
		l.scrollOffset = 0
	}
}

func (l *ListViewModel) View() string {
	if l.width == 0 || l.height == 0 {
		return ""
	}
	
	// Reload events
	l.LoadEvents()
	
	var allLines []string
	eventIndex := 0
	
	// First, build ALL lines (not just visible ones)
	for _, date := range l.dateOrder {
		dateKey := date.Format("2006-01-02")
		events := l.events[dateKey]
		
		// Date header
		allLines = append(allLines, l.renderDateHeader(date))
		
		if len(events) == 0 {
			// Special case: no events at all in the date range
			noEventsStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				PaddingLeft(2)
			allLines = append(allLines, noEventsStyle.Render("No upcoming events"))
		} else {
			// Render each event
			for _, evt := range events {
				isSelected := eventIndex == l.selectedIndex
				line := l.renderEventLine(evt, date, isSelected)
				
				// Add jump key overlay if in jump mode
				if l.jumpMode && eventIndex < len(l.jumpKeys) {
					jumpKey := l.jumpKeys[eventIndex]
					jumpStyle := lipgloss.NewStyle().
						Background(lipgloss.Color("196")).
						Foreground(lipgloss.Color("15")).
						Bold(true).
						Padding(0, 1)
					jumpOverlay := jumpStyle.Render(jumpKey)
					line = jumpOverlay + " " + line
				}
				
				allLines = append(allLines, line)
				eventIndex++
			}
		}
		
		// Add spacing between days (only if there are more days to show)
		if date != l.dateOrder[len(l.dateOrder)-1] {
			allLines = append(allLines, "")
		}
	}
	
	// Now extract the visible portion
	visibleHeight := l.height - 4 // Account for borders and padding
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	
	var lines []string
	endLine := l.scrollOffset + visibleHeight
	
	for i := l.scrollOffset; i < endLine && i < len(allLines); i++ {
		lines = append(lines, allLines[i])
	}
	
	// Show scroll indicators in the padding area
	scrollIndicator := ""
	totalLines := len(allLines)
	if l.scrollOffset > 0 && endLine < totalLines {
		scrollIndicator = " [↑↓]"
	} else if l.scrollOffset > 0 {
		scrollIndicator = " [↑]"
	} else if endLine < totalLines {
		scrollIndicator = " [↓]"
	}
	
	// Add scroll indicator to the last line if needed
	if scrollIndicator != "" {
		indicatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("220")).
			Bold(true)
		lines = append(lines, indicatorStyle.Render(scrollIndicator))
	}
	
	// Pad to fill height
	for len(lines) < visibleHeight {
		lines = append(lines, "")
	}
	
	content := strings.Join(lines, "\n")
	
	return lipgloss.NewStyle().
		Width(l.width).
		Height(l.height).
		Render(content)
}

func (l *ListViewModel) renderDateHeader(date time.Time) string {
	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)
	yesterday := today.AddDate(0, 0, -1)
	
	var dateLabel string
	if sameDay(date, today) {
		dateLabel = "Today"
	} else if sameDay(date, tomorrow) {
		dateLabel = "Tomorrow"
	} else if sameDay(date, yesterday) {
		dateLabel = "Yesterday"
	} else {
		dateLabel = date.Format("Monday")
	}
	
	fullDate := date.Format("January 2, 2006")
	
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("245")).  // Subtle gray instead of blue
		Background(lipgloss.Color("235")).
		Width(l.width).
		Padding(0, 1)
	
	return headerStyle.Render(fmt.Sprintf("%s - %s", dateLabel, fullDate))
}

func (l *ListViewModel) renderEventLine(evt *model.Event, date time.Time, selected bool) string {
	// Get category color
	categoryColor := lipgloss.Color("15") // Default white
	if l.config != nil && evt.Category != "" {
		categoryColor = lipgloss.Color(l.config.GetCategoryColor(evt.Category))
	}
	
	// Build event time string
	var timeStr string
	if evt.IsAllDay() {
		timeStr = "All day"
	} else if evt.EndTime != "" {
		timeStr = fmt.Sprintf("%s-%s", evt.StartTime, evt.EndTime)
	} else {
		timeStr = evt.StartTime
	}
	
	// Time styling
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Width(13) // Fixed width for alignment
	
	// Title styling with category color
	titleStyle := lipgloss.NewStyle().
		Foreground(categoryColor)
	
	// Title and category
	titleStr := evt.Title
	if evt.Category != "" {
		titleStr = fmt.Sprintf("%s [%s]", evt.Title, evt.Category)
	}
	
	// Build the complete line
	line := fmt.Sprintf("%s %s", timeStyle.Render(timeStr), titleStyle.Render(titleStr))
	
	// Apply selection styling
	lineStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		Width(l.width - 2)
	
	if selected {
		lineStyle = lineStyle.
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("15")).
			Bold(true)
		line = "▶ " + line
	} else {
		line = "  " + line
	}
	
	return lineStyle.Render(line)
}