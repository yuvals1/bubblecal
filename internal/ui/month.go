package ui

import (
	"fmt"
	"simple-tui-cal/internal/storage"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type MonthView struct {
	uiState *UIState
	table   *tview.Table
}

func NewMonthView(state *UIState) *MonthView {
	t := tview.NewTable()
	t.SetBorders(true)
	t.SetBorder(true).SetTitle("Monthly")  // Add border and title
	t.SetFixed(1, 0)
    // Allow the table to display a strong selection highlight.
    t.SetSelectable(true, true)
    t.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack).Bold(true))
	mv := &MonthView{uiState: state, table: t}
	mv.buildStatic()
	mv.Refresh()
    // Keep UI state in sync whenever selection changes via keyboard (including hjkl)
    t.SetSelectionChangedFunc(func(row, column int) {
        if cell := t.GetCell(row, column); cell != nil {
            if ref, ok := cell.GetReference().(time.Time); ok {
                mv.uiState.SelectedDate = ref
            }
        }
    })
	return mv
}

func (m *MonthView) Primitive() tview.Primitive { return m.table }

// SetFocused updates the visual style when month view gains/loses focus
func (m *MonthView) SetFocused(focused bool) {
	if focused {
		m.table.SetBorderColor(tcell.ColorYellow)
		m.table.SetTitleColor(tcell.ColorYellow)
	} else {
		m.table.SetBorderColor(tcell.ColorDefault)
		m.table.SetTitleColor(tcell.ColorDefault)
	}
}

func (m *MonthView) buildStatic() {
	// Header row Sun..Sat
	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
    for c, name := range weekdays {
        cell := tview.NewTableCell(fmt.Sprintf("[::b]%s", name)).
            SetAlign(tview.AlignCenter).
            SetExpansion(1)
        m.table.SetCell(0, c, cell)
    }
}

func (m *MonthView) Refresh() {
	m.table.Clear()
	m.buildStatic()

	now := m.uiState.SelectedDate
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startWeekday := int(firstOfMonth.Weekday()) // 0=Sun
	// Determine how many days in month
	firstOfNext := firstOfMonth.AddDate(0, 1, 0)
	daysInMonth := int(firstOfNext.Sub(firstOfMonth).Hours()/24 + 0.5)

	// Previous month tail
	prevLast := firstOfMonth.AddDate(0, 0, -1)
	prevDays := int(firstOfMonth.Sub(time.Date(firstOfMonth.Year(), firstOfMonth.Month(), 0, 0, 0, 0, 0, now.Location())).Hours()/24 + 0.5)
	_ = prevDays // not used directly, but keep for clarity

	row := 1
	col := startWeekday

	// Fill leading days from previous month
	for d := startWeekday - 1; d >= 0; d-- {
		day := prevLast.Day() - (startWeekday-1-d)
		date := time.Date(prevLast.Year(), prevLast.Month(), day, 0, 0, 0, 0, now.Location())
		m.table.SetCell(row, d, m.renderDayCell(date, true))
	}

	// Fill current month
	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location())
		m.table.SetCell(row, col, m.renderDayCell(date, false))
		col++
		if col > 6 {
			col = 0
			row++
		}
	}

	// Fill trailing days from next month to complete 6 rows
	for row <= 6 {
		for col <= 6 {
			// compute date offset from firstOfNext
			offsetDays := (row-1)*7 + col - startWeekday - daysInMonth
			date := firstOfNext.AddDate(0, 0, offsetDays)
			m.table.SetCell(row, col, m.renderDayCell(date, true))
			col++
		}
		col = 0
		row++
	}

    // Move the table's selection to the selected date for a clear hover effect
    m.syncSelectionToTable()
}

func (m *MonthView) renderDayCell(date time.Time, otherMonth bool) *tview.TableCell {
	label := fmt.Sprintf("%d", date.Day())
	style := tcell.StyleDefault

	// Weekend dim
	if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		style = style.Foreground(colorWeekendText)
	}
	if otherMonth {
		style = style.Foreground(colorOtherMonthText)
	}

	// Selected highlight
	if sameDay(date, m.uiState.SelectedDate) {
		label = fmt.Sprintf("[::b]%d[::-]", date.Day())
		style = style.Background(colorSelectedBackground).Foreground(colorSelectedText)
	}

    // Today: emphasize with background; keep foreground dynamic so inline tags (red T) work
    if sameDay(date, time.Now()) {
        style = style.Background(colorTodayBackground).Foreground(tcell.ColorDefault)
        // Ensure the day number renders white on the today background
        label = fmt.Sprintf("[white]%s[::-]", label)
    }

	// Show real event count
	badge := ""
	if events, err := storage.LoadDayEvents(date); err == nil && len(events) > 0 {
		badge = fmt.Sprintf(" [green]●%d[::-]", len(events))
	}

    // Add a red "T" marker to clearly denote today
    if sameDay(date, time.Now()) {
        label += " [red]T[::-]"
    }

    cell := tview.NewTableCell(label + badge)
    cell.SetAlign(tview.AlignLeft).SetStyle(style).SetExpansion(1)
    cell.SetReference(date)
    return cell
}

// syncSelectionToTable finds the cell that matches the current selected date
// and moves the tview.Table selection there so it gets the selected style.
func (m *MonthView) syncSelectionToTable() {
    // We know the month grid is 6 rows x 7 columns starting at row 1
    for r := 1; r <= 6; r++ {
        for c := 0; c <= 6; c++ {
            if cell := m.table.GetCell(r, c); cell != nil {
                if ref, ok := cell.GetReference().(time.Time); ok {
                    if sameDay(ref, m.uiState.SelectedDate) {
                        m.table.Select(r, c)
                        return
                    }
                }
            }
        }
    }
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}

