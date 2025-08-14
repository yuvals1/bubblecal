package tui

import (
	"fmt"
	"bubblecal/internal/config"
	"bubblecal/internal/model"
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
	config        *config.Config // Added to access categories
	// Jump mode
	jumpMode      bool
	jumpKeys      []string
}

// NewAgendaViewModel creates a new agenda view model
func NewAgendaViewModel(selectedDate *time.Time, styles *Styles, config *config.Config) *AgendaViewModel {
	return &AgendaViewModel{
		selectedDate:  selectedDate,
		styles:        styles,
		selectedIndex: 0,
		config:        config,
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

func (a *AgendaViewModel) GoToTop() {
	a.selectedIndex = 0
	a.scrollOffset = 0
}

func (a *AgendaViewModel) GoToBottom() {
	if len(a.events) > 0 {
		a.selectedIndex = len(a.events) - 1
		a.ensureVisible()
	}
}

func (a *AgendaViewModel) SetJumpMode(jumpMode bool, jumpKeys []string) {
	a.jumpMode = jumpMode
	a.jumpKeys = jumpKeys
}

func (a *AgendaViewModel) JumpToIndex(index int) {
	if index >= 0 && index < len(a.events) {
		a.selectedIndex = index
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
	if len(a.events) == 0 {
		return
	}
	visibleHeight := a.height - 2 // Account for minimal padding
	if visibleHeight < 1 {
		visibleHeight = 1
	}
	if a.selectedIndex < a.scrollOffset {
		a.scrollOffset = a.selectedIndex
	} else if a.selectedIndex >= a.scrollOffset+visibleHeight {
		a.scrollOffset = a.selectedIndex - visibleHeight + 1
	}
	if a.scrollOffset < 0 {
		a.scrollOffset = 0
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
			Width(a.width - 6).
			Align(lipgloss.Center).
			Render("No events scheduled")
		lines = append(lines, noEvents)
	} else {
		visibleHeight := a.height - 2 // Account for borders/padding
		a.ensureVisible() // Make sure selected item is visible
		endIndex := a.scrollOffset + visibleHeight
		if endIndex > len(a.events) {
			endIndex = len(a.events)
		}
		
		for i := a.scrollOffset; i < endIndex; i++ {
			evt := a.events[i]
			line := a.renderEventLine(evt, i == a.selectedIndex)
			
			// Add jump key overlay if in jump mode
			if a.jumpMode && i < len(a.jumpKeys) {
				jumpKey := a.jumpKeys[i]
				jumpStyle := lipgloss.NewStyle().
					Background(lipgloss.Color("196")).
					Foreground(lipgloss.Color("15")).
					Bold(true).
					Padding(0, 1)
				jumpOverlay := jumpStyle.Render(jumpKey)
				line = jumpOverlay + " " + line
			}
			
			lines = append(lines, line)
		}
		
		// Add scroll indicators if needed
		if a.scrollOffset > 0 {
			scrollUp := lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
				Width(a.width - 6).
				Align(lipgloss.Center).
				Render("↑ more")
			lines = append([]string{scrollUp}, lines...)
		}
		if endIndex < len(a.events) {
			scrollDown := lipgloss.NewStyle().
				Foreground(lipgloss.Color("220")).
				Width(a.width - 6).
				Align(lipgloss.Center).
				Render("↓ more")
			lines = append(lines, scrollDown)
		}
	}
	
	// Pad to fill height if needed
	for len(lines) < a.height-2 {
		lines = append(lines, "")
	}
	
	content := strings.Join(lines, "\n")
	
	return lipgloss.NewStyle().
		Width(a.width - 2).
		Padding(0, 1).
		Render(content)
}

func (a *AgendaViewModel) renderEventLine(evt *model.Event, selected bool) string {
	// Get category color
	categoryColor := lipgloss.Color("15") // Default white
	if a.config != nil && evt.Category != "" {
		categoryColor = lipgloss.Color(a.config.GetCategoryColor(evt.Category))
	}
	
	var label string
	if evt.IsAllDay() {
		// All day event
		allDayText := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Render("All day")
		titleText := lipgloss.NewStyle().
			Foreground(categoryColor).
			Render(evt.Title)
		label = fmt.Sprintf("%s %s", allDayText, titleText)
	} else {
		// Timed event
		var timeStr string
		if evt.EndTime != "" {
			timeStr = fmt.Sprintf("%s-%s", evt.StartTime, evt.EndTime)
		} else {
			timeStr = evt.StartTime
		}
		timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		titleStyle := lipgloss.NewStyle().Foreground(categoryColor)
		label = fmt.Sprintf("%s %s", timeStyle.Render(timeStr), titleStyle.Render(evt.Title))
	}
	
	// Build the final string with selection indicator
	if selected {
		finalLabel := lipgloss.NewStyle().
			Background(lipgloss.Color("238")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Width(a.width - 4).
			Render("▶ " + label)
		return finalLabel
	}
	
	// Non-selected item
	return lipgloss.NewStyle().
		Width(a.width - 4).
		Render("  " + label)
}