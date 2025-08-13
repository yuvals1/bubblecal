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
    helpText += "  Arrow keys: Move by day\n"
    helpText += "  g: Go to today\n"
    helpText += "\n"
    helpText += "[::b]Week view extras[::-]\n"
    helpText += "  Ctrl+U: Previous week\n"
    helpText += "  Ctrl+D: Next week\n"
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
