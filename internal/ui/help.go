package ui

import (
    "fmt"
    
    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

func ShowHelpModal(app *tview.Application, pages *tview.Pages, currentView ViewKind, focusedPane PaneKind) {
    helpText := ""
    
    // View-specific title
    viewName := ""
    switch currentView {
    case ViewMonth:
        viewName = "Month View"
    case ViewWeek:
        viewName = "Week View"
    case ViewDay:
        viewName = "Day View"
    }
    helpText += fmt.Sprintf("[::b]%s Help[::-]\n\n", viewName)
    
    // View switching
    helpText += "[::b]Views[::-]\n"
    helpText += "  Space: Cycle views (Month→Week→Day)\n"
    helpText += "\n"
    
    // Navigation (context-specific)
    helpText += "[::b]Navigation[::-]\n"
    switch currentView {
    case ViewMonth:
        helpText += "  h/l or ←/→: Previous/next day\n"
        helpText += "  j/k or ↑/↓: Previous/next week\n"
        helpText += "  Ctrl+U/D: Previous/next month\n"
    case ViewWeek:
        helpText += "  h/l or ←/→: Previous/next day\n"
        helpText += "  j/k or ↑/↓: Move between hours\n"
        helpText += "  Ctrl+U/D: Previous/next week\n"
    case ViewDay:
        helpText += "  h/l or ←/→: Previous/next day\n"
        helpText += "  j/k or ↑/↓: Move between hours\n"
        helpText += "  Ctrl+U/D: Previous/next day\n"
    }
    helpText += "  Tab: Toggle focus (calendar ↔ agenda)\n"
    switch currentView {
    case ViewMonth:
        helpText += "  t: Go to today\n"
    case ViewWeek:
        helpText += "  t: Go to today (and current hour)\n"
    case ViewDay:
        helpText += "  t: Go to today (and current hour)\n"
    }
    helpText += "\n"
    
    // Events section
    helpText += "[::b]Events[::-]\n"
    if currentView == ViewWeek || currentView == ViewDay {
        helpText += "  a: Add event (at selected hour)\n"
    } else {
        helpText += "  a: Add event\n"
    }
    
    if focusedPane == PaneAgenda {
        helpText += "  e: Edit selected event\n"
        helpText += "  d: Delete selected event\n"
    } else {
        helpText += "  (Focus agenda to edit/delete)\n"
    }
    helpText += "\n"
    
    helpText += "[::b]General[::-]\n"
    helpText += "  ?: Help    q: Quit\n"

    modal := tview.NewModal().
        SetText(helpText).
        AddButtons([]string{"Close"}).
        SetDoneFunc(func(buttonIndex int, buttonLabel string) {
            pages.RemovePage("help")
        })

    modal.SetBackgroundColor(tcell.ColorBlack)
    modal.SetTextColor(tcell.ColorWhite)
    modal.SetButtonBackgroundColor(tcell.ColorDarkCyan)
    modal.SetButtonTextColor(tcell.ColorBlack)

	pages.AddAndSwitchToPage("help", modal, true)
	app.SetFocus(modal)
}
