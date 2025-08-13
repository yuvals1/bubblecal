package tui

import (
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
	
	// Data
	events       []*model.Event // Events for selected date
	
	// UI state
	width        int
	height       int
	
	// Components
	monthView    *MonthViewModel
	weekView     *WeekViewModel
	dayView      *DayViewModel
	agendaView   *AgendaViewModel
	
	// Modal state
	modalStack   []tea.Model
	
	// Styling
	styles       *Styles
}

// Styles holds all the lipgloss styles
type Styles struct {
	Base           lipgloss.Style
	Header         lipgloss.Style
	CalendarBorder lipgloss.Style
	AgendaBorder   lipgloss.Style
	SelectedDate   lipgloss.Style
	TodayDate      lipgloss.Style
	OtherMonth     lipgloss.Style
	Weekend        lipgloss.Style
	EventBadge     lipgloss.Style
	FocusedBorder  lipgloss.Style
	UnfocusedBorder lipgloss.Style
}

func DefaultStyles() *Styles {
	return &Styles{
		Base:           lipgloss.NewStyle(),
		Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")),
		CalendarBorder: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		AgendaBorder:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
		SelectedDate:   lipgloss.NewStyle().Background(lipgloss.Color("33")).Foreground(lipgloss.Color("0")).Bold(true),
		TodayDate:      lipgloss.NewStyle().Background(lipgloss.Color("21")).Foreground(lipgloss.Color("15")),
		OtherMonth:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		Weekend:        lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		EventBadge:     lipgloss.NewStyle().Foreground(lipgloss.Color("34")),
		FocusedBorder:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("220")),
		UnfocusedBorder: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("240")),
	}
}

// NewModel creates a new application model
func NewModel() *Model {
	now := time.Now()
	m := &Model{
		selectedDate: now,
		currentView:  MonthView,
		focusedPane:  CalendarPane,
		styles:       DefaultStyles(),
	}
	
	// Initialize views
	m.monthView = NewMonthViewModel(&m.selectedDate, m.styles)
	m.weekView = NewWeekViewModel(&m.selectedDate, m.styles)
	m.dayView = NewDayViewModel(&m.selectedDate, m.styles)
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
			case WeekView:
				m.currentView = DayView
			case DayView:
				m.currentView = MonthView
			}
			
		case "t", ".":
			// Go to today
			m.selectedDate = time.Now()
			cmds = append(cmds, loadEventsCmd(m.selectedDate))
			
		case "a":
			// Add new event
			modal := NewEventModal(m.selectedDate, nil, m.styles)
			m.modalStack = append(m.modalStack, modal)
			return m, modal.Init()
			
		case "?":
			// Show help
			modal := NewHelpModal(m.currentView, m.focusedPane, m.styles)
			m.modalStack = append(m.modalStack, modal)
			return m, nil
			
		default:
			// Handle navigation based on focused pane
			if m.focusedPane == CalendarPane {
				oldDate := m.selectedDate
				m.handleCalendarNavigation(msg)
				if !sameDay(oldDate, m.selectedDate) {
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
		} else {
			// In week/day view, this would move through hours
			// For now, just move to next day
			m.selectedDate = m.selectedDate.AddDate(0, 0, 1)
		}
	case "k", "up":
		if m.currentView == MonthView {
			m.selectedDate = m.selectedDate.AddDate(0, 0, -7)
		} else {
			// In week/day view, this would move through hours
			// For now, just move to previous day
			m.selectedDate = m.selectedDate.AddDate(0, 0, -1)
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
		if event := m.agendaView.GetSelectedEvent(); event != nil {
			modal := NewEventModal(m.selectedDate, event, m.styles)
			m.modalStack = append(m.modalStack, modal)
			return modal.Init()
		}
	case "d":
		// Delete selected event
		if event := m.agendaView.GetSelectedEvent(); event != nil {
			modal := NewDeleteModal(m.selectedDate, event, m.agendaView.GetSelectedIndex(), m.styles)
			m.modalStack = append(m.modalStack, modal)
			return nil
		}
	case "h", "left":
		m.selectedDate = m.selectedDate.AddDate(0, 0, -1)
		return loadEventsCmd(m.selectedDate)
	case "l", "right":
		m.selectedDate = m.selectedDate.AddDate(0, 0, 1)
		return loadEventsCmd(m.selectedDate)
	}
	return nil
}

func (m *Model) updateViewSizes() {
	// Calculate sizes for views
	agendaWidth := 30
	calendarWidth := m.width - agendaWidth - 1 // -1 for border
	
	if m.monthView != nil {
		m.monthView.SetSize(calendarWidth, m.height-2) // -2 for header
	}
	if m.weekView != nil {
		m.weekView.SetSize(calendarWidth, m.height-2)
	}
	if m.dayView != nil {
		m.dayView.SetSize(calendarWidth, m.height-2)
	}
	if m.agendaView != nil {
		m.agendaView.SetSize(agendaWidth, m.height-2)
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
	switch m.currentView {
	case MonthView:
		calendarView = m.monthView.View()
	case WeekView:
		calendarView = m.weekView.View()
	case DayView:
		calendarView = m.dayView.View()
	}
	
	// Apply border styling based on focus
	calendarStyle := m.styles.UnfocusedBorder
	agendaStyle := m.styles.UnfocusedBorder
	
	if m.focusedPane == CalendarPane {
		calendarStyle = m.styles.FocusedBorder
	} else {
		agendaStyle = m.styles.FocusedBorder
	}
	
	// Get view titles
	viewTitle := ""
	switch m.currentView {
	case MonthView:
		viewTitle = "Monthly"
	case WeekView:
		viewTitle = "Weekly"
	case DayView:
		viewTitle = "Daily"
	}
	
	calendarView = calendarStyle.
		Width(m.width - 32).
		Height(m.height - 2).
		SetString(viewTitle).
		Render(calendarView)
	
	agendaView := agendaStyle.
		Width(30).
		Height(m.height - 2).
		SetString("Agenda").
		Render(m.agendaView.View())
	
	// Combine calendar and agenda side by side
	main := lipgloss.JoinHorizontal(lipgloss.Top, calendarView, agendaView)
	
	// Combine header and main content
	return lipgloss.JoinVertical(lipgloss.Left, header, main)
}

func (m *Model) renderHeader() string {
	headerText := m.selectedDate.Format("Jan 2006") + " · Simple TUI Cal"
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