package tui

import (
	"fmt"
	"bubblecal/internal/config"
	"bubblecal/internal/model"
	"bubblecal/internal/storage"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FieldType represents different field types in the modal
type FieldType int

const (
	FieldTitle FieldType = iota
	FieldAllDay
	FieldStartTime
	FieldEndTime
	FieldCategory
	FieldDescription
)

// EventModal for creating/editing events
type EventModal struct {
	date            time.Time
	editingEvent    *model.Event
	inputs          []textinput.Model
	focusedField    FieldType
	allDay          bool
	styles          *Styles
	width           int
	height          int
	errorMsg        string
	categories      []config.Category
	selectedCatIdx  int
	categoryMode    bool
}

const (
	inputTitle = iota
	inputStartTime
	inputEndTime
	inputDescription
)

func NewEventModalWithTime(date time.Time, event *model.Event, defaultTime string, styles *Styles, categories []config.Category) *EventModal {
	m := NewEventModal(date, event, styles, categories)
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

func NewEventModal(date time.Time, event *model.Event, styles *Styles, categories []config.Category) *EventModal {
	m := &EventModal{
		date:         date,
		editingEvent: event,
		styles:       styles,
		inputs:       make([]textinput.Model, 4), // Reduced from 5 to 4 (removed category input)
		categories:   categories,
		focusedField: FieldTitle, // Start with title focused
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
		// Find the category index
		for i, cat := range m.categories {
			if cat.Name == event.Category {
				m.selectedCatIdx = i
				break
			}
		}
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
			if m.categoryMode {
				m.categoryMode = false
				return m, nil
			}
			return m, func() tea.Msg { return ModalCloseMsg(true) }
			
		case "tab", "down", "j":
			return m.handleNavigation(1), nil
			
		case "shift+tab", "up", "k":
			return m.handleNavigation(-1), nil
			
		case "ctrl+s", "enter":
			newModel, cmd := m.handleAction()
			return newModel, cmd
			
		case " ":
			// Only handle space for non-text fields
			if m.focusedField == FieldAllDay {
				m.toggleAllDay()
				return m, nil
			} else if m.focusedField == FieldCategory {
				m.categoryMode = !m.categoryMode
				return m, nil
			}
			// Fall through to default for text inputs
			fallthrough
			
		default:
			// Handle text input
			if m.focusedField == FieldTitle || m.focusedField == FieldStartTime || 
			   m.focusedField == FieldEndTime || m.focusedField == FieldDescription {
				inputIdx := m.getInputIndex()
				if inputIdx >= 0 && inputIdx < len(m.inputs) {
					var cmd tea.Cmd
					m.inputs[inputIdx], cmd = m.inputs[inputIdx].Update(msg)
					return m, cmd
				}
			}
		}
	}
	
	return m, nil
}

// Helper methods for the new navigation system
func (m *EventModal) handleNavigation(direction int) *EventModal {
	if m.categoryMode {
		m.selectedCatIdx += direction
		if m.selectedCatIdx >= len(m.categories) {
			m.selectedCatIdx = 0
		} else if m.selectedCatIdx < 0 {
			m.selectedCatIdx = len(m.categories) - 1
		}
		return m
	}
	
	fields := []FieldType{FieldTitle, FieldAllDay, FieldStartTime, FieldEndTime, FieldCategory, FieldDescription}
	if m.allDay {
		fields = []FieldType{FieldTitle, FieldAllDay, FieldCategory, FieldDescription}
	}
	
	currentIdx := -1
	for i, field := range fields {
		if field == m.focusedField {
			currentIdx = i
			break
		}
	}
	
	currentIdx += direction
	if currentIdx >= len(fields) {
		currentIdx = 0
	} else if currentIdx < 0 {
		currentIdx = len(fields) - 1
	}
	
	m.focusedField = fields[currentIdx]
	m.updateFocus()
	return m
}

func (m *EventModal) handleAction() (*EventModal, tea.Cmd) {
	if m.categoryMode {
		m.categoryMode = false
		return m, nil
	}
	
	if m.focusedField == FieldCategory {
		m.categoryMode = true
		return m, nil
	}
	
	// Save event
	if err := m.saveEvent(); err == nil {
		return m, func() tea.Msg { return ModalCloseMsg(true) }
	} else {
		m.errorMsg = err.Error()
		return m, nil
	}
}

func (m *EventModal) toggleAllDay() {
	m.allDay = !m.allDay
	if m.allDay {
		m.inputs[inputStartTime].SetValue("")
		m.inputs[inputEndTime].SetValue("")
		if m.focusedField == FieldStartTime || m.focusedField == FieldEndTime {
			m.focusedField = FieldCategory
		}
	} else {
		m.inputs[inputStartTime].SetValue("09:00")
		m.inputs[inputEndTime].SetValue("10:00")
	}
	m.updateFocus()
}

func (m *EventModal) getInputIndex() int {
	switch m.focusedField {
	case FieldTitle:
		return inputTitle
	case FieldStartTime:
		return inputStartTime
	case FieldEndTime:
		return inputEndTime
	case FieldDescription:
		return inputDescription
	}
	return -1
}

func (m *EventModal) updateFocus() {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	
	inputIdx := m.getInputIndex()
	if inputIdx >= 0 && inputIdx < len(m.inputs) {
		m.inputs[inputIdx].Focus()
	}
}

func (m *EventModal) saveEvent() error {
	title := strings.TrimSpace(m.inputs[inputTitle].Value())
	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}
	
	// Get selected category name
	categoryName := ""
	if m.selectedCatIdx < len(m.categories) {
		categoryName = m.categories[m.selectedCatIdx].Name
	}
	
	event := &model.Event{
		Title:       title,
		Category:    categoryName,
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
	
	// Header
	title := "✨ New Event"
	if m.editingEvent != nil {
		title = "✏️ Edit Event"
	}
	
	dateStr := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(m.date.Format("Monday, January 2, 2006"))
	
	header := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")).Render(title),
		dateStr,
		"",
	)
	
	// Build form fields
	var content []string
	
	// Title field with focus indicator
	content = append(content, m.renderField("📝 Title", FieldTitle, m.inputs[inputTitle].View()))
	
	// All-day toggle
	allDayDisplay := "☐ All Day Event"
	if m.allDay {
		allDayDisplay = "☑ All Day Event"
	}
	content = append(content, m.renderField("", FieldAllDay, allDayDisplay))
	
	// Time fields (only if not all-day)
	if !m.allDay {
		content = append(content, m.renderField("🕒 Start Time", FieldStartTime, m.inputs[inputStartTime].View()))
		content = append(content, m.renderField("🕕 End Time", FieldEndTime, m.inputs[inputEndTime].View()))
	}
	
	// Category selector
	if m.categoryMode {
		content = append(content, m.renderCategorySelector())
	} else {
		content = append(content, m.renderField("🏷️ Category", FieldCategory, m.renderSelectedCategory()))
	}
	
	// Description field
	content = append(content, m.renderField("📄 Description", FieldDescription, m.inputs[inputDescription].View()))
	
	// Error message
	if m.errorMsg != "" {
		errorBox := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Background(lipgloss.Color("52")).
			Padding(0, 1).
			Margin(1, 0).
			Bold(true).
			Render("❌ " + m.errorMsg)
		content = append(content, errorBox)
	}
	
	// Instructions
	instructions := m.renderInstructions()
	
	// Combine all content
	mainContent := lipgloss.JoinVertical(lipgloss.Left, content...)
	fullContent := lipgloss.JoinVertical(lipgloss.Left, header, mainContent, "", instructions)
	
	// Modal styling
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(2, 3).
		Width(70).
		Background(lipgloss.Color("0"))
	
	modal := modalStyle.Render(fullContent)
	
	// Center the modal
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal)
}

