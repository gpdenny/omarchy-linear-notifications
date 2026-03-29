package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
)

func loadAPIKey() (string, error) {
	if key := os.Getenv("LINEAR_API_KEY"); key != "" {
		return key, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "linear-api-key"))
	if err != nil {
		return "", fmt.Errorf("LINEAR_API_KEY not set and ~/.config/linear-api-key not readable")
	}
	key := strings.TrimSpace(string(data))
	if key == "" {
		return "", fmt.Errorf("~/.config/linear-api-key is empty")
	}
	return key, nil
}

func main() {
	waybar := flag.Bool("waybar", false, "Print Waybar JSON and exit")
	demo := flag.Bool("demo", false, "Run with mock data for screenshots")
	flag.Parse()

	if *demo {
		p := tea.NewProgram(newDemoModel(), tea.WithColorProfile(colorprofile.TrueColor))
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	apiKey, err := loadAPIKey()
	if err != nil {
		if *waybar {
			printWaybarError(err.Error())
			return
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	client := NewClient(apiKey)

	if *waybar {
		runWaybar(client)
		return
	}

	p := tea.NewProgram(newModel(client), tea.WithColorProfile(colorprofile.TrueColor))
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if m, ok := result.(model); ok && m.openURL != "" {
		_ = exec.Command("xdg-open", m.openURL).Run()
	}
}
