package tui

import (
	"fmt"
	"simple-tui-cal/internal/model"
	"simple-tui-cal/internal/storage"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EventModal for creating/editing events
type EventModal struct {
	date         time.Time
	editingEvent *model.Event
	inputs       []textinput.Model
	focusIndex   int
	allDay       bool
	styles       *Styles
	width        int
	height       int
	errorMsg     string // Display error messages
}

const (
	inputTitle = iota
	inputStartTime
	inputEndTime
	inputCategory
	inputDescription
)

func NewEventModalWithTime(date time.Time, event *model.Event, defaultTime string, styles *Styles) *EventModal {
	m := NewEventModal(date, event, styles)
	// Override start time if provided and not editing
	if defaultTime != "" && event == nil && !m.allDay {
		m.inputs[inputStartTime].SetValue(defaultTime)
		// Set end time to one hour after start time
		if endTime := calculateEndTime(defaultTime); endTime != "" {
			m.inputs[inputEndTime].SetValue(endTime)
		}
	}
	return m
}

func NewEventModal(date time.Time, event *model.Event, styles *Styles) *EventModal {
	m := &EventModal{
		date:         date,
		editingEvent: event,
		styles:       styles,
		inputs:       make([]textinput.Model, 5),
	}
	
	// Initialize inputs
	for i := range m.inputs {
		m.inputs[i] = textinput.New()
	}
	
	// Title input
	m.inputs[inputTitle].Placeholder = "Event title"
	m.inputs[inputTitle].Focus()
	m.inputs[inputTitle].CharLimit = 100
	
	// Start time input
	m.inputs[inputStartTime].Placeholder = "09:00"
	m.inputs[inputStartTime].CharLimit = 5
	
	// End time input
	m.inputs[inputEndTime].Placeholder = "10:00 (optional)"
	m.inputs[inputEndTime].CharLimit = 5
	
	// Category input
	m.inputs[inputCategory].Placeholder = "work, personal, etc."
	m.inputs[inputCategory].CharLimit = 50
	
	// Description input
	m.inputs[inputDescription].Placeholder = "Event description (optional)"
	m.inputs[inputDescription].CharLimit = 200
	
	// Pre-fill if editing
	if event != nil {
		m.inputs[inputTitle].SetValue(event.Title)
		if !event.IsAllDay() {
			m.inputs[inputStartTime].SetValue(event.StartTime)
			m.inputs[inputEndTime].SetValue(event.EndTime)
		} else {
			m.allDay = true
		}
		m.inputs[inputCategory].SetValue(event.Category)
		m.inputs[inputDescription].SetValue(event.Description)
	} else {
		// Default to 09:00-10:00 for new events
		m.inputs[inputStartTime].SetValue("09:00")
		m.inputs[inputEndTime].SetValue("10:00")
	}
	
	return m
}

func (m *EventModal) Init() tea.Cmd {
	return textinput.Blink
}

func (m *EventModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return ModalCloseMsg(true) }
			
		case "tab", "down":
			m.focusIndex++
			if m.focusIndex >= len(m.inputs) {
				m.focusIndex = 0
			}
			m.updateFocus()
			
		case "shift+tab", "up":
			m.focusIndex--
			if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs) - 1
			}
			m.updateFocus()
			
		case "ctrl+a":
			// Toggle all-day
			m.allDay = !m.allDay
			if m.allDay {
				m.inputs[inputStartTime].SetValue("")
				m.inputs[inputEndTime].SetValue("")
			} else {
				m.inputs[inputStartTime].SetValue("09:00")
				m.inputs[inputEndTime].SetValue("10:00")
			}
			
		case "enter":
			// Save the event from any field
			if err := m.saveEvent(); err == nil {
				return m, func() tea.Msg { return ModalCloseMsg(true) }
			} else {
				m.errorMsg = err.Error()
			}
			
		default:
			// Update the focused input
			if m.focusIndex < len(m.inputs) {
				var cmd tea.Cmd
				m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)
				return m, cmd
			}
		}
	}
	
	return m, nil
}

