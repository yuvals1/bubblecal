package ui

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type App struct {
	app        *tview.Application
	root       *tview.Flex
	pages      *tview.Pages
	header     *tview.TextView
	status     *tview.TextView

	monthView  *MonthView
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
	status := buildStatusBar()

	uiState := &UIState{
		SelectedDate: time.Now(),
		CurrentView:  ViewMonth,
		FocusedPane:  PaneMonth,
	}

	monthView := NewMonthView(uiState)
	agendaView := NewAgendaView(uiState)

	mainArea := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(monthView.Primitive(), 0, 2, true).
		AddItem(agendaView.Primitive(), 30, 0, false)

	pages := tview.NewPages()
	pages.AddPage("main", mainArea, true, true)

	root := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(header, 1, 0, false).
		AddItem(pages, 0, 1, true).
		AddItem(status, 1, 0, false)

	app := &App{
		app:        application,
		root:       root,
		pages:      pages,
		header:     header,
		status:     status,
		monthView:  monthView,
		agendaView: agendaView,
		uiState:    uiState,
	}

	app.bindKeys()
    app.refreshAll()
	application.SetRoot(root, true).EnableMouse(true)
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
		case 'g':
			a.uiState.SelectedDate = time.Now()
            a.refreshAll()
			return nil
		}
		switch ev.Key() {
		case tcell.KeyLeft:
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -1)
		case tcell.KeyRight:
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 1)
		case tcell.KeyUp:
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, -7)
		case tcell.KeyDown:
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 0, 7)
		case tcell.KeyPgUp:
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, -1, 0)
		case tcell.KeyPgDn:
			a.uiState.SelectedDate = a.uiState.SelectedDate.AddDate(0, 1, 0)
		default:
			return ev
		}
        a.refreshAll()
		return nil
	})
}

func (a *App) refreshAll() {
    a.header.SetText(renderHeader(a.uiState.SelectedDate))
    a.status.SetText(renderStatus(time.Now(), a.uiState.SelectedDate))
    a.monthView.Refresh()
    a.agendaView.Refresh()
}
