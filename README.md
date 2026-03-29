# omarchy-linear-notifications

A [Linear](https://linear.app) notification manager for your terminal. Built for [Omarchy](https://omarchy.org).

## Features

- Split-pane TUI: notification list on the left, detail on the right
- Waybar integration via `--waybar` flag
- Mark notifications as read (`r`) or mark all as read (`R`)
- Open notifications in your browser (`Enter` / `o`)
- Supports issue, pull request, and project notifications

## Install

### From release

Download the latest binary from [Releases](https://github.com/gdenny/omarchy-linear-notifications/releases) and place it on your `PATH`:

```bash
cp omarchy-linear-notifications ~/.local/bin/
```

### From source

Requires Go 1.24+.

```bash
git clone https://github.com/gdenny/omarchy-linear-notifications.git
cd omarchy-linear-notifications
go build -o ~/.local/bin/omarchy-linear-notifications .
```

## Setup

### API key

Create a personal API key at [Linear Settings → API](https://linear.app/settings/api) with **read and write** scope.

Provide it via environment variable or file (checked in order):

1. `LINEAR_API_KEY` environment variable
2. `~/.config/linear-api-key` file

```bash
# Option 1: environment variabled
export LINEAR_API_KEY="lin_api_..."

# Option 2: file
echo "lin_api_..." > ~/.config/linear-api-key
chmod 600 ~/.config/linear-api-key
```

For Waybar to see the variable, set it in your graphical session:

- **Hyprland**: `env = LINEAR_API_KEY,lin_api_...` in `~/.config/hypr/envs.conf`
- **systemd**: `LINEAR_API_KEY=lin_api_...` in `~/.config/environment.d/linear.conf`

### Waybar

Add to `~/.config/waybar/config.jsonc`:

```jsonc
"custom/linear": {
  "exec": "omarchy-linear-notifications --waybar",
  "return-type": "json",
  "interval": 120,
  "hide-empty-text": true,
  "on-click": "omarchy-launch-or-focus-tui omarchy-linear-notifications"
}
```

### Hyprland window rule

To float the TUI window (like `omarchy-vpn`), add to `~/.config/hypr/hyprland.conf`:

```
windowrule = tag +floating-window, match:class org.omarchy.omarchy-linear-notifications
```

## Usage

```bash
# Interactive TUI
omarchy-linear-notifications

# Waybar JSON output
omarchy-linear-notifications --waybar
```

### Keybinds

| Key | Action |
|-----|--------|
| `j` / `k` / `↑` / `↓` | Navigate |
| `Enter` / `o` | Open in browser |
| `r` | Mark as read |
| `R` | Mark all as read |
| `q` / `Esc` | Quit |

## License

MIT
