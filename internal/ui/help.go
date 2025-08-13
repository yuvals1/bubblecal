package ui

import (
	"github.com/rivo/tview"
)

func ShowHelpModal(app *tview.Application, pages *tview.Pages) {
	helpText := ""
	helpText += "[::b]Navigation[::-]\n"
	helpText += "  Arrow keys: Move day\n"
	helpText += "  PgUp/PgDn: Previous/Next month\n"
	helpText += "  g: Go to today\n"
	helpText += "\n"
	helpText += "[::b]General[::-]\n"
	helpText += "  ?: Help  q: Quit\n"

	modal := tview.NewModal().
		SetText(helpText).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("help")
		})

	pages.AddAndSwitchToPage("help", modal, true)
	app.SetFocus(modal)
}
