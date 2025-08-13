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
- **Event Management**: Create, edit, and delete events with persistent storage
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
| `a` | Add event |
| `e` | Edit event (in agenda) |
| `d` | Delete event (in agenda) |
| `?` | Show help |
| `q` | Quit |

### Navigation
- **Move**: `h/j/k/l` or arrow keys
- **Jump to today**: `t` or `.`
- **Switch panes**: `Tab`

### Customization
- **Toggle mini-month**: `m` (in week view)
- **Move agenda**: `p` (right ↔ bottom)
- **Themes**: `s` (6 themes available)

## Data Storage

Events are stored in `~/.bubblecal/days/` as individual files, making them easy to backup or sync. Configuration is saved in `~/.bubblecal/config.json`.

## Architecture

Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) using The Elm Architecture for predictable state management and [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling.
