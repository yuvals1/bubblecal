# BubbleCal

A fast, keyboard-driven terminal calendar built with Bubble Tea.

![Go](https://img.shields.io/badge/Go-1.24.5-blue)
![TUI](https://img.shields.io/badge/TUI-Bubble%20Tea-pink)

## Screenshots

<div align="center">
  <img src="examples/screenshots/Screenshot 2025-08-13 at 20.45.21.png" width="49%" alt="Month View">
  <img src="examples/screenshots/Screenshot 2025-08-13 at 20.37.27.png" width="49%" alt="Week View">
  <img src="examples/screenshots/Screenshot 2025-08-13 at 20.45.28.png" width="49%" alt="Day View">
  <img src="examples/screenshots/Screenshot 2025-08-13 at 20.46.07.png" width="49%" alt="Event Modal">
</div>

## Features

- **Three Views**: Month, Week, and Day views with seamless navigation
- **Event Management**: Create, edit, and delete events with modern modal interface
- **Category System**: Color-coded event categories with persistent configuration
- **KeyJump Navigation**: EasyMotion-style instant date jumping (press `f`)
- **Themes**: 6 built-in color schemes (cycle with `s`)
- **Vim Keys**: Full keyboard navigation with vim-style keybindings
- **Smart Layout**: Toggleable agenda position and mini-month display

## Installation

```bash
go build -o bubblecal ./cmd/bubblecal/
./bubblecal
```

## Quick Start

| Key | Action |
|-----|--------|
| `Space` | Cycle views |
| `f` | **KeyJump mode** - instant date navigation |
| `a` | Add event |
| `e` | Edit event (in agenda) |
| `d` | Delete event (in agenda) |
| `?` | Show help |
| `q` | Quit |

### Navigation
- **Move**: `h/j/k/l` or arrow keys
- **Jump to today**: `t` or `.`
- **KeyJump**: `f` → type letter to jump to any visible date
- **Switch panes**: `Tab`

### Customization
- **Toggle mini-month**: `m` (in week view)
- **Move agenda**: `p` (right ↔ bottom)
- **Themes**: `s` (6 themes available)

## Advanced Features

### KeyJump Navigation
Inspired by vim's EasyMotion, press `f` to enter jump mode where letter overlays appear on all visible dates. Type any letter to instantly teleport to that date - no more arrow key navigation!

### Event Categories
Create and assign color-coded categories to your events. Categories are displayed throughout the interface with their assigned colors and can be easily selected when creating events.

### Modern Event Modal
Beautiful, intuitive event creation interface with:
- Field-by-field navigation
- Visual category selector
- All-day event toggle
- Smart time field handling
- Clear error messages and instructions

## Data Storage

Events are stored in `~/.bubblecal/days/` as individual files, making them easy to backup or sync. Configuration is saved in `~/.bubblecal/config.json`.

### Categories

Event categories are configured in `~/.bubblecal/config.json` with customizable colors:

```json
{
  "categories": [
    {"name": "Work", "color": "#4287f5"},
    {"name": "Personal", "color": "#42f554"},
    {"name": "Health", "color": "#f54242"}
  ]
}
```

## Architecture

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) using The Elm Architecture for predictable state management and [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling.
