package ui

import (
	"simple-tui-cal/internal/ui/modals"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	app        *tview.Application
	root       *tview.Flex
    pages      *tview.Pages
    center     *tview.Pages
	header     *tview.TextView

	monthView  *MonthView
    weekView   *WeekView
    dayView    *DayView
	agendaView *AgendaView

	uiState    *UIState
}

type UIState struct {
	SelectedDate time.Time
	CurrentView  ViewKind
	FocusedPane  PaneKind
}

type ViewKind int

const (
	ViewMonth ViewKind = iota
	ViewWeek
	ViewDay
)

type PaneKind int

const (
	PaneMonth PaneKind = iota
	PaneAgenda
)

func NewApp() (*App, error) {
	application := tview.NewApplication()

	header := buildHeader()

	uiState := &UIState{
		SelectedDate: time.Now(),
		CurrentView:  ViewMonth,
		FocusedPane:  PaneMonth,
	}

	monthView := NewMonthView(uiState)
	agendaView := NewAgendaView(uiState)
    weekView := NewWeekView(uiState)
    dayView := NewDayView(uiState)

    // Router: pages for month, week, and day
    centerPages := tview.NewPages()
    centerPages.AddPage("month", monthView.Primitive(), true, true)
    centerPages.AddPage("week", weekView.Primitive(), true, false)
    centerPages.AddPage("day", dayView.Primitive(), true, false)

    mainArea := tview.NewFlex().SetDirection(tview.FlexColumn).
        AddItem(centerPages, 0, 1, true).
        AddItem(agendaView.Primitive(), 30, 0, false)

	pages := tview.NewPages()
	pages.AddPage("main", mainArea, true, true)

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
        AddItem(pages, 0, 1, true)

    app := &App{
		app:        application,
		root:       root,
		pages:      pages,
        center:     centerPages,
		header:     header,
		monthView:  monthView,
        weekView:   weekView,
        dayView:    dayView,
		agendaView: agendaView,
		uiState:    uiState,
	}

	app.bindKeys()
    app.refreshAll()
    // Set initial focus on month view
    app.monthView.SetFocused(true)
    app.agendaView.SetFocused(false)
    application.SetRoot(root, true).EnableMouse(true)

	// Keep the table's selection synchronized if the user moves it with arrow keys
	monthView.table.SetSelectedFunc(func(row, column int) {
		if cell := monthView.table.GetCell(row, column); cell != nil {
			if ref, ok := cell.GetReference().(time.Time); ok {
				uiState.SelectedDate = ref
				app.refreshAll()
			}
		}
	})
	return app, nil
}

func (a *App) Run() error {
	return a.app.Run()
}

