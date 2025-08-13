package tui

import (
	"bubblecal/internal/config"
	"bubblecal/internal/model"
	"bubblecal/internal/storage"
	"fmt"
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
	ListView
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
	agendaBottom  bool // Show agenda at bottom instead of right
	currentTheme  ThemeType // Current UI theme
	
	// Components
	monthView    *MonthViewModel
	weekView     *WeekViewModel
	dayView      *DayViewModel
	listView     *ListViewModel
	agendaView   *AgendaViewModel
	
	// Modal state
	modalStack   []tea.Model
	
	// Jump mode state
	jumpMode     bool
	jumpKeys     []string
	jumpTargets  []time.Time
	
	// Key tracking for double-key combinations
	lastKey      string
	
	// Config
	config       *config.Config
	
	// Styling
	styles       *Styles
}

// ThemeType represents different UI themes
type ThemeType int

const (
	ThemeDefault ThemeType = iota
	ThemeDark
	ThemeLight
	ThemeNeon
	ThemeSolarized
	ThemeNord
	ThemeCount // Keep this last to track number of themes
)

// Jump key sequences (similar to EasyMotion)
var JUMP_KEYS = []string{
	"a", "s", "d", "f", "g", "h", "j", "k", "l",
	"q", "w", "e", "r", "t", "y", "u", "i", "o", "p",
	"z", "x", "c", "v", "b", "n", "m",
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

func GetThemeName(theme ThemeType) string {
	switch theme {
	case ThemeDefault:
		return "Default"
	case ThemeDark:
		return "Dark"
	case ThemeLight:
		return "Light"
	case ThemeNeon:
		return "Neon"
	case ThemeSolarized:
		return "Solarized"
	case ThemeNord:
		return "Nord"
	default:
		return "Unknown"
	}
}

func GetStyles(theme ThemeType) *Styles {
	switch theme {
	case ThemeDark:
		return &Styles{
			Base:           lipgloss.NewStyle(),
			Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250")).Padding(0, 1),
			SelectedDate:   lipgloss.NewStyle().Background(lipgloss.Color("238")).Foreground(lipgloss.Color("15")).Bold(true),
			TodayDate:      lipgloss.NewStyle().Background(lipgloss.Color("17")).Foreground(lipgloss.Color("15")),
			OtherMonth:     lipgloss.NewStyle().Foreground(lipgloss.Color("237")),
			Weekend:        lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
			EventBadge:     lipgloss.NewStyle().Foreground(lipgloss.Color("29")),
		}
	case ThemeLight:
		return &Styles{
			Base:           lipgloss.NewStyle(),
			Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("16")).Padding(0, 1),
			SelectedDate:   lipgloss.NewStyle().Background(lipgloss.Color("39")).Foreground(lipgloss.Color("15")).Bold(true),
			TodayDate:      lipgloss.NewStyle().Background(lipgloss.Color("220")).Foreground(lipgloss.Color("16")),
			OtherMonth:     lipgloss.NewStyle().Foreground(lipgloss.Color("250")),
			Weekend:        lipgloss.NewStyle().Foreground(lipgloss.Color("27")),
			EventBadge:     lipgloss.NewStyle().Foreground(lipgloss.Color("28")),
		}
	case ThemeNeon:
		return &Styles{
			Base:           lipgloss.NewStyle(),
			Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("201")).Padding(0, 1),
			SelectedDate:   lipgloss.NewStyle().Background(lipgloss.Color("201")).Foreground(lipgloss.Color("16")).Bold(true),
			TodayDate:      lipgloss.NewStyle().Background(lipgloss.Color("51")).Foreground(lipgloss.Color("16")),
			OtherMonth:     lipgloss.NewStyle().Foreground(lipgloss.Color("239")),
			Weekend:        lipgloss.NewStyle().Foreground(lipgloss.Color("165")),
			EventBadge:     lipgloss.NewStyle().Foreground(lipgloss.Color("226")),
		}
	case ThemeSolarized:
		return &Styles{
			Base:           lipgloss.NewStyle(),
			Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("136")).Padding(0, 1),
			SelectedDate:   lipgloss.NewStyle().Background(lipgloss.Color("33")).Foreground(lipgloss.Color("230")).Bold(true),
			TodayDate:      lipgloss.NewStyle().Background(lipgloss.Color("64")).Foreground(lipgloss.Color("230")),
			OtherMonth:     lipgloss.NewStyle().Foreground(lipgloss.Color("241")),
			Weekend:        lipgloss.NewStyle().Foreground(lipgloss.Color("37")),
			EventBadge:     lipgloss.NewStyle().Foreground(lipgloss.Color("125")),
		}
	case ThemeNord:
		return &Styles{
			Base:           lipgloss.NewStyle(),
			Header:         lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("109")).Padding(0, 1),
			SelectedDate:   lipgloss.NewStyle().Background(lipgloss.Color("67")).Foreground(lipgloss.Color("231")).Bold(true),
			TodayDate:      lipgloss.NewStyle().Background(lipgloss.Color("96")).Foreground(lipgloss.Color("231")),
			OtherMonth:     lipgloss.NewStyle().Foreground(lipgloss.Color("60")),
			Weekend:        lipgloss.NewStyle().Foreground(lipgloss.Color("103")),
			EventBadge:     lipgloss.NewStyle().Foreground(lipgloss.Color("110")),
		}
	default: // ThemeDefault
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
}

