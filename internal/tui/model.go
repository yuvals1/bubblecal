package tui

import (
	"simple-tui-cal/internal/config"
	"simple-tui-cal/internal/model"
	"simple-tui-cal/internal/storage"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current view
type ViewMode int

const (
	MonthView ViewMode = iota
	WeekView
	DayView
)

// FocusedPane represents which pane has focus
type FocusedPane int

const (
	CalendarPane FocusedPane = iota
	AgendaPane
)

// Model is the main application model
type Model struct {
	// Core state
	selectedDate time.Time
	currentView  ViewMode
	focusedPane  FocusedPane
	selectedHour int // Selected hour for week/day views
	
	// Data
	events       []*model.Event // Events for selected date
	
	// UI state
	width        int
	height       int
	showMiniMonth bool // Toggle mini-month view in week view
	
	// Components
	monthView    *MonthViewModel
	weekView     *WeekViewModel
	dayView      *DayViewModel
	agendaView   *AgendaViewModel
	
	// Modal state
	modalStack   []tea.Model
	
	// Config
	config       *config.Config
	
	// Styling
	styles       *Styles
}

// Styles holds all the lipgloss styles
type Styles struct {
	Base           lipgloss.Style
	Header         lipgloss.Style
	SelectedDate   lipgloss.Style
	TodayDate      lipgloss.Style
	OtherMonth     lipgloss.Style
	Weekend        lipgloss.Style
	EventBadge     lipgloss.Style
}

func DefaultStyles() *Styles {
	return &Styles{
		Base:           lipgloss.NewStyle(),
		Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Padding(0, 1),
		SelectedDate:   lipgloss.NewStyle().Background(lipgloss.Color("33")).Foreground(lipgloss.Color("0")).Bold(true),
		TodayDate:      lipgloss.NewStyle().Background(lipgloss.Color("21")).Foreground(lipgloss.Color("15")),
		OtherMonth:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Weekend:        lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		EventBadge:     lipgloss.NewStyle().Foreground(lipgloss.Color("34")),
	}
}

// NewModel creates a new application model
func NewModel() *Model {
	now := time.Now()
	
	// Load configuration
	cfg, _ := config.Load()
	
	m := &Model{
		selectedDate: now,
		currentView:  MonthView,
		focusedPane:  CalendarPane,
		selectedHour: 12, // Default to noon
		showMiniMonth: cfg.ShowMiniMonth,
		config:       cfg,
		styles:       DefaultStyles(),
	}
	
	// Initialize views
	m.monthView = NewMonthViewModel(&m.selectedDate, m.styles)
	m.weekView = NewWeekViewModel(&m.selectedDate, &m.selectedHour, m.styles)
	m.weekView.SetShowMiniMonth(m.showMiniMonth)
	m.dayView = NewDayViewModel(&m.selectedDate, &m.selectedHour, m.styles)
	m.agendaView = NewAgendaViewModel(&m.selectedDate, m.styles)
	
	// Load initial events
	m.loadEvents()
	
	return m
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		loadEventsCmd(m.selectedDate),
	)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	// Handle modal updates first
	if len(m.modalStack) > 0 {
		modal := m.modalStack[len(m.modalStack)-1]
		newModal, cmd := modal.Update(msg)
		m.modalStack[len(m.modalStack)-1] = newModal
		
		// Check if modal wants to close
		if _, ok := msg.(ModalCloseMsg); ok {
			m.modalStack = m.modalStack[:len(m.modalStack)-1]
			// Reload events after modal closes
			m.loadEvents()
			cmds = append(cmds, loadEventsCmd(m.selectedDate))
		}
		
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
		
		// Don't process other messages when modal is open
		if len(m.modalStack) > 0 {
			return m, tea.Batch(cmds...)
		}
	}
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewSizes()
		
	case EventsLoadedMsg:
		m.events = msg.Events
		m.agendaView.SetEvents(m.events)
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
			
		case "tab":
			// Toggle focus between calendar and agenda
			if m.focusedPane == CalendarPane {
				m.focusedPane = AgendaPane
			} else {
				m.focusedPane = CalendarPane
			}
			
		case " ":
			// Cycle through views
			switch m.currentView {
			case MonthView:
				m.currentView = WeekView
				// Set selected hour to current hour if today, otherwise noon
				if sameDay(m.selectedDate, time.Now()) {
					hour := time.Now().Hour()
					if hour >= 8 && hour <= 20 {
						m.selectedHour = hour
					} else {
						m.selectedHour = 12
					}
				} else {
					m.selectedHour = 12
				}
			case WeekView:
				m.currentView = DayView
				// Adjust selected hour for day view range if needed
				if m.selectedHour < 6 {
					m.selectedHour = 6
				} else if m.selectedHour > 22 {
					m.selectedHour = 22
				}
			case DayView:
				m.currentView = MonthView
			}
			
		case "t", ".":
			// Go to today
			m.selectedDate = time.Now()
			// Also set to current hour in week/day views
			if m.currentView == WeekView {
				hour := time.Now().Hour()
				if hour >= 8 && hour <= 20 {
					m.selectedHour = hour
				} else if hour < 8 {
					m.selectedHour = 8
				} else {
					m.selectedHour = 20
				}
			} else if m.currentView == DayView {
				hour := time.Now().Hour()
				if hour >= 6 && hour <= 22 {
					m.selectedHour = hour
				} else if hour < 6 {
					m.selectedHour = 6
				} else {
					m.selectedHour = 22
				}
			}
			m.loadEvents()
			cmds = append(cmds, loadEventsCmd(m.selectedDate))
			
		case "a":
			// Add new event
			defaultTime := ""
			// Get the selected hour if in week or day view and calendar is focused
			if m.focusedPane == CalendarPane {
				switch m.currentView {
				case WeekView:
					defaultTime = m.weekView.GetSelectedHour()
				case DayView:
					defaultTime = m.dayView.GetSelectedHour()
				}
			}
			modal := NewEventModalWithTime(m.selectedDate, nil, defaultTime, m.styles)
			// Set modal window size
			modal.width = m.width
			modal.height = m.height
			m.modalStack = append(m.modalStack, modal)
			return m, modal.Init()
			
		case "m":
			// Toggle mini-month view (only in week view)
			if m.currentView == WeekView {
				m.showMiniMonth = !m.showMiniMonth
				// Update week view with the new setting
				m.weekView.SetShowMiniMonth(m.showMiniMonth)
				// Save the preference
				m.config.ShowMiniMonth = m.showMiniMonth
				m.config.Save()
			}
			
		case "?":
			// Show help
			modal := NewHelpModal(m.currentView, m.focusedPane, m.styles)
			// Set modal window size
			modal.width = m.width
			modal.height = m.height
			m.modalStack = append(m.modalStack, modal)
			return m, nil
			
		default:
			// Handle navigation based on focused pane
			if m.focusedPane == CalendarPane {
				oldDate := m.selectedDate
				m.handleCalendarNavigation(msg)
				if !sameDay(oldDate, m.selectedDate) {
					m.loadEvents()
					cmds = append(cmds, loadEventsCmd(m.selectedDate))
				}
			} else {
				// Handle agenda navigation
				cmd := m.handleAgendaNavigation(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
	}
	
	return m, tea.Batch(cmds...)
}

func (m *Model) handleCalendarNavigation(msg tea.KeyMsg) {
	switch msg.String() {
	case "h", "left":
		m.selectedDate = m.selectedDate.AddDate(0, 0, -1)
	case "l", "right":
		m.selectedDate = m.selectedDate.AddDate(0, 0, 1)
	case "j", "down":
		if m.currentView == MonthView {
			m.selectedDate = m.selectedDate.AddDate(0, 0, 7)
		} else if m.currentView == WeekView {
			// Move down one hour in week view
			if m.selectedHour < 20 { // Max hour is 20:00
				m.selectedHour++
			}
		} else if m.currentView == DayView {
			// Move down one hour in day view
			if m.selectedHour < 22 { // Day view goes to 22:00
				m.selectedHour++
			}
		}
	case "k", "up":
		if m.currentView == MonthView {
			m.selectedDate = m.selectedDate.AddDate(0, 0, -7)
		} else if m.currentView == WeekView {
			// Move up one hour in week view
			if m.selectedHour > 8 { // Min hour is 8:00
				m.selectedHour--
			}
		} else if m.currentView == DayView {
			// Move up one hour in day view
			if m.selectedHour > 6 { // Day view starts at 6:00
				m.selectedHour--
			}
		}
	case "ctrl+u":
		if m.currentView == WeekView {
			m.selectedDate = m.selectedDate.AddDate(0, 0, -7)
		} else {
			m.selectedDate = m.selectedDate.AddDate(0, -1, 0)
		}
	case "ctrl+d":
		if m.currentView == WeekView {
			m.selectedDate = m.selectedDate.AddDate(0, 0, 7)
		} else {
			m.selectedDate = m.selectedDate.AddDate(0, 1, 0)
		}
	}
}

func (m *Model) handleAgendaNavigation(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "j", "down":
		m.agendaView.MoveDown()
	case "k", "up":
		m.agendaView.MoveUp()
	case "e":
		// Edit selected event
		if idx := m.agendaView.GetSelectedIndex(); idx >= 0 && idx < len(m.events) {
			event := m.events[idx]
			modal := NewEventModal(m.selectedDate, event, m.styles)
			// Set modal window size
			modal.width = m.width
			modal.height = m.height
			m.modalStack = append(m.modalStack, modal)
			return modal.Init()
		}
	case "d":
		// Delete selected event directly
		if idx := m.agendaView.GetSelectedIndex(); idx >= 0 && idx < len(m.events) {
			event := m.events[idx]
			// Delete the event immediately
			if err := storage.DeleteEvent(m.selectedDate, event); err == nil {
				// Reload events after successful deletion
				m.loadEvents()
				return loadEventsCmd(m.selectedDate)
			}
		}
	case "h", "left":
		m.selectedDate = m.selectedDate.AddDate(0, 0, -1)
		m.loadEvents()
		return loadEventsCmd(m.selectedDate)
	case "l", "right":
		m.selectedDate = m.selectedDate.AddDate(0, 0, 1)
		m.loadEvents()
		return loadEventsCmd(m.selectedDate)
	}
	return nil
}

