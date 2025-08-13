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

    // Router: pages for month and week
    centerPages := tview.NewPages()
    centerPages.AddPage("month", monthView.Primitive(), true, true)
    centerPages.AddPage("week", weekView.Primitive(), true, false)

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
        // keep agenda reference; week view is swapped via centerPages
		agendaView: agendaView,
		uiState:    uiState,
	}

	app.bindKeys()
    app.refreshAll()
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
		switch ev.Rune() {
		case 'q':
			a.app.Stop()
			return nil
		case '?':
			ShowHelpModal(a.app, a.pages)
			return nil
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
        case 'k':
            // vim-style up
            if a.uiState.CurrentView == ViewWeek {
                // In week view, delegate to the table's navigation (moves by hour)
                return ev
            } else {
                // In month view: previous week
                a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -7)
                a.refreshAll()
                return nil
            }
        case 'j':
            // vim-style down
            if a.uiState.CurrentView == ViewWeek {
                // In week view, delegate to the table's navigation (moves by hour)
                return ev
            } else {
                // In month view: next week
                a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 7)
                a.refreshAll()
                return nil
            }
        case 'w':
            a.uiState.CurrentView = ViewWeek
            a.center.SwitchToPage("week")
            a.refreshAll()
            return nil
        case 'm':
            a.uiState.CurrentView = ViewMonth
            a.center.SwitchToPage("month")
            a.refreshAll()
            return nil
		case 'g':
			a.uiState.SelectedDate = time.Now()
            a.refreshAll()
			return nil
		case 'a':
			// Add new event
			modals.ShowNewEventModal(a.app, a.pages, a.uiState.SelectedDate, func() {
				a.refreshAll()
			})
			return nil
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
		}
		switch ev.Key() {
		case tcell.KeyLeft:
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -1)
		case tcell.KeyRight:
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 1)
		case tcell.KeyUp:
			if a.uiState.CurrentView == ViewWeek {
				// In week view, delegate to the table's navigation (moves by hour row)
				return ev
			}
			// In month view: previous week
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -7)
		case tcell.KeyDown:
			if a.uiState.CurrentView == ViewWeek {
				// In week view, delegate to the table's navigation (moves by hour row)
				return ev
			}
			// In month view: next week
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 7)
        case tcell.KeyCtrlU:
            // Week view: previous week; Month view: previous month
            if a.uiState.CurrentView == ViewWeek {
                a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -7)
            } else {
                a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, -1, 0)
            }
        case tcell.KeyCtrlD:
            // Week view: next week; Month view: next month
            if a.uiState.CurrentView == ViewWeek {
                a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 7)
            } else {
                a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 1, 0)
            }
		default:
			return ev
		}
        a.refreshAll()
		return nil
	})
}

func (a *App) refreshAll() {
    a.header.SetText(renderHeader(a.uiState.SelectedDate))
    a.monthView.Refresh()
    if a.weekView != nil {
        a.weekView.Refresh()
    }
    a.agendaView.Refresh()
}
