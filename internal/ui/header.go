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
	// Example: " Aug 2025 · Simple TUI Cal    [n] New  [?] Help  [q] Quit"
	return fmt.Sprintf(" [::b]%s %d[::-] · Simple TUI Cal    [::d][n] New  [?] Help  [q] Quit[::-]",
		now.Month().String()[:3], now.Year())
}
