package ui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type WeekView struct {
    uiState   *UIState
    root      *tview.Flex
    table     *tview.Table
    miniMonth *tview.Table
}

func NewWeekView(state *UIState) *WeekView {
    t := tview.NewTable()
    t.SetBorders(true)
    // Row 0 = weekdays header, Col 0 = hour labels
    t.SetFixed(1, 1)
    t.SetSelectable(true, true)
    t.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack).Bold(true))

    mini := tview.NewTable()
    mini.SetBorders(false)
    mini.SetFixed(1, 0)

    root := tview.NewFlex().SetDirection(tview.FlexRow).
        AddItem(t, 0, 1, true).
        AddItem(mini, 7, 0, false) // small fixed-height mini calendar

    wv := &WeekView{uiState: state, root: root, table: t, miniMonth: mini}
    wv.Refresh()
    return wv
}

func (w *WeekView) Primitive() tview.Primitive { return w.root }

func (w *WeekView) Refresh() {
    w.table.Clear()

	// Compute start of week (Sunday)
	sel := w.uiState.SelectedDate
	weekStart := sel.AddDate(0, 0, -int(sel.Weekday()))

    // Header row: weekdays + day of month (colorize today)
	w.table.SetCell(0, 0, tview.NewTableCell(" ").SetSelectable(false))
	for d := 0; d < 7; d++ {
		date := weekStart.AddDate(0, 0, d)
        label := fmt.Sprintf("%s %d", date.Weekday().String()[:3], date.Day())
        style := tcell.StyleDefault
        if sameDay(date, time.Now()) {
            style = style.Background(colorTodayBackground).Foreground(colorTodayText)
            label = fmt.Sprintf("[::b]%s[::-]", label)
        }
        cell := tview.NewTableCell(label).SetAlign(tview.AlignCenter).SetStyle(style)
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
            // Highlight today cells with background; keep any text
            if sameDay(date, time.Now()) {
                cell = cell.SetStyle(tcell.StyleDefault.Background(colorTodayBackground).Foreground(colorTodayText))
            }
			w.table.SetCell(r, c+1, cell)
		}
	}

	// Move selection to current day column around midday
	w.syncSelectionToTable()

    // Refresh the mini month below
    w.refreshMiniMonth()
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

func (w *WeekView) refreshMiniMonth() {
    mini := w.miniMonth
    mini.Clear()

    sel := w.uiState.SelectedDate
    loc := sel.Location()
    firstOfMonth := time.Date(sel.Year(), sel.Month(), 1, 0, 0, 0, 0, loc)
    startWeekday := int(firstOfMonth.Weekday())
    firstOfNext := firstOfMonth.AddDate(0, 1, 0)
    daysInMonth := int(firstOfNext.Sub(firstOfMonth).Hours()/24 + 0.5)

    // Compute current week interval [weekStart, weekEnd]
    weekStart := sel.AddDate(0, 0, -int(sel.Weekday()))
    weekEnd := weekStart.AddDate(0, 0, 6)

    // Header row: S M T W T F S
    headers := []string{"S", "M", "T", "W", "T", "F", "S"}
    for c, h := range headers {
        mini.SetCell(0, c, tview.NewTableCell(fmt.Sprintf("[::b]%s", h)).SetAlign(tview.AlignCenter))
    }

    row := 1
    col := startWeekday

    // Leading previous month days
    prevLast := firstOfMonth.AddDate(0, 0, -1)
    for d := startWeekday - 1; d >= 0; d-- {
        day := prevLast.Day() - (startWeekday-1-d)
        date := time.Date(prevLast.Year(), prevLast.Month(), day, 0, 0, 0, 0, loc)
        mini.SetCell(row, d, w.renderMiniCell(date, true, weekStart, weekEnd))
    }

    // Current month
    for day := 1; day <= daysInMonth; day++ {
        date := time.Date(sel.Year(), sel.Month(), day, 0, 0, 0, 0, loc)
        mini.SetCell(row, col, w.renderMiniCell(date, false, weekStart, weekEnd))
        col++
        if col > 6 {
            col = 0
            row++
        }
    }

    // Trailing next month
    for row <= 6 {
        for col <= 6 {
            offsetDays := (row-1)*7 + col - startWeekday - daysInMonth
            date := firstOfNext.AddDate(0, 0, offsetDays)
            mini.SetCell(row, col, w.renderMiniCell(date, true, weekStart, weekEnd))
            col++
        }
        col = 0
        row++
    }
}

func (w *WeekView) renderMiniCell(date time.Time, otherMonth bool, weekStart, weekEnd time.Time) *tview.TableCell {
    label := fmt.Sprintf("%2d", date.Day())
    style := tcell.StyleDefault
    
    // Normalize dates to compare only the date part, not time
    dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
    weekStartOnly := time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
    weekEndOnly := time.Date(weekEnd.Year(), weekEnd.Month(), weekEnd.Day(), 0, 0, 0, 0, weekEnd.Location())
    
    // Check if date is in current week
    inCurrentWeek := !dateOnly.Before(weekStartOnly) && !dateOnly.After(weekEndOnly)
    
    if inCurrentWeek {
        // Current week days in red (fallback)
        style = style.Foreground(tcell.ColorRed)
    } else if otherMonth {
        style = style.Foreground(colorOtherMonthText)
    } else {
        style = style.Foreground(tcell.ColorWhite)
    }
    // Today: use distinct background/text and bold
    if sameDay(date, time.Now()) {
        style = style.Background(colorTodayBackground).Foreground(colorTodayText)
        label = fmt.Sprintf("[::b]%s[::-]", label)
    }
    cell := tview.NewTableCell(label).SetAlign(tview.AlignRight).SetStyle(style)
    return cell
}