func (m *EventModal) updateFocus() {
	for i := range m.inputs {
		if i == m.focusIndex {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *EventModal) saveEvent() error {
	title := strings.TrimSpace(m.inputs[inputTitle].Value())
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	
	event := &model.Event{
		Title:       title,
		Category:    strings.TrimSpace(m.inputs[inputCategory].Value()),
		Description: strings.TrimSpace(m.inputs[inputDescription].Value()),
	}
	
	if m.allDay {
		event.StartTime = "all-day"
		event.EndTime = ""
	} else {
		event.StartTime = strings.TrimSpace(m.inputs[inputStartTime].Value())
		event.EndTime = strings.TrimSpace(m.inputs[inputEndTime].Value())
		if event.StartTime == "" {
			return fmt.Errorf("start time required for timed events")
		}
		// Validate time format
		if !isValidTime(event.StartTime) {
			return fmt.Errorf("invalid start time format (use HH:MM)")
		}
		if event.EndTime != "" && !isValidTime(event.EndTime) {
			return fmt.Errorf("invalid end time format (use HH:MM)")
		}
	}
	
	if m.editingEvent != nil {
		// Update existing event using storage layer
		return storage.UpdateEvent(m.date, m.editingEvent, event)
	} else {
		// Add new event
		return storage.SaveEvent(m.date, event)
	}
}

func (m *EventModal) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	
	title := "New Event"
	if m.editingEvent != nil {
		title = "Edit Event"
	}
	title += fmt.Sprintf(" - %s", m.date.Format("Jan 2, 2006"))
	
	var fields []string
	
	// Title field
	fields = append(fields, "Title:")
	fields = append(fields, m.inputs[inputTitle].View())
	fields = append(fields, "")
	
	// All-day checkbox
	allDayLabel := "[ ] All Day Event (Ctrl+A to toggle)"
	if m.allDay {
		allDayLabel = "[✓] All Day Event (Ctrl+A to toggle)"
	}
	fields = append(fields, allDayLabel)
	fields = append(fields, "")
	
	// Time fields (disabled if all-day)
	if !m.allDay {
		fields = append(fields, "Start Time:")
		fields = append(fields, m.inputs[inputStartTime].View())
		fields = append(fields, "")
		
		fields = append(fields, "End Time (optional):")
		fields = append(fields, m.inputs[inputEndTime].View())
		fields = append(fields, "")
	}
	
	// Category field
	fields = append(fields, "Category:")
	fields = append(fields, m.inputs[inputCategory].View())
	fields = append(fields, "")
	
	// Description field
	fields = append(fields, "Description:")
	fields = append(fields, m.inputs[inputDescription].View())
	fields = append(fields, "")
	
	// Error message if any
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		fields = append(fields, errorStyle.Render("Error: " + m.errorMsg))
		fields = append(fields, "")
	}
	
	// Buttons - Enter saves from any field
	saveBtn := "[ Save (Enter) ]"
	cancelBtn := "[ Cancel (Esc) ]"
	
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, saveBtn, "  ", cancelBtn)
	fields = append(fields, buttons)
	
	content := lipgloss.JoinVertical(lipgloss.Left, fields...)
	
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("220")).
		Padding(1, 2).
		Width(60).
		Background(lipgloss.Color("235"))
	
	modal := modalStyle.Render(lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render(title),
		"",
		content,
	))
	
	// Center the modal
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal)
}

// DeleteModal for confirming deletion
type DeleteModal struct {
	date     time.Time
	event    *model.Event
	index    int
	styles   *Styles
	width    int
	height   int
	selected int // 0 = Delete, 1 = Cancel
}

func NewDeleteModal(date time.Time, event *model.Event, index int, styles *Styles) *DeleteModal {
	return &DeleteModal{
		date:     date,
		event:    event,
		index:    index,
		styles:   styles,
		selected: 1, // Default to Cancel
	}
}

func (m *DeleteModal) Init() tea.Cmd {
	return nil
}

func (m *DeleteModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, func() tea.Msg { return ModalCloseMsg(true) }
			
		case "left", "h":
			m.selected = 0
			
		case "right", "l":
			m.selected = 1
			
		case "enter":
			if m.selected == 0 {
				// Delete
				m.deleteEvent()
			}
			return m, func() tea.Msg { return ModalCloseMsg(true) }
		}
	}
	
	return m, nil
}

func (m *DeleteModal) deleteEvent() error {
	// Use storage layer's delete function
	return storage.DeleteEvent(m.date, m.event)
}

func (m *DeleteModal) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	
	content := fmt.Sprintf("Delete this event?\n\n%s\n%s",
		m.event.FormatEventLine(),
		m.date.Format("Monday, January 2, 2006"))
	
	deleteBtn := "[ Delete ]"
	cancelBtn := "[ Cancel ]"
	
	if m.selected == 0 {
		deleteBtn = lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Render(deleteBtn)
	} else {
		cancelBtn = lipgloss.NewStyle().
			Background(lipgloss.Color("33")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Render(cancelBtn)
	}
	
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, deleteBtn, "  ", cancelBtn)
	
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(50).
		Background(lipgloss.Color("235"))
	
	modal := modalStyle.Render(lipgloss.JoinVertical(lipgloss.Center,
		content,
		"",
		buttons,
	))
	
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal)
}