func (m *Model) updateViewSizes() {
	// Calculate sizes for views
	agendaWidth := 35
	if m.width < 100 {
		agendaWidth = 30
	}
	calendarWidth := m.width - agendaWidth - 3 // Account for borders and spacing
	calendarHeight := m.height - 3 // Account for header
	
	if m.monthView != nil {
		m.monthView.SetSize(calendarWidth, calendarHeight)
	}
	if m.weekView != nil {
		m.weekView.SetSize(calendarWidth, calendarHeight)
	}
	if m.dayView != nil {
		m.dayView.SetSize(calendarWidth, calendarHeight)
	}
	if m.agendaView != nil {
		m.agendaView.SetSize(agendaWidth, calendarHeight)
	}
}

func (m *Model) loadEvents() {
	events, _ := storage.LoadDayEvents(m.selectedDate)
	m.events = events
	if m.agendaView != nil {
		m.agendaView.SetEvents(m.events)
	}
}

// View renders the UI
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	
	// If modal is open, render it on top
	if len(m.modalStack) > 0 {
		modal := m.modalStack[len(m.modalStack)-1]
		return modal.View()
	}
	
	// Render header
	header := m.renderHeader()
	
	// Render main content
	var calendarView string
	viewTitle := ""
	switch m.currentView {
	case MonthView:
		calendarView = m.monthView.View()
		viewTitle = " Monthly "
	case WeekView:
		calendarView = m.weekView.View()
		viewTitle = " Weekly "
	case DayView:
		calendarView = m.dayView.View()
		viewTitle = " Daily "
	}
	
	// Create borders for calendar and agenda
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))
	
	focusedBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("220"))
	
	// Apply border styling based on focus
	calendarBorder := borderStyle
	agendaBorder := borderStyle
	
	if m.focusedPane == CalendarPane {
		calendarBorder = focusedBorderStyle
	} else {
		agendaBorder = focusedBorderStyle
	}
	
	// Calculate dimensions
	agendaWidth := 35
	if m.width < 100 {
		agendaWidth = 30
	}
	calendarWidth := m.width - agendaWidth - 3
	contentHeight := m.height - 3
	
	// Render with borders
	calendarBox := calendarBorder.
		Width(calendarWidth).
		Height(contentHeight).
		Render(lipgloss.NewStyle().Padding(0, 1).Render(viewTitle) + "\n" + calendarView)
	
	agendaBox := agendaBorder.
		Width(agendaWidth).
		Height(contentHeight).
		Render(lipgloss.NewStyle().Padding(0, 1).Render(" Agenda ") + "\n" + m.agendaView.View())
	
	// Combine calendar and agenda side by side
	main := lipgloss.JoinHorizontal(lipgloss.Top, calendarBox, " ", agendaBox)
	
	// Combine header and main content
	return lipgloss.JoinVertical(lipgloss.Left, header, main)
}

func (m *Model) renderHeader() string {
	headerText := " " + m.selectedDate.Format("Jan 2006") + " · Simple TUI Cal"
	return m.styles.Header.Width(m.width).Render(headerText)
}

// Helper functions

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

// Messages

type EventsLoadedMsg struct {
	Events []*model.Event
}

type ModalCloseMsg bool

// Commands

func loadEventsCmd(date time.Time) tea.Cmd {
	return func() tea.Msg {
		events, _ := storage.LoadDayEvents(date)
		return EventsLoadedMsg{Events: events}
	}
}