// Helper methods for rendering UI components
func (m *EventModal) renderField(label string, fieldType FieldType, content string) string {
	isFocused := m.focusedField == fieldType && !m.categoryMode
	
	// Create field container
	fieldStyle := lipgloss.NewStyle().Margin(0, 0, 1, 0)
	if isFocused {
		fieldStyle = fieldStyle.BorderLeft(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("39"))
	}
	
	var parts []string
	if label != "" {
		labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("247"))
		if isFocused {
			labelStyle = labelStyle.Foreground(lipgloss.Color("39")).Bold(true)
		}
		parts = append(parts, labelStyle.Render(label))
	}
	
	// Add focus indicator for non-input fields
	if isFocused && (fieldType == FieldAllDay || fieldType == FieldCategory) {
		content = "▶ " + content
	}
	
	parts = append(parts, content)
	
	field := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return fieldStyle.Render(field)
}

func (m *EventModal) renderSelectedCategory() string {
	if m.selectedCatIdx >= len(m.categories) {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("(none)")
	}
	
	cat := m.categories[m.selectedCatIdx]
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(cat.Color)).
		Render(fmt.Sprintf("● %s", cat.Name))
}

func (m *EventModal) renderCategorySelector() string {
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Render("🏷️ Select Category:")
	
	var categories []string
	for i, cat := range m.categories {
		catStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cat.Color))
		catText := fmt.Sprintf("● %s", cat.Name)
		
		if i == m.selectedCatIdx {
			catText = lipgloss.NewStyle().
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("0")).
				Padding(0, 1).
				Bold(true).
				Render("▶ " + cat.Name + " ◀")
		} else {
			catText = "  " + catStyle.Render(catText)
		}
		
		categories = append(categories, catText)
	}
	
	categoryList := lipgloss.JoinVertical(lipgloss.Left, categories...)
	
	selectorBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Margin(0, 0, 1, 0).
		Background(lipgloss.Color("235"))
	
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		selectorBox.Render(categoryList),
	)
}

