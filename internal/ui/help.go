package ui

import (
    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

func ShowHelpModal(app *tview.Application, pages *tview.Pages) {
    helpText := ""
    helpText += "[::b]Views[::-]\n"
    helpText += "  m: Month view\n"
    helpText += "  w: Week view\n"
    helpText += "\n"
    helpText += "[::b]Navigation[::-]\n"
    helpText += "  h/l or ←/→: Move by day\n"
    helpText += "  j/k or ↑/↓: Month: ±1 week, Week: ±1 hour\n"
    helpText += "  Tab: Toggle focus (calendar ↔ agenda)\n"
    helpText += "  g: Go to today\n"
    helpText += "\n"
    helpText += "[::b]Paging[::-]\n"
    helpText += "  Ctrl+U: Prev wk/mo\n"
    helpText += "  Ctrl+D: Next wk/mo\n"
    helpText += "\n"
    helpText += "[::b]Events[::-]\n"
    helpText += "  a: Add new event\n"
    helpText += "  e: Edit selected event\n"
    helpText += "  d: Delete selected event\n"
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
