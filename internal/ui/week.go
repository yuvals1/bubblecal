package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type WeekView struct {
	uiState *UIState
	table   *tview.Table
}

func NewWeekView(state *UIState) *WeekView {
	t := tview.NewTable()
	t.SetBorders(true)
	// Row 0 = weekdays header, Col 0 = hour labels
	t.SetFixed(1, 1)
	t.SetSelectable(true, true)
	t.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack).Bold(true))
	wv := &WeekView{uiState: state, table: t}
	wv.Refresh()
	return wv
}

func (w *WeekView) Primitive() tview.Primitive { return w.table }

func (w *WeekView) Refresh() {
	w.table.Clear()

	// Compute start of week (Sunday)
	sel := w.uiState.SelectedDate
	weekStart := sel.AddDate(0, 0, -int(sel.Weekday()))

	// Header row: weekdays + day of month
	w.table.SetCell(0, 0, tview.NewTableCell(" ").SetSelectable(false))
	for d := 0; d < 7; d++ {
		date := weekStart.AddDate(0, 0, d)
		label := fmt.Sprintf("[::b]%s %d[::-]", date.Weekday().String()[:3], date.Day())
		cell := tview.NewTableCell(label).SetAlign(tview.AlignCenter)
		w.table.SetCell(0, d+1, cell)
	}

	// Hours rows (08:00 - 20:00)
	startHour := 8
	endHour := 20
	for r, h := 1, startHour; h <= endHour; r, h = r+1, h+1 {
		w.table.SetCell(r, 0, tview.NewTableCell(fmt.Sprintf("%02d:00", h)).SetAlign(tview.AlignRight).SetSelectable(false))
		for c := 0; c < 7; c++ {
			date := weekStart.AddDate(0, 0, c)
			cell := tview.NewTableCell(" ").SetExpansion(1)
			// Mock event placement: render title at the event hour
			for _, evt := range mockEventsFor(date) {
				if evt.Time.Hour() == h {
					cell.SetText(fmt.Sprintf("[green]●[::-] %s", evt.Title))
				}
			}
			// Highlight today by underlining content
			if sameDay(date, time.Now()) {
				cell = cell.SetText(fmt.Sprintf("[::u]%s[::-]", cell.Text))
			}
			w.table.SetCell(r, c+1, cell)
		}
	}

	// Move selection to current day column around midday
	w.syncSelectionToTable()
}

func (w *WeekView) syncSelectionToTable() {
	sel := w.uiState.SelectedDate
	weekStart := sel.AddDate(0, 0, -int(sel.Weekday()))
	col := int(sel.Sub(weekStart).Hours()/24) + 1 // +1 for hour label column
	if col < 1 {
		col = 1
	}
	if col > 7 {
		col = 7
	}
	midRow := 1 + (12 - 8) // 12:00 row index
	w.table.Select(midRow, col)
}
