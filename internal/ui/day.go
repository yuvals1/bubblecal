package ui

import (
	"fmt"
	"simple-tui-cal/internal/storage"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DayView struct {
	uiState *UIState
	table   *tview.Table
}

func NewDayView(state *UIState) *DayView {
	t := tview.NewTable()
	t.SetBorders(true)
	t.SetBorder(true)
	t.SetFixed(1, 1) // Fixed header row and time column
	t.SetSelectable(true, false) // Can select rows (hours), not columns
	t.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack).Bold(true))
	
	dv := &DayView{uiState: state, table: t}
	dv.Refresh()
	return dv
}

func (d *DayView) Primitive() tview.Primitive { return d.table }

// GetSelectedHour returns the hour of the currently selected row
func (d *DayView) GetSelectedHour() string {
	row, _ := d.table.GetSelection()
	// Row 0 is header, rows 1+ are hours starting at 6am
	if row > 0 {
		hour := 6 + (row - 1)  // Start hour is 6, row 1 = 6am
		if hour >= 0 && hour <= 23 {
			return fmt.Sprintf("%02d:00", hour)
		}
	}
	return ""
}

// SetFocused updates the visual style when day view gains/loses focus
func (d *DayView) SetFocused(focused bool) {
	if focused {
		d.table.SetBorderColor(tcell.ColorYellow)
		d.table.SetTitleColor(tcell.ColorYellow)
	} else {
		d.table.SetBorderColor(tcell.ColorDefault)
		d.table.SetTitleColor(tcell.ColorDefault)
	}
}

func (d *DayView) Refresh() {
	d.table.Clear()
	
	date := d.uiState.SelectedDate
	
	// Set title with the current date
	d.table.SetTitle(fmt.Sprintf(" %s ", date.Format("Monday, January 2, 2006")))
	
	// Header row
	d.table.SetCell(0, 0, tview.NewTableCell("Time").
		SetAlign(tview.AlignCenter).
		SetSelectable(false).
		SetStyle(tcell.StyleDefault.Bold(true)))
	d.table.SetCell(0, 1, tview.NewTableCell("Events").
		SetAlign(tview.AlignCenter).
		SetSelectable(false).
		SetStyle(tcell.StyleDefault.Bold(true)))
	
	// Load events for this day
	events, _ := storage.LoadDayEvents(date)
	
	// Create a map of hour -> events for quick lookup
	hourEvents := make(map[int][]string)
	var allDayEvents []string
	
	for _, evt := range events {
		if evt.IsAllDay() {
			allDayEvents = append(allDayEvents, evt.Title)
		} else {
			// Parse start hour
			var hour int
			if _, err := fmt.Sscanf(evt.StartTime, "%d:", &hour); err == nil {
				eventText := evt.Title
				if evt.EndTime != "" {
					eventText = fmt.Sprintf("%s-%s %s", evt.StartTime, evt.EndTime, evt.Title)
				} else {
					eventText = fmt.Sprintf("%s %s", evt.StartTime, evt.Title)
				}
				if len(evt.Categories) > 0 {
					eventText += fmt.Sprintf(" [%s]", strings.Join(evt.Categories, ","))
				}
				hourEvents[hour] = append(hourEvents[hour], eventText)
			}
		}
	}
	
	// Show all-day events at the top if any
	if len(allDayEvents) > 0 {
		d.table.SetCell(1, 0, tview.NewTableCell("All Day").
			SetAlign(tview.AlignRight).
			SetSelectable(false).
			SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGray)))
		d.table.SetCell(1, 1, tview.NewTableCell(strings.Join(allDayEvents, " | ")).
			SetExpansion(1).
			SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGreen)))
	}
	
	// Hour rows from 6:00 to 22:00 (covering most active hours)
	startHour := 6
	endHour := 22
	startRow := 2 // Row 2 because row 0 is header, row 1 might be all-day events
	if len(allDayEvents) == 0 {
		startRow = 1
	}
	
	for r, h := startRow, startHour; h <= endHour; r, h = r+1, h+1 {
		// Time column
		timeCell := tview.NewTableCell(fmt.Sprintf("%02d:00", h)).
			SetAlign(tview.AlignRight).
			SetSelectable(false)
		
		// Style current hour
		if sameDay(date, time.Now()) && h == time.Now().Hour() {
			timeCell.SetStyle(tcell.StyleDefault.Background(colorTodayBackground).Foreground(colorTodayText))
		}
		
		d.table.SetCell(r, 0, timeCell)
		
		// Events column
		eventsText := ""
		if events, ok := hourEvents[h]; ok {
			eventsText = strings.Join(events, " | ")
		}
		
		eventCell := tview.NewTableCell(eventsText).SetExpansion(1)
		
		// Highlight today's current hour
		if sameDay(date, time.Now()) && h == time.Now().Hour() {
			eventCell.SetStyle(tcell.StyleDefault.Background(colorTodayBackground).Foreground(colorTodayText))
		}
		
		d.table.SetCell(r, 1, eventCell)
	}
	
	// Select current hour if today, otherwise select first hour
	if sameDay(date, time.Now()) {
		currentHour := time.Now().Hour()
		if currentHour >= startHour && currentHour <= endHour {
			d.table.Select(startRow+currentHour-startHour, 1)
		} else {
			d.table.Select(startRow, 1)
		}
	} else {
		d.table.Select(startRow, 1)
	}
}