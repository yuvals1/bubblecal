package ui

import (
	"fmt"
	"simple-tui-cal/internal/model"
	"simple-tui-cal/internal/storage"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AgendaView struct {
	uiState *UIState
	list    *tview.List
	events  []*model.Event  // Store current events for reference
}

func NewAgendaView(state *UIState) *AgendaView {
	l := tview.NewList()
    l.ShowSecondaryText(false)
    l.SetBorder(true).SetTitle("Agenda")
    l.SetWrapAround(false)
	
	// Add vim key bindings for the list
	l.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			// Move down
			current := l.GetCurrentItem()
			if current < l.GetItemCount()-1 {
				l.SetCurrentItem(current + 1)
			}
			return nil
		case 'k':
			// Move up
			current := l.GetCurrentItem()
			if current > 0 {
				l.SetCurrentItem(current - 1)
			}
			return nil
		}
		return event
	})
	
	ag := &AgendaView{uiState: state, list: l}
	ag.Refresh()
	return ag
}

func (a *AgendaView) Primitive() tview.Primitive { return a.list }

// SetFocused updates the visual style when agenda gains/loses focus
func (a *AgendaView) SetFocused(focused bool) {
	if focused {
		a.list.SetBorderColor(tcell.ColorYellow)
		a.list.SetTitleColor(tcell.ColorYellow)
	} else {
		a.list.SetBorderColor(tcell.ColorDefault)
		a.list.SetTitleColor(tcell.ColorDefault)
	}
}

func (a *AgendaView) Refresh() {
	a.list.Clear()
	date := a.uiState.SelectedDate
	
	// Load real events from file
	events, err := storage.LoadDayEvents(date)
	if err != nil {
		a.list.AddItem(fmt.Sprintf("Error loading events: %v", err), "", 0, nil)
		a.events = nil
		return
	}
	
	// Store events for reference
	a.events = events
	
	if len(events) == 0 {
		a.list.AddItem("No events scheduled", "", 0, nil)
		return
	}
	
	for i, evt := range events {
		var timeStr string
		var label string
		
		if evt.IsAllDay() {
			// Green color for "All day" text
			label = fmt.Sprintf("[green]All day[-] %s", evt.Title)
		} else {
			if evt.EndTime != "" {
				timeStr = fmt.Sprintf("%s-%s", evt.StartTime, evt.EndTime)
			} else {
				timeStr = evt.StartTime
			}
			label = fmt.Sprintf("%s %s", timeStr, evt.Title)
		}
		
		a.list.AddItem(label, "", rune('1'+i), nil)
	}
}

// GetEventsForDate loads events from storage (replacing old mockEventsFor)
func GetEventsForDate(date time.Time) ([]*model.Event, error) {
	return storage.LoadDayEvents(date)
}

// GetSelectedEvent returns the currently selected event, or nil if none
func (a *AgendaView) GetSelectedEvent() *model.Event {
	if a.events == nil || len(a.events) == 0 {
		return nil
	}
	
	index := a.list.GetCurrentItem()
	if index < 0 || index >= len(a.events) {
		return nil
	}
	
	return a.events[index]
}

// GetSelectedIndex returns the index of the currently selected event
func (a *AgendaView) GetSelectedIndex() int {
	return a.list.GetCurrentItem()
}
