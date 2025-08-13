package ui

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

func buildStatusBar() *tview.TextView {
	v := tview.NewTextView()
	v.SetDynamicColors(true)
	v.SetTextAlign(tview.AlignLeft)
	v.SetText(renderStatus(time.Now(), time.Now()))
	return v
}

func renderStatus(today, selected time.Time) string {
    return fmt.Sprintf(" ◉ Today: %s · Selected: %s · [m] Month [w] Week · [←/→/↑/↓] Move  [Ctrl+U/D] Week/Month  [g] Today  [?] Help  [q] Quit",
        today.Format("Mon Jan 2"), selected.Format("Mon Jan 2"))
}
