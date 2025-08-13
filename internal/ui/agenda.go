package ui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type AgendaView struct {
	uiState *UIState
	list    *tview.List
}

func NewAgendaView(state *UIState) *AgendaView {
	l := tview.NewList()
    l.ShowSecondaryText(false)
    l.SetBorder(true).SetTitle("Agenda")
    l.SetWrapAround(false)
	ag := &AgendaView{uiState: state, list: l}
	ag.Refresh()
	return ag
}

func (a *AgendaView) Primitive() tview.Primitive { return a.list }

func (a *AgendaView) Refresh() {
	a.list.Clear()
	// Mock entries
	date := a.uiState.SelectedDate
	for i, evt := range mockEventsFor(date) {
		a.list.AddItem(fmt.Sprintf("%s %s", evt.Time.Format("15:04"), evt.Title), "", rune('1'+i), nil)
	}
}

type MockEvent struct {
	Time  time.Time
	Title string
}

func mockEventsFor(date time.Time) []MockEvent {
	// Deterministic mock: events on even days
	var res []MockEvent
	if date.Day()%2 == 0 {
		res = append(res, MockEvent{Time: time.Date(date.Year(), date.Month(), date.Day(), 9, 30, 0, 0, date.Location()), Title: "Standup"})
		res = append(res, MockEvent{Time: time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location()), Title: "Lunch"})
		res = append(res, MockEvent{Time: time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location()), Title: "Gym"})
	}
	return res
}
