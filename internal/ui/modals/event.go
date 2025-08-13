package modals

import (
	"fmt"
	"simple-tui-cal/internal/model"
	"simple-tui-cal/internal/storage"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ShowNewEventModal displays a form to create a new event
func ShowNewEventModal(app *tview.Application, pages *tview.Pages, date time.Time, onComplete func()) {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(fmt.Sprintf(" New Event - %s ", date.Format("Jan 2, 2006")))
	form.SetTitleAlign(tview.AlignCenter)
	
	// Form fields
	form.AddInputField("Title", "", 40, nil, nil)
	form.AddInputField("Start Time", "09:00", 10, nil, nil)
	form.AddInputField("End Time (optional)", "", 10, nil, nil)
	form.AddCheckbox("All Day", false, nil)
	form.AddInputField("Categories (comma-separated)", "", 30, nil, nil)
	
	// Handle all-day checkbox
	form.GetFormItemByLabel("All Day").(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		startItem := form.GetFormItemByLabel("Start Time").(*tview.InputField)
		endItem := form.GetFormItemByLabel("End Time (optional)").(*tview.InputField)
		if checked {
			startItem.SetDisabled(true)
			endItem.SetDisabled(true)
			startItem.SetText("")
			endItem.SetText("")
		} else {
			startItem.SetDisabled(false)
			endItem.SetDisabled(false)
			startItem.SetText("09:00")
		}
	})
	
	// Add buttons
	form.AddButton("Save", func() {
		// Gather form data
		title := form.GetFormItemByLabel("Title").(*tview.InputField).GetText()
		startTime := form.GetFormItemByLabel("Start Time").(*tview.InputField).GetText()
		endTime := form.GetFormItemByLabel("End Time (optional)").(*tview.InputField).GetText()
		allDay := form.GetFormItemByLabel("All Day").(*tview.Checkbox).IsChecked()
		categoriesStr := form.GetFormItemByLabel("Categories (comma-separated)").(*tview.InputField).GetText()
		
		// Validate
		if strings.TrimSpace(title) == "" {
			showError(app, pages, "Title cannot be empty")
			return
		}
		
		// Create event
		event := &model.Event{
			Title: title,
		}
		
		if allDay {
			event.StartTime = "all-day"
			event.EndTime = ""
		} else {
			if strings.TrimSpace(startTime) == "" {
				showError(app, pages, "Start time is required for timed events")
				return
			}
			event.StartTime = startTime
			event.EndTime = endTime
		}
		
		// Parse categories
		if categoriesStr != "" {
			for _, cat := range strings.Split(categoriesStr, ",") {
				if trimmed := strings.TrimSpace(cat); trimmed != "" {
					event.Categories = append(event.Categories, trimmed)
				}
			}
		}
		
		// Save event
		if err := saveEvent(date, event); err != nil {
			showError(app, pages, fmt.Sprintf("Failed to save event: %v", err))
			return
		}
		
		// Close modal and refresh
		pages.RemovePage("new-event")
		if onComplete != nil {
			onComplete()
		}
	})
	
	form.AddButton("Cancel", func() {
		pages.RemovePage("new-event")
	})
	
	// Handle Escape key to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.RemovePage("new-event")
			return nil
		}
		return event
	})
	
	// Style
	form.SetBackgroundColor(tcell.ColorBlack)
	form.SetFieldBackgroundColor(tcell.ColorDarkBlue)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetButtonBackgroundColor(tcell.ColorDarkCyan)
	form.SetButtonTextColor(tcell.ColorBlack)
	
	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 15, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)
	
	pages.AddAndSwitchToPage("new-event", flex, true)
	app.SetFocus(form)
}

func saveEvent(date time.Time, event *model.Event) error {
	// Load existing events
	events, err := storage.LoadDayEvents(date)
	if err != nil {
		return err
	}
	
	// Add new event
	events = append(events, event)
	
	// Save back to file
	return storage.SaveDayEvents(date, events)
}

