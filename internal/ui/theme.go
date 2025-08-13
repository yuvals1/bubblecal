package ui

import "github.com/gdamore/tcell/v2"

// High-contrast theme tuned for dark terminals.
var (
    colorSelectedBackground = tcell.ColorDarkCyan // strong backdrop
    colorSelectedText       = tcell.ColorBlack    // readable on cyan
    colorTodayUnderline     = tcell.ColorAqua     // cyan underline accent
    colorWeekendText        = tcell.ColorLightSlateGray
    colorOtherMonthText     = tcell.ColorGray
    colorBadge              = tcell.ColorGreen
)