func (m *EventModal) renderInstructions() string {
	var instructions []string
	
	if m.categoryMode {
		instructions = append(instructions,
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("↑↓ Navigate categories"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Enter/Space Select"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Esc Cancel"),
		)
	} else {
		instructions = append(instructions,
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Tab/↑↓ Navigate fields"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Space Toggle all-day/category"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true).Render("Enter Save event"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Esc Cancel"),
		)
	}
	
	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		PaddingTop(1).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, instructions...))
}

// DeleteModal for confirming deletion
type DeleteModal struct {
	date     time.Time
	event    *model.Event
	index    int
	styles   *Styles
	width    int
	height   int
	confirmed bool
}

func NewDeleteModal(date time.Time, event *model.Event, index int, styles *Styles) *DeleteModal {
	return &DeleteModal{
		date:      date,
		event:     event,
		index:     index,
		styles:    styles,
		confirmed: false,
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
		case "ctrl+c", "esc", "n", "N":
			// Cancel without deleting
			return m, func() tea.Msg { return ModalCloseMsg(true) }
			
		case "y", "Y", "enter":
			// Confirm deletion
			if err := m.deleteEvent(); err == nil {
				m.confirmed = true
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
	
	// Format event details
	eventTitle := m.event.Title
	if m.event.IsAllDay() {
		eventTitle = fmt.Sprintf("🌅 %s (All Day)", eventTitle)
	} else {
		timeStr := m.event.StartTime
		if m.event.EndTime != "" {
			timeStr = fmt.Sprintf("%s - %s", m.event.StartTime, m.event.EndTime)
		}
		eventTitle = fmt.Sprintf("🕐 %s (%s)", eventTitle, timeStr)
	}
	
	if m.event.Category != "" {
		eventTitle = fmt.Sprintf("%s\n   Category: %s", eventTitle, m.event.Category)
	}
	
	// Build the modal content
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Render("⚠️  Confirm Delete")
	
	dateStr := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(m.date.Format("Monday, January 2, 2006"))
	
	eventBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Margin(1, 0).
		Render(eventTitle)
	
	question := lipgloss.NewStyle().
		Bold(true).
		Render("Delete this event?")
	
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("[Y]es / [N]o")
	
	content := lipgloss.JoinVertical(lipgloss.Center,
		header,
		"",
		dateStr,
		eventBox,
		"",
		question,
		"",
		instructions,
	)
	
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 3).
		Width(60).
		Background(lipgloss.Color("0"))
	
	modal := modalStyle.Render(content)
	
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