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
	// Example: " Aug 2025 · Simple TUI Cal    [a] Add  [e] Edit  [d] Delete  [?] Help  [q] Quit"
	return fmt.Sprintf(" [::b]%s %d[::-] · Simple TUI Cal    [yellow][a][::-] Add  [yellow][e][::-] Edit  [yellow][d][::-] Delete  [yellow][?][::-] Help  [yellow][q][::-] Quit",
		now.Month().String()[:3], now.Year())
}