// ShowEditEventModal displays a form to edit an existing event
func ShowEditEventModal(app *tview.Application, pages *tview.Pages, date time.Time, event *model.Event, index int, onComplete func()) {
	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(fmt.Sprintf(" Edit Event - %s ", date.Format("Jan 2, 2006")))
	form.SetTitleAlign(tview.AlignCenter)
	
	// Pre-populate form fields
	form.AddInputField("Title", event.Title, 40, nil, nil)
	
	isAllDay := event.IsAllDay()
	startTime := event.StartTime
	if isAllDay {
		startTime = ""
	}
	
	form.AddInputField("Start Time", startTime, 10, nil, nil)
	form.AddInputField("End Time (optional)", event.EndTime, 10, nil, nil)
	form.AddCheckbox("All Day", isAllDay, nil)
	form.AddInputField("Categories (comma-separated)", strings.Join(event.Categories, ","), 30, nil, nil)
	
	// Handle all-day checkbox
	form.GetFormItemByLabel("All Day").(*tview.Checkbox).SetChangedFunc(func(checked bool) {
		startItem := form.GetFormItemByLabel("Start Time").(*tview.InputField)
		endItem := form.GetFormItemByLabel("End Time (optional)").(*tview.InputField)
		if checked {
			startItem.SetDisabled(true)
			endItem.SetDisabled(true)
			startItem.SetText("")
			endItem.SetText("")
		} else {
			startItem.SetDisabled(false)
			endItem.SetDisabled(false)
			if startItem.GetText() == "" {
				startItem.SetText("09:00")
			}
		}
	})
	
	// Set initial disabled state
	if isAllDay {
		form.GetFormItemByLabel("Start Time").(*tview.InputField).SetDisabled(true)
		form.GetFormItemByLabel("End Time (optional)").(*tview.InputField).SetDisabled(true)
	}
	
	// Add buttons
	form.AddButton("Save", func() {
		// Gather form data
		title := form.GetFormItemByLabel("Title").(*tview.InputField).GetText()
		startTime := form.GetFormItemByLabel("Start Time").(*tview.InputField).GetText()
		endTime := form.GetFormItemByLabel("End Time (optional)").(*tview.InputField).GetText()
		allDay := form.GetFormItemByLabel("All Day").(*tview.Checkbox).IsChecked()
		categoriesStr := form.GetFormItemByLabel("Categories (comma-separated)").(*tview.InputField).GetText()
		
		// Validate
		if strings.TrimSpace(title) == "" {
			showError(app, pages, "Title cannot be empty")
			return
		}
		
		// Update event
		event.Title = title
		
		if allDay {
			event.StartTime = "all-day"
			event.EndTime = ""
		} else {
			if strings.TrimSpace(startTime) == "" {
				showError(app, pages, "Start time is required for timed events")
				return
			}
			event.StartTime = startTime
			event.EndTime = endTime
		}
		
		// Parse categories
		event.Categories = nil
		if categoriesStr != "" {
			for _, cat := range strings.Split(categoriesStr, ",") {
				if trimmed := strings.TrimSpace(cat); trimmed != "" {
					event.Categories = append(event.Categories, trimmed)
				}
			}
		}
		
		// Save updated events
		if err := updateEvent(date, index, event); err != nil {
			showError(app, pages, fmt.Sprintf("Failed to update event: %v", err))
			return
		}
		
		// Close modal and refresh
		pages.RemovePage("edit-event")
		if onComplete != nil {
			onComplete()
		}
	})
	
	form.AddButton("Cancel", func() {
		pages.RemovePage("edit-event")
	})
	
	// Handle Escape key to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			pages.RemovePage("edit-event")
			return nil
		}
		return event
	})
	
	// Style
	form.SetBackgroundColor(tcell.ColorBlack)
	form.SetFieldBackgroundColor(tcell.ColorDarkBlue)
	form.SetFieldTextColor(tcell.ColorWhite)
	form.SetLabelColor(tcell.ColorWhite)
	form.SetButtonBackgroundColor(tcell.ColorDarkCyan)
	form.SetButtonTextColor(tcell.ColorBlack)
	
	// Center the form
	flex := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 15, 1, true).
			AddItem(nil, 0, 1, false), 60, 1, true).
		AddItem(nil, 0, 1, false)
	
	pages.AddAndSwitchToPage("edit-event", flex, true)
	app.SetFocus(form)
}

// ShowDeleteConfirmModal shows a confirmation dialog for deleting an event
func ShowDeleteConfirmModal(app *tview.Application, pages *tview.Pages, date time.Time, event *model.Event, index int, onComplete func()) {
	text := fmt.Sprintf("Delete this event?\n\n%s\n%s", 
		event.FormatEventLine(),
		date.Format("Monday, January 2, 2006"))
	
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Delete", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Delete" {
				if err := deleteEvent(date, index); err != nil {
					showError(app, pages, fmt.Sprintf("Failed to delete event: %v", err))
				} else if onComplete != nil {
					onComplete()
				}
			}
			pages.RemovePage("delete-confirm")
		})
	
	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetTextColor(tcell.ColorWhite)
	modal.SetButtonBackgroundColor(tcell.ColorDarkRed)
	modal.SetButtonTextColor(tcell.ColorWhite)
	
	pages.AddAndSwitchToPage("delete-confirm", modal, true)
	app.SetFocus(modal)
}

func updateEvent(date time.Time, index int, updatedEvent *model.Event) error {
	// Load existing events
	events, err := storage.LoadDayEvents(date)
	if err != nil {
		return err
	}
	
	// Validate index
	if index < 0 || index >= len(events) {
		return fmt.Errorf("invalid event index")
	}
	
	// Update the event
	events[index] = updatedEvent
	
	// Save back to file
	return storage.SaveDayEvents(date, events)
}

func deleteEvent(date time.Time, index int) error {
	// Load existing events
	events, err := storage.LoadDayEvents(date)
	if err != nil {
		return err
	}
	
	// Validate index
	if index < 0 || index >= len(events) {
		return fmt.Errorf("invalid event index")
	}
	
	// Remove the event
	events = append(events[:index], events[index+1:]...)
	
	// Save back to file
	return storage.SaveDayEvents(date, events)
}

func showError(app *tview.Application, pages *tview.Pages, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("error")
		})
	
	modal.SetBackgroundColor(tcell.ColorBlack)
	modal.SetTextColor(tcell.ColorRed)
	
	pages.AddAndSwitchToPage("error", modal, true)
	app.SetFocus(modal)
}