func DefaultStyles() *Styles {
	return GetStyles(ThemeDefault)
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
		agendaBottom:  cfg.AgendaBottom,
		currentTheme:  ThemeType(cfg.Theme),
		config:       cfg,
		styles:       GetStyles(ThemeType(cfg.Theme)),
	}
	
	// Initialize views
	m.monthView = NewMonthViewModel(&m.selectedDate, m.styles, cfg)
	m.weekView = NewWeekViewModel(&m.selectedDate, &m.selectedHour, m.styles, cfg)
	m.weekView.SetShowMiniMonth(m.showMiniMonth)
	m.dayView = NewDayViewModel(&m.selectedDate, &m.selectedHour, m.styles, cfg)
	m.listView = NewListViewModel(&m.selectedDate, m.styles, cfg)
	m.agendaView = NewAgendaViewModel(&m.selectedDate, m.styles, cfg)
	
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
		// Handle jump mode first
		if m.jumpMode {
			return m.handleJumpMode(msg.String()), nil
		}
		
		// Handle double key combinations (gg)
		if msg.String() == "g" {
			if m.lastKey == "g" {
				// gg - go to top
				if m.focusedPane == CalendarPane {
					switch m.currentView {
					case ListView:
						m.listView.GoToTop()
					case WeekView:
						// Set to earliest event hour or default
						m.selectedHour = m.getEarliestHourForWeek()
					case DayView:
						// Set to earliest event hour or default
						m.selectedHour = m.getEarliestHourForDay()
					}
				} else {
					// Agenda pane
					m.agendaView.GoToTop()
				}
				m.lastKey = ""
				return m, nil
			}
			m.lastKey = "g"
			return m, nil
		}
		
		// Clear lastKey if it's not a follow-up g
		if m.lastKey == "g" && msg.String() != "g" {
			m.lastKey = ""
		}
		
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
				m.currentView = ListView
			case ListView:
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
			
		case "f":
			// Enter jump mode
			m.initJumpMode()
			
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
			modal := NewEventModalWithTime(m.selectedDate, nil, defaultTime, m.styles, m.config.Categories)
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
			
		case "p":
			// Toggle agenda position (bottom/right)
			m.agendaBottom = !m.agendaBottom
			// Save the preference
			m.config.AgendaBottom = m.agendaBottom
			m.config.Save()
			// Update view sizes
			m.updateViewSizes()
			
		case "s":
			// Cycle through themes
			m.currentTheme = (m.currentTheme + 1) % ThemeCount
			m.styles = GetStyles(m.currentTheme)
			// Update all views with new styles
			m.updateViewStyles()
			// Save the preference
			m.config.Theme = int(m.currentTheme)
			m.config.Save()
			
		case "G":
			// Go to bottom
			if m.focusedPane == CalendarPane {
				switch m.currentView {
				case ListView:
					m.listView.GoToBottom()
				case WeekView:
					// Set to latest event hour or default
					m.selectedHour = m.getLatestHourForWeek()
				case DayView:
					// Set to latest event hour or default
					m.selectedHour = m.getLatestHourForDay()
				}
			} else {
				// Agenda pane
				m.agendaView.GoToBottom()
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
				cmd := m.handleCalendarNavigation(msg)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
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

func (m *Model) handleCalendarNavigation(msg tea.KeyMsg) tea.Cmd {
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
			maxHour := m.getLatestHourForWeek()
			if m.selectedHour < maxHour {
				m.selectedHour++
			}
		} else if m.currentView == DayView {
			// Move down one hour in day view
			maxHour := m.getLatestHourForDay()
			if m.selectedHour < maxHour {
				m.selectedHour++
			}
		} else if m.currentView == ListView {
			// Move down in list
			m.listView.MoveDown()
		}
	case "k", "up":
		if m.currentView == MonthView {
			m.selectedDate = m.selectedDate.AddDate(0, 0, -7)
		} else if m.currentView == WeekView {
			// Move up one hour in week view
			minHour := m.getEarliestHourForWeek()
			if m.selectedHour > minHour {
				m.selectedHour--
			}
		} else if m.currentView == DayView {
			// Move up one hour in day view
			minHour := m.getEarliestHourForDay()
			if m.selectedHour > minHour {
				m.selectedHour--
			}
		} else if m.currentView == ListView {
			// Move up in list
			m.listView.MoveUp()
		}
	case "ctrl+u":
		if m.currentView == WeekView {
			m.selectedDate = m.selectedDate.AddDate(0, 0, -7)
		} else if m.currentView == ListView {
			m.listView.PageUp()
		} else {
			m.selectedDate = m.selectedDate.AddDate(0, -1, 0)
		}
	case "ctrl+d":
		if m.currentView == WeekView {
			m.selectedDate = m.selectedDate.AddDate(0, 0, 7)
		} else if m.currentView == ListView {
			m.listView.PageDown()
		} else {
			m.selectedDate = m.selectedDate.AddDate(0, 1, 0)
		}
	case "e":
		// Edit selected event in list view
		if m.currentView == ListView {
			if evt := m.listView.GetSelectedEvent(); evt != nil {
				modal := NewEventModal(evt.Date, evt.Event, m.styles, m.config.Categories)
				modal.width = m.width
				modal.height = m.height
				m.modalStack = append(m.modalStack, modal)
				return modal.Init()
			}
		}
	case "d":
		// Delete selected event in list view
		if m.currentView == ListView {
			if evt := m.listView.GetSelectedEvent(); evt != nil {
				if err := storage.DeleteEvent(evt.Date, evt.Event); err == nil {
					m.listView.LoadEvents()
					return loadEventsCmd(m.selectedDate)
				}
			}
		}
	}
	return nil
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
			modal := NewEventModal(m.selectedDate, event, m.styles, m.config.Categories)
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

