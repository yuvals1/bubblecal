package ui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

func buildHeader() *tview.TextView {
	v := tview.NewTextView()
	v.SetDynamicColors(true)
	v.SetTextAlign(tview.AlignLeft)
	v.SetText(renderHeader(time.Now()))
	return v
}

func renderHeader(now time.Time) string {
	return fmt.Sprintf(" [::b]%s %d[::-] · Simple TUI Cal",
		now.Month().String()[:3], now.Year())
}