// HelpModal for showing help
type HelpModal struct {
	currentView ViewMode
	focusedPane FocusedPane
	styles      *Styles
	width       int
	height      int
}

func NewHelpModal(view ViewMode, pane FocusedPane, styles *Styles) *HelpModal {
	return &HelpModal{
		currentView: view,
		focusedPane: pane,
		styles:      styles,
	}
}

func (m *HelpModal) Init() tea.Cmd {
	return nil
}

func (m *HelpModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "?", "enter", "q":
			return m, func() tea.Msg { return ModalCloseMsg(true) }
		}
	}
	
	return m, nil
}

func (m *HelpModal) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	
	viewName := ""
	switch m.currentView {
	case MonthView:
		viewName = "Month View"
	case WeekView:
		viewName = "Week View"
	case DayView:
		viewName = "Day View"
	}
	
	var helpText []string
	helpText = append(helpText, lipgloss.NewStyle().Bold(true).Render(viewName+" Help"))
	helpText = append(helpText, "")
	
	helpText = append(helpText, lipgloss.NewStyle().Bold(true).Render("Views:"))
	helpText = append(helpText, "  Space     Cycle views (Month→Week→Day)")
	helpText = append(helpText, "")
	
	helpText = append(helpText, lipgloss.NewStyle().Bold(true).Render("Navigation:"))
	switch m.currentView {
	case MonthView:
		helpText = append(helpText, "  h/l ←/→   Previous/next day")
		helpText = append(helpText, "  j/k ↑/↓   Previous/next week")
		helpText = append(helpText, "  Ctrl+U/D  Previous/next month")
	case WeekView:
		helpText = append(helpText, "  h/l ←/→   Previous/next day")
		helpText = append(helpText, "  j/k ↑/↓   Move between hours")
		helpText = append(helpText, "  Ctrl+U/D  Previous/next week")
		helpText = append(helpText, "  m         Toggle mini-month view")
	case DayView:
		helpText = append(helpText, "  h/l ←/→   Previous/next day")
		helpText = append(helpText, "  j/k ↑/↓   Move between hours")
		helpText = append(helpText, "  Ctrl+U/D  Previous/next day")
	}
	helpText = append(helpText, "  Tab       Toggle focus (calendar ↔ agenda)")
	helpText = append(helpText, "  t or .    Go to today (current hour in Week/Day)")
	helpText = append(helpText, "")
	
	helpText = append(helpText, lipgloss.NewStyle().Bold(true).Render("Events:"))
	helpText = append(helpText, "  a         Add event")
	if m.focusedPane == AgendaPane {
		helpText = append(helpText, "  e         Edit selected event")
		helpText = append(helpText, "  d         Delete selected event")
	} else {
		helpText = append(helpText, "  (Focus agenda to edit/delete)")
	}
	helpText = append(helpText, "")
	
	helpText = append(helpText, lipgloss.NewStyle().Bold(true).Render("General:"))
	helpText = append(helpText, "  p         Toggle agenda position (right/bottom)")
	helpText = append(helpText, "  s         Cycle through themes")
	helpText = append(helpText, "  ?         Help")
	helpText = append(helpText, "  q         Quit")
	helpText = append(helpText, "")
	helpText = append(helpText, "Press any key to close")
	
	content := lipgloss.JoinVertical(lipgloss.Left, helpText...)
	
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("33")).
		Padding(1, 2).
		Width(50).
		Background(lipgloss.Color("235"))
	
	modal := modalStyle.Render(content)
	
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal)
}

// Helper function to validate time format
func isValidTime(timeStr string) bool {
	if len(timeStr) != 5 {
		return false
	}
	if timeStr[2] != ':' {
		return false
	}
	var hour, min int
	if _, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &min); err != nil {
		return false
	}
	return hour >= 0 && hour <= 23 && min >= 0 && min <= 59
}

// Helper function to calculate end time (one hour after start)
func calculateEndTime(startTime string) string {
	if !isValidTime(startTime) {
		return ""
	}
	var hour, min int
	fmt.Sscanf(startTime, "%d:%d", &hour, &min)
	hour++
	if hour > 23 {
		hour = 23 // Cap at 23:00
	}
	return fmt.Sprintf("%02d:%02d", hour, min)
}