func (a *App) bindKeys() {
	a.app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey {
		// Check if a modal is currently active (by checking if "main" is the front page)
		frontPage, _ := a.pages.GetFrontPage()
		isModalOpen := frontPage != "main"
		
		// If a modal is open, let it handle all keys (including Tab)
		if isModalOpen {
			// Only handle global quit and help when modal is open
			if ev.Rune() == 'q' && ev.Modifiers() == tcell.ModNone {
				// Let 'q' work in form fields
				return ev
			}
			// Let all keys including Tab pass through to the modal
			return ev
		}
		
		// Tab toggles focus only when no modal is open
		if ev.Key() == tcell.KeyTab {
			if a.uiState.FocusedPane == PaneMonth {
				a.uiState.FocusedPane = PaneAgenda
				a.app.SetFocus(a.agendaView.Primitive())
				// Update visual indicators
				a.monthView.SetFocused(false)
				a.weekView.SetFocused(false)
				a.dayView.SetFocused(false)
				a.agendaView.SetFocused(true)
			} else {
				a.uiState.FocusedPane = PaneMonth
				a.agendaView.SetFocused(false)
				a.monthView.SetFocused(false)
				a.weekView.SetFocused(false)
				a.dayView.SetFocused(false)
				
				switch a.uiState.CurrentView {
				case ViewWeek:
					a.app.SetFocus(a.weekView.Primitive())
					a.weekView.SetFocused(true)
				case ViewDay:
					a.app.SetFocus(a.dayView.Primitive())
					a.dayView.SetFocused(true)
				default:
					a.app.SetFocus(a.monthView.Primitive())
					a.monthView.SetFocused(true)
				}
			}
			return nil
		}

		// Space key cycles through views
		if ev.Rune() == ' ' {
			// Cycle: Month -> Week -> Day -> Month
			switch a.uiState.CurrentView {
			case ViewMonth:
				a.switchToView(ViewWeek)
			case ViewWeek:
				a.switchToView(ViewDay)
			case ViewDay:
				a.switchToView(ViewMonth)
			}
			return nil
		}
		
		// Global shortcuts that work regardless of focus
		switch ev.Rune() {
		case 'q':
			a.app.Stop()
			return nil
		case '?':
			ShowHelpModal(a.app, a.pages, a.uiState.CurrentView, a.uiState.FocusedPane)
			return nil
		case 't', '.':
			// Both 't' and '.' go to today
			a.uiState.SelectedDate = time.Now()
			a.refreshAll()
			return nil
		case 'a':
			// Add new event
			defaultTime := ""
			// Get the selected hour if in week or day view and calendar is focused
			if a.uiState.FocusedPane == PaneMonth {
				switch a.uiState.CurrentView {
				case ViewWeek:
					defaultTime = a.weekView.GetSelectedHour()
				case ViewDay:
					defaultTime = a.dayView.GetSelectedHour()
				}
			}
			modals.ShowNewEventModal(a.app, a.pages, a.uiState.SelectedDate, defaultTime, func() {
				a.refreshAll()
			})
			return nil
		}

		// Handle navigation based on which pane is focused
		if a.uiState.FocusedPane == PaneAgenda {
			// Agenda is focused - handle agenda-specific keys
			switch ev.Rune() {
			case 'e':
				// Edit selected event
				if a.agendaView.GetSelectedEvent() != nil {
					modals.ShowEditEventModal(a.app, a.pages, a.uiState.SelectedDate, 
						a.agendaView.GetSelectedEvent(), a.agendaView.GetSelectedIndex(), func() {
						a.refreshAll()
					})
				}
				return nil
			case 'd':
				// Delete selected event
				if a.agendaView.GetSelectedEvent() != nil {
					modals.ShowDeleteConfirmModal(a.app, a.pages, a.uiState.SelectedDate,
						a.agendaView.GetSelectedEvent(), a.agendaView.GetSelectedIndex(), func() {
						a.refreshAll()
					})
				}
				return nil
			case 'h', 'l':
				// Still allow h/l to change dates even when agenda is focused
				if ev.Rune() == 'h' {
					a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -1)
				} else {
					a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 1)
				}
				a.refreshAll()
				return nil
			}
			
			switch ev.Key() {
			case tcell.KeyUp, tcell.KeyDown:
				// Let the list handle these
				return ev
			case tcell.KeyLeft:
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -1)
				a.refreshAll()
				return nil
			case tcell.KeyRight:
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 1)
				a.refreshAll()
				return nil
			}
		} else {
			// Calendar is focused - handle calendar navigation
			switch ev.Rune() {
			case 'h':
				// vim-style left: previous day
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -1)
				a.refreshAll()
				return nil
			case 'l':
				// vim-style right: next day
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 1)
				a.refreshAll()
				return nil
			case 'j':
				// vim-style down
				if a.uiState.CurrentView == ViewWeek || a.uiState.CurrentView == ViewDay {
					// In week/day view, let table handle it (moves by hour)
					return ev
				}
				// In month view: next week
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 7)
				a.refreshAll()
				return nil
			case 'k':
				// vim-style up
				if a.uiState.CurrentView == ViewWeek || a.uiState.CurrentView == ViewDay {
					// In week/day view, let table handle it (moves by hour)
					return ev
				}
				// In month view: previous week
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -7)
				a.refreshAll()
				return nil
			}

			switch ev.Key() {
			case tcell.KeyLeft:
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -1)
				a.refreshAll()
				return nil
			case tcell.KeyRight:
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 1)
				a.refreshAll()
				return nil
			case tcell.KeyUp:
				if a.uiState.CurrentView == ViewWeek || a.uiState.CurrentView == ViewDay {
					// In week/day view, let table handle it (moves by hour row)
					return ev
				}
				// In month view: previous week
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -7)
				a.refreshAll()
				return nil
			case tcell.KeyDown:
				if a.uiState.CurrentView == ViewWeek || a.uiState.CurrentView == ViewDay {
					// In week/day view, let table handle it (moves by hour row)
					return ev
				}
				// In month view: next week
				a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 7)
				a.refreshAll()
				return nil
			case tcell.KeyCtrlU:
				// Week view: previous week; Month view: previous month
				if a.uiState.CurrentView == ViewWeek {
					a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -7)
				} else {
					a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, -1, 0)
				}
				a.refreshAll()
				return nil
			case tcell.KeyCtrlD:
				// Week view: next week; Month view: next month
				if a.uiState.CurrentView == ViewWeek {
					a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 7)
				} else {
					a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 1, 0)
				}
				a.refreshAll()
				return nil
			}
		}

		// Default: pass through
		return ev
	})
}

func (a *App) switchToView(view ViewKind) {
	a.uiState.CurrentView = view
	
	// Update focus indicators if calendar is focused
	if a.uiState.FocusedPane == PaneMonth {
		a.monthView.SetFocused(false)
		a.weekView.SetFocused(false)
		a.dayView.SetFocused(false)
		
		switch view {
		case ViewMonth:
			a.center.SwitchToPage("month")
			a.monthView.SetFocused(true)
			a.app.SetFocus(a.monthView.Primitive())
		case ViewWeek:
			a.center.SwitchToPage("week")
			a.weekView.SetFocused(true)
			a.app.SetFocus(a.weekView.Primitive())
		case ViewDay:
			a.center.SwitchToPage("day")
			a.dayView.SetFocused(true)
			a.app.SetFocus(a.dayView.Primitive())
		}
	} else {
		// Just switch the page without changing focus
		switch view {
		case ViewMonth:
			a.center.SwitchToPage("month")
		case ViewWeek:
			a.center.SwitchToPage("week")
		case ViewDay:
			a.center.SwitchToPage("day")
		}
	}
	
	a.refreshAll()
}

func (a *App) refreshAll() {
    a.header.SetText(renderHeader(a.uiState.SelectedDate))
    a.monthView.Refresh()
    if a.weekView != nil {
        a.weekView.Refresh()
    }
    if a.dayView != nil {
        a.dayView.Refresh()
    }
    a.agendaView.Refresh()
}
