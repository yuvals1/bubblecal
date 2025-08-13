# BubbleCal - Terminal Calendar with Bubble Tea

## Status
✅ **A fully functional terminal calendar application built with Bubble Tea**

BubbleCal offers complete calendar functionality with an intuitive TUI:
- ✅ Month, Week, and Day views
- ✅ Agenda sidebar
- ✅ Keyboard navigation (vim keys + arrows)
- ✅ Focus switching between panes (Tab)
- ✅ View cycling (Space)
- ✅ Event creation modal
- ✅ Event editing and deletion
- ✅ Help modal
- ✅ Visual highlighting (today, selected date, focused pane)
- ✅ Multiple themes (6 color schemes)
- ✅ Persistent preferences
- ✅ Toggleable mini-month view
- ✅ Adjustable agenda position

## Running BubbleCal
```bash
# Build
go build -o bubblecal ./cmd/bubblecal/

# Run
./bubblecal
```

## Architecture

BubbleCal uses Bubble Tea's Elm Architecture:
- The Elm Architecture (Model-Update-View)
- Message passing for state changes
- Functional updates
- Explicit state management
- Lipgloss for terminal styling

## File Structure
```
cmd/
└── bubblecal/      # Main application entry point

internal/
├── tui/            # Bubble Tea UI implementation
│   ├── model.go    # Main app model and update logic
│   ├── month.go    # Month view
│   ├── week.go     # Week view  
│   ├── day.go      # Day view
│   ├── agenda.go   # Agenda sidebar
│   └── modals.go   # Event creation, deletion, help modals
├── storage/        # Event storage layer
├── model/          # Data models
└── config/         # Configuration management
```

## Key Features
- **Clean Architecture** - Separation of concerns with dedicated storage and config layers
- **Persistent Storage** - Events saved in `~/.bubblecal/days/` directory
- **Persistent Preferences** - Settings saved in `~/.bubblecal/config.json`
- **Multiple Themes** - 6 built-in color schemes (Default, Dark, Light, Neon, Solarized, Nord)
- **Flexible Layout** - Toggle agenda position (right/bottom) and mini-month view
- **Smart Time Input** - Automatic time validation and formatting
- **Multi-line Support** - Events display on separate lines when multiple exist
- **Visual Feedback** - Clear focus indicators and selected state highlighting

## Keyboard Shortcuts

### Navigation
- `Space` - Cycle views (Month → Week → Day)
- `Tab` - Toggle focus between calendar and agenda
- `h/l` or `←/→` - Navigate days
- `j/k` or `↑/↓` - Navigate weeks (month) or hours (week/day)
- `Ctrl+U/D` - Previous/next month or week
- `t` or `.` - Go to today

### Events
- `a` - Add new event
- `e` - Edit event (when agenda focused)
- `d` - Delete event (when agenda focused)
- `Ctrl+A` - Toggle all-day event (in event modal)
- `Enter` - Save event (from any field in modal)

### UI Customization
- `m` - Toggle mini-month view (in week view)
- `p` - Toggle agenda position (right/bottom)
- `s` - Cycle through themes

### General
- `?` - Show help
- `q` - Quit

## Configuration

BubbleCal stores its configuration in `~/.bubblecal/config.json`. The configuration includes:
- `show_mini_month` - Whether to show the mini calendar in week view
- `agenda_bottom` - Whether to position the agenda at the bottom
- `theme` - Selected theme index (0-5)