func (m *Model) updateViewStyles() {
	// Update all views with new styles
	if m.monthView != nil {
		width, height := m.monthView.width, m.monthView.height
		m.monthView = NewMonthViewModel(&m.selectedDate, m.styles, m.config)
		m.monthView.SetSize(width, height)
	}
	if m.weekView != nil {
		width, height := m.weekView.width, m.weekView.height
		showMiniMonth := m.weekView.showMiniMonth
		m.weekView = NewWeekViewModel(&m.selectedDate, &m.selectedHour, m.styles, m.config)
		m.weekView.SetShowMiniMonth(showMiniMonth)
		m.weekView.SetSize(width, height)
	}
	if m.dayView != nil {
		width, height := m.dayView.width, m.dayView.height
		m.dayView = NewDayViewModel(&m.selectedDate, &m.selectedHour, m.styles, m.config)
		m.dayView.SetSize(width, height)
	}
	if m.listView != nil {
		width, height := m.listView.width, m.listView.height
		m.listView = NewListViewModel(&m.selectedDate, m.styles, m.config)
		m.listView.SetSize(width, height)
	}
	if m.agendaView != nil {
		width, height := m.agendaView.width, m.agendaView.height
		events := m.agendaView.events
		selectedIndex := m.agendaView.selectedIndex
		m.agendaView = NewAgendaViewModel(&m.selectedDate, m.styles, m.config)
		m.agendaView.SetEvents(events)
		m.agendaView.selectedIndex = selectedIndex
		m.agendaView.SetSize(width, height)
	}
}

