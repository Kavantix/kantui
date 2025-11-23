<a href="https://www.buymeacoffee.com/kavantix" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="41" width="174"></a>

# Kantui

Kantui is a terminal UI kanban ticketing system that lets you manage tasks quickly without internet

It uses sqlite3 to store the tickets and renders the board using [lipgloss](https://github.com/charmbracelet/lipgloss) and [bubbletea](https://github.com/charmbracelet/bubbletea) to render the TUI

## Installation

### Go install

```sh
go install github.com/Kavantix/kantui/cmd/kantui@latest
```

## Usage

### Open on shortcut (macos)

On macos a tool like [Keyboard Cowboad](https://github.com/zenangst/KeyboardCowboy) can be used to always have access to the kanban board with a single keybinding

Add a shortcut with two steps

1. Open kantui in your terminal
2. Center the kantui window on your screen

## Kitty

In order for the board to be a nice size set the following example config for kitty

```ini
initial_window_width  200c # reasonable default kantui width
initial_window_height 50c # reasonable default kantui height
remember_window_size no # prevent kitty from using old window sizes
macos_quit_when_last_window_closed yes # do not keep kitty open if no window is open
```

In the shell script step of the shortcut add the following script which will launch kantui in a new os window

```bash
/Applications/Kitty.app/Contents/MacOS/kitty \
  --title kantui \
  --single-instance \
  --override=initial_window_width=200c \
  --override=initial_window_height=50c \
  -d "$HOME" \
  kantui
```

The next steps is the following apple script step that centers the window

```scpt
-- get main screen height in points
tell application "Finder"
    set screen_bounds to bounds of window of desktop
    set sw to item 3 of screen_bounds
    set sh to item 4 of screen_bounds
end tell

tell application "System Events"
    tell process "kitty"
        set win to window 1
                -- wait until the window has a real size
        repeat
            try
                set axsize to value of attribute "AXSize" of win
                set ww to item 1 of axsize
                set wh to item 2 of axsize
                if ww > 0 and wh > 0 then exit repeat
            end try
            delay 0.01
        end repeat
        set newX to sw/2 - ww/2
        set newY to sh/2 - wh/2
        set position of win to {newX, newY}
    end tell
end tell
```
