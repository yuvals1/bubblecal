package tui

import (
	"fmt"
	"simple-tui-cal/internal/model"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// AgendaViewModel represents the agenda view
type AgendaViewModel struct {
	selectedDate  *time.Time
	events        []*model.Event
	selectedIndex int
	styles        *Styles
	width         int
	height        int
	scrollOffset  int
}

// NewAgendaViewModel creates a new agenda view model
func NewAgendaViewModel(selectedDate *time.Time, styles *Styles) *AgendaViewModel {
	return &AgendaViewModel{
		selectedDate:  selectedDate,
		styles:        styles,
		selectedIndex: 0,
	}
}

func (a *AgendaViewModel) SetSize(width, height int) {
	a.width = width
	a.height = height
}

func (a *AgendaViewModel) SetEvents(events []*model.Event) {
	a.events = events
	if a.selectedIndex >= len(events) {
		a.selectedIndex = len(events) - 1
	}
	if a.selectedIndex < 0 {
		a.selectedIndex = 0
	}
}

func (a *AgendaViewModel) MoveUp() {
	if a.selectedIndex > 0 {
		a.selectedIndex--
		a.ensureVisible()
	}
}

func (a *AgendaViewModel) MoveDown() {
	if a.selectedIndex < len(a.events)-1 {
		a.selectedIndex++
		a.ensureVisible()
	}
}

func (a *AgendaViewModel) GetSelectedEvent() *model.Event {
	if a.selectedIndex >= 0 && a.selectedIndex < len(a.events) {
		return a.events[a.selectedIndex]
	}
	return nil
}

func (a *AgendaViewModel) GetSelectedIndex() int {
	return a.selectedIndex
}

func (a *AgendaViewModel) ensureVisible() {
	visibleHeight := a.height - 4 // Account for borders and padding
	if a.selectedIndex < a.scrollOffset {
		a.scrollOffset = a.selectedIndex
	} else if a.selectedIndex >= a.scrollOffset+visibleHeight {
		a.scrollOffset = a.selectedIndex - visibleHeight + 1
	}
}

func (a *AgendaViewModel) View() string {
	if a.width == 0 || a.height == 0 {
		return ""
	}
	
	var lines []string
	
	if len(a.events) == 0 {
		noEvents := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Width(a.width - 4).
			Align(lipgloss.Center).
			Render("No events scheduled")
		lines = append(lines, noEvents)
	} else {
		visibleHeight := a.height - 4
		endIndex := a.scrollOffset + visibleHeight
		if endIndex > len(a.events) {
			endIndex = len(a.events)
		}
		
		for i := a.scrollOffset; i < endIndex; i++ {
			evt := a.events[i]
			line := a.renderEventLine(evt, i == a.selectedIndex)
			lines = append(lines, line)
		}
	}
	
	// Pad to fill height
	for len(lines) < a.height-4 {
		lines = append(lines, "")
	}
	
	content := strings.Join(lines, "\n")
	
	// Add scroll indicators if needed
	if a.scrollOffset > 0 {
		content = "↑ more\n" + content
	}
	if a.scrollOffset+a.height-4 < len(a.events) {
		content = content + "\n↓ more"
	}
	
	return lipgloss.NewStyle().
		Width(a.width - 4).
		Padding(1, 2).
		Render(content)
}

func (a *AgendaViewModel) renderEventLine(evt *model.Event, selected bool) string {
	var timeStr string
	var label string
	
	if evt.IsAllDay() {
		timeStr = lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")).
			Render("All day")
		label = fmt.Sprintf("%s %s", timeStr, evt.Title)
	} else {
		if evt.EndTime != "" {
			timeStr = fmt.Sprintf("%s-%s", evt.StartTime, evt.EndTime)
		} else {
			timeStr = evt.StartTime
		}
		label = fmt.Sprintf("%s %s", timeStr, evt.Title)
	}
	
	style := lipgloss.NewStyle().Width(a.width - 6)
	
	if selected {
		style = style.
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("15")).
			Bold(true)
		label = "▶ " + label
	} else {
		label = "  " + label
	}
	
	// Truncate if too long
	maxLen := a.width - 8
	if len(label) > maxLen {
		label = label[:maxLen-1] + "…"
	}
	
	return style.Render(label)
}