func (m *Model) updateViewSizes() {
	// Calculate sizes for views based on agenda position
	if m.agendaBottom {
		// Agenda at bottom - give it more space (about 1/3 of screen)
		agendaHeight := m.height / 3
		if agendaHeight < 10 {
			agendaHeight = 10 // Minimum height
		}
		calendarWidth := m.width - 2
		calendarHeight := m.height - agendaHeight - 4 // Account for header and borders
		
		if m.monthView != nil {
			m.monthView.SetSize(calendarWidth, calendarHeight)
		}
		if m.weekView != nil {
			m.weekView.SetSize(calendarWidth, calendarHeight)
		}
		if m.dayView != nil {
			m.dayView.SetSize(calendarWidth, calendarHeight)
		}
		if m.listView != nil {
			m.listView.SetSize(m.width - 2, m.height - 3)
		}
		if m.agendaView != nil {
			m.agendaView.SetSize(calendarWidth, agendaHeight)
		}
	} else {
		// Agenda on right (default)
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
		if m.listView != nil {
			m.listView.SetSize(m.width - 2, m.height - 3)
		}
		if m.agendaView != nil {
			m.agendaView.SetSize(agendaWidth, calendarHeight)
		}
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
	
	// Update jump mode state in views
	if m.monthView != nil {
		m.monthView.SetJumpMode(m.jumpMode, m.jumpKeys, m.jumpTargets)
	}
	// TODO: Add jump mode to week and day views
	
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
	case ListView:
		calendarView = m.listView.View()
		viewTitle = " List "
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
	
	// Render based on view mode and agenda position
	var main string
	
	// List view doesn't need agenda pane (it IS the agenda)
	if m.currentView == ListView {
		// Full width list view
		listBox := focusedBorderStyle.
			Width(m.width - 2).
			Height(m.height - 3).
			Render(lipgloss.NewStyle().Padding(0, 1).Render(viewTitle) + "\n" + calendarView)
		main = listBox
	} else if m.agendaBottom {
		// Agenda at bottom layout - give it more space (about 1/3 of screen)
		agendaHeight := m.height / 3
		if agendaHeight < 10 {
			agendaHeight = 10 // Minimum height
		}
		calendarWidth := m.width - 2
		calendarHeight := m.height - agendaHeight - 4
		
		// Render calendar
		calendarBox := calendarBorder.
			Width(calendarWidth).
			MaxHeight(calendarHeight).
			Render(lipgloss.NewStyle().Padding(0, 1).Render(viewTitle) + "\n" + calendarView)
		
		// Render agenda
		agendaBox := agendaBorder.
			Width(calendarWidth).
			Height(agendaHeight).
			Render(lipgloss.NewStyle().Padding(0, 1).Render(" Agenda ") + "\n" + m.agendaView.View())
		
		// Stack calendar on top of agenda
		main = lipgloss.JoinVertical(lipgloss.Left, calendarBox, agendaBox)
	} else {
		// Agenda on right layout (default)
		agendaWidth := 35
		if m.width < 100 {
			agendaWidth = 30
		}
		calendarWidth := m.width - agendaWidth - 3
		contentHeight := m.height - 3
		
		// Render with borders
		calendarBox := calendarBorder.
			Width(calendarWidth).
			MaxHeight(contentHeight).
			Render(lipgloss.NewStyle().Padding(0, 1).Render(viewTitle) + "\n" + calendarView)
		
		agendaBox := agendaBorder.
			Width(agendaWidth).
			Height(contentHeight).
			Render(lipgloss.NewStyle().Padding(0, 1).Render(" Agenda ") + "\n" + m.agendaView.View())
		
		// Combine calendar and agenda side by side
		main = lipgloss.JoinHorizontal(lipgloss.Top, calendarBox, " ", agendaBox)
	}
	
	// Combine header and main content
	return lipgloss.JoinVertical(lipgloss.Left, header, main)
}

func (m *Model) renderHeader() string {
	headerText := " " + m.selectedDate.Format("Jan 2006") + " · BubbleCal · " + GetThemeName(m.currentTheme)
	
	if m.jumpMode {
		jumpStatus := lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Padding(0, 1).
			Render("JUMP MODE")
		headerText += " · " + jumpStatus
	}
	
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

// Jump mode functions

func (m *Model) initJumpMode() {
	m.jumpMode = true
	m.jumpKeys = []string{}
	m.jumpTargets = []time.Time{}
	
	// Check if we're in agenda pane
	if m.focusedPane == AgendaPane {
		m.generateAgendaJumpTargets()
		return
	}
	
	// Generate jump targets based on current view
	switch m.currentView {
	case MonthView:
		m.generateMonthJumpTargets()
	case WeekView:
		m.generateWeekJumpTargets()
	case DayView:
		m.generateDayJumpTargets()
	case ListView:
		m.generateListJumpTargets()
	}
}

func (m *Model) handleJumpMode(key string) *Model {
	// Exit jump mode on escape
	if key == "esc" || key == "ctrl+c" {
		m.jumpMode = false
		m.jumpKeys = nil
		m.jumpTargets = nil
		m.agendaView.SetJumpMode(false, nil)
		m.listView.SetJumpMode(false, nil)
		return m
	}
	
	// Check if key matches any jump target
	for i, jumpKey := range m.jumpKeys {
		if jumpKey == key {
			// Handle different jump types
			if m.focusedPane == AgendaPane {
				// Jump to agenda item
				if i < len(m.events) {
					m.agendaView.JumpToIndex(i)
				}
			} else if m.currentView == ListView {
				// Jump to list item
				m.listView.JumpToIndex(i)
			} else {
				// Jump to calendar date
				if i < len(m.jumpTargets) {
					m.selectedDate = m.jumpTargets[i]
					m.loadEvents()
				}
			}
			// Exit jump mode
			m.jumpMode = false
			m.jumpKeys = nil
			m.jumpTargets = nil
			m.agendaView.SetJumpMode(false, nil)
			m.listView.SetJumpMode(false, nil)
			return m
		}
	}
	
	return m
}

func (m *Model) generateMonthJumpTargets() {
	now := m.selectedDate
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	firstOfNext := firstOfMonth.AddDate(0, 1, 0)
	daysInMonth := int(firstOfNext.Sub(firstOfMonth).Hours()/24 + 0.5)
	
	// Generate targets for all visible days in the month view
	keyIndex := 0
	
	// Add days from current month
	for day := 1; day <= daysInMonth && keyIndex < len(JUMP_KEYS); day++ {
		date := time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location())
		m.jumpTargets = append(m.jumpTargets, date)
		m.jumpKeys = append(m.jumpKeys, JUMP_KEYS[keyIndex])
		keyIndex++
	}
}

func (m *Model) generateWeekJumpTargets() {
	// Generate targets for 7 days of the week
	startOfWeek := m.getStartOfWeek(m.selectedDate)
	
	for i := 0; i < 7 && i < len(JUMP_KEYS); i++ {
		date := startOfWeek.AddDate(0, 0, i)
		m.jumpTargets = append(m.jumpTargets, date)
		m.jumpKeys = append(m.jumpKeys, JUMP_KEYS[i])
	}
}

func (m *Model) generateDayJumpTargets() {
	// For day view, just target the current day (not much to jump to)
	m.jumpTargets = append(m.jumpTargets, m.selectedDate)
	m.jumpKeys = append(m.jumpKeys, JUMP_KEYS[0])
}

func (m *Model) getStartOfWeek(date time.Time) time.Time {
	// Get start of week (Sunday)
	weekday := int(date.Weekday())
	return date.AddDate(0, 0, -weekday)
}

func (m *Model) generateAgendaJumpTargets() {
	// Generate jump keys for visible agenda items
	keyIndex := 0
	for range m.events {
		if keyIndex < len(JUMP_KEYS) {
			m.jumpKeys = append(m.jumpKeys, JUMP_KEYS[keyIndex])
			keyIndex++
		}
	}
	// Pass jump keys to agenda view
	m.agendaView.SetJumpMode(true, m.jumpKeys)
}

func (m *Model) generateListJumpTargets() {
	// Generate jump keys for visible list items
	m.listView.LoadEvents() // Ensure events are loaded
	eventCount := m.listView.GetEventCount()
	
	keyIndex := 0
	for i := 0; i < eventCount && keyIndex < len(JUMP_KEYS); i++ {
		m.jumpKeys = append(m.jumpKeys, JUMP_KEYS[keyIndex])
		keyIndex++
	}
	// Pass jump keys to list view
	m.listView.SetJumpMode(true, m.jumpKeys)
}

// Helper functions for dynamic hour ranges

func (m *Model) getEarliestHourForDay() int {
	events, _ := storage.LoadDayEvents(m.selectedDate)
	minHour := 6 // Default
	
	for _, evt := range events {
		if !evt.IsAllDay() && evt.StartTime != "" {
			var hour int
			if _, err := fmt.Sscanf(evt.StartTime, "%d:", &hour); err == nil {
				if hour < minHour {
					minHour = hour
				}
			}
		}
	}
	
	if minHour < 0 {
		minHour = 0
	}
	return minHour
}

func (m *Model) getLatestHourForDay() int {
	events, _ := storage.LoadDayEvents(m.selectedDate)
	maxHour := 22 // Default
	
	for _, evt := range events {
		if !evt.IsAllDay() && evt.StartTime != "" {
			var hour int
			if _, err := fmt.Sscanf(evt.StartTime, "%d:", &hour); err == nil {
				if hour > maxHour {
					maxHour = hour
				}
			}
		}
	}
	
	if maxHour > 23 {
		maxHour = 23
	}
	return maxHour
}

func (m *Model) getEarliestHourForWeek() int {
	weekStart := m.getStartOfWeek(m.selectedDate)
	minHour := 8 // Default
	
	for d := 0; d < 7; d++ {
		date := weekStart.AddDate(0, 0, d)
		events, _ := storage.LoadDayEvents(date)
		
		for _, evt := range events {
			if !evt.IsAllDay() && evt.StartTime != "" {
				var hour int
				if _, err := fmt.Sscanf(evt.StartTime, "%d:", &hour); err == nil {
					if hour < minHour {
						minHour = hour
					}
				}
			}
		}
	}
	
	if minHour < 0 {
		minHour = 0
	}
	return minHour
}

func (m *Model) getLatestHourForWeek() int {
	weekStart := m.getStartOfWeek(m.selectedDate)
	maxHour := 20 // Default
	
	for d := 0; d < 7; d++ {
		date := weekStart.AddDate(0, 0, d)
		events, _ := storage.LoadDayEvents(date)
		
		for _, evt := range events {
			if !evt.IsAllDay() && evt.StartTime != "" {
				var hour int
				if _, err := fmt.Sscanf(evt.StartTime, "%d:", &hour); err == nil {
					if hour > maxHour {
						maxHour = hour
					}
				}
			}
		}
	}
	
	if maxHour > 23 {
		maxHour = 23
	}
	return maxHour
}