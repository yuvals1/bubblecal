package ui

import "github.com/gdamore/tcell/v2"

// High-contrast theme tuned for dark terminals.
var (
    colorSelectedBackground = tcell.ColorDarkCyan // strong backdrop
    colorSelectedText       = tcell.ColorBlack    // readable on cyan
    colorTodayUnderline     = tcell.ColorAqua     // cyan underline accent (legacy)
    colorTodayBackground    = tcell.ColorDarkBlue // distinct background for today
    colorTodayText          = tcell.ColorWhite    // readable on dark blue
    colorWeekendText        = tcell.ColorLightSlateGray
    colorOtherMonthText     = tcell.ColorGray
    colorBadge              = tcell.ColorGreen
)
