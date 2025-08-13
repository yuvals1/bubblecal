# Migration from tview to Bubble Tea

## Status
✅ **Migration Complete!** The Bubble Tea implementation now has full feature parity with improvements:
- ✅ Month, Week, and Day views
- ✅ Agenda sidebar
- ✅ Keyboard navigation (vim keys + arrows)
- ✅ Focus switching between panes (Tab)
- ✅ View cycling (Space)
- ✅ Event creation modal
- ✅ Event editing and deletion
- ✅ Help modal
- ✅ Visual highlighting (today, selected date, focused pane)

## Running the Bubble Tea Version
```bash
# Build
go build -o simple-tui-cal-tea ./cmd/simple-tui-cal-tea/

# Run
./simple-tui-cal-tea
```

## Architecture Changes

### tview (Old)
- Widget-based with pre-built components
- Event callbacks
- Imperative updates
- Automatic focus management

### Bubble Tea (New)
- The Elm Architecture (Model-Update-View)
- Message passing
- Functional updates
- Manual state management

## File Structure
```
internal/
├── ui/          # Original tview implementation
└── tui/         # New Bubble Tea implementation
    ├── model.go    # Main app model and update logic
    ├── month.go    # Month view
    ├── week.go     # Week view  
    ├── day.go      # Day view
    ├── agenda.go   # Agenda sidebar
    └── modals.go   # Event creation, deletion, help modals
```

## Key Improvements
1. **Better testability** - Pure functions and separated concerns
2. **Predictable state** - All state changes go through Update()
3. **Customizable rendering** - Complete control over output
4. **Modern patterns** - Follows The Elm Architecture

## Improvements Over tview Version
- ✅ Better border rendering with proper focus indicators
- ✅ Improved event saving/updating using storage layer properly
- ✅ Fixed agenda scrolling and selection
- ✅ Better month view cell sizing and alignment
- ✅ Cleaner week/day view layouts
- ✅ Time validation (HH:MM format)
- ✅ Red "T" marker for today
- ✅ Proper event editing (updates existing events correctly)

## Keyboard Shortcuts
- `Space` - Cycle views (Month → Week → Day)
- `Tab` - Toggle focus between calendar and agenda
- `h/l` or `←/→` - Navigate days
- `j/k` or `↑/↓` - Navigate weeks (month) or hours (week/day)
- `Ctrl+U/D` - Previous/next month or week
- `t` or `.` - Go to today
- `a` - Add new event
- `e` - Edit event (when agenda focused)
- `d` - Delete event (when agenda focused)
- `?` - Show help
- `q` - Quit

## Known Differences from tview Version
1. Forms use basic text inputs instead of tview's form widgets
2. Table borders are drawn manually with box characters
3. Modal stacking is handled manually
4. No mouse support yet (can be added with Bubble Tea's mouse events)