package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Config struct {
	Provider string `json:"provider"`
	Key      string `json:"key"`
}

func configPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".deprompt", "config.json")
}

func LoadConfig() (Config, error) {
	path := configPath()
	if path == "" {
		return Config{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path := configPath()
	if path == "" {
		return fmt.Errorf("could not determine home directory")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

type setupModel struct {
	state     int
	providers []string
	labels    []string
	cursor    int
	textInput textinput.Model
	cfg       Config
	err       error
	done      bool
}

func initialSetupModel() setupModel {
	ti := textinput.New()
	ti.Placeholder = "paste your API key here..."
	ti.CharLimit = 512
	ti.Width = 60

	return setupModel{
		state:     0,
		providers: []string{"groq", "gemini", "openrouter"},
		labels:    []string{"Groq", "Google Gemini", "OpenRouter"},
		cursor:    0,
		textInput: ti,
	}
}

func (m setupModel) Init() tea.Cmd {
	return nil
}

func (m setupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "enter":
			if m.state == 0 {
				m.cfg.Provider = m.providers[m.cursor]
				m.state = 1
				m.textInput.Focus()
				return m, textinput.Blink
			} else if m.state == 1 {
				key := strings.TrimSpace(m.textInput.Value())
				if key == "" {
					return m, nil
				}
				m.cfg.Key = key
				if err := SaveConfig(m.cfg); err != nil {
					m.err = err
				}
				m.done = true
				return m, tea.Quit
			}

		case "up", "k":
			if m.state == 0 && m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.state == 0 && m.cursor < len(m.labels)-1 {
				m.cursor++
			}
		}
	}

	if m.state == 1 {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m setupModel) View() string {
	if m.done {
		view := "\n"
		if m.err != nil {
			view += errorStyle.Render("  ✗ Error: " + m.err.Error()) + "\n"
		} else {
			view += successStyle.Render("  ✓ You're all set!") + "\n\n"
			view += mutedStyle.Render(fmt.Sprintf("  Config saved to %s", configPath())) + "\n"
			view += mutedStyle.Render("  Run deprompt <image.png> to get started") + "\n"
		}
		view += "\n"
		return view
	}

	view := "\n"
	view += headerStyle.Render("██████  ███████ ██████  ██████   ██████  ███    ███ ██████  ████████") + "\n"
	view += headerStyle.Render("██   ██ ██      ██   ██ ██   ██ ██    ██ ████  ████ ██   ██    ██") + "\n"
	view += headerStyle.Render("██   ██ █████   ██████  ██████  ██    ██ ██ ████ ██ ██████     ██") + "\n"
	view += headerStyle.Render("██   ██ ██      ██      ██   ██ ██    ██ ██  ██  ██ ██         ██") + "\n"
	view += headerStyle.Render("██████  ███████ ██      ██   ██  ██████  ██      ██ ██         ██") + "\n"
	view += "\n"
	view += mutedStyle.Render("  reverse engineer any image into a prompt") + "\n"
	view += mutedStyle.Render("  ─────────────────────────────────────────") + "\n"
	view += titleStyle.Render("  Setup") + "\n"
	view += mutedStyle.Render("  ─────────────────────────────────────────") + "\n\n"

	if m.state == 0 {
		view += titleStyle.Render("  Select your AI provider") + "\n\n"
		for i, label := range m.labels {
			if i == m.cursor {
				view += selectedStyle.Render("  ❯ "+label) + "\n"
			} else {
				view += itemStyle.Render("    "+label) + "\n"
			}
		}
		view += "\n"
		view += helpStyle.Render("  ↑/↓ navigate  •  enter select  •  ctrl+c quit")
	} else if m.state == 1 {
		view += successStyle.Render("  ✓ Provider: "+m.labels[indexOf(m.providers, m.cfg.Provider)]) + "\n\n"
		view += titleStyle.Render("  API Key") + "\n"
		view += mutedStyle.Render("  Enter your API key for "+m.cfg.Provider) + "\n\n"
		view += "  " + m.textInput.View() + "\n\n"
		view += helpStyle.Render("  enter confirm  •  ctrl+c quit")
	}

	view += "\n\n"
	return view
}

func indexOf(slice []string, val string) int {
	for i, v := range slice {
		if v == val {
			return i
		}
	}
	return 0
}

func RunSetup() (Config, error) {
	p := tea.NewProgram(initialSetupModel())
	m, err := p.Run()
	if err != nil {
		return Config{}, err
	}
	model := m.(setupModel)
	if model.err != nil {
		return Config{}, model.err
	}
	if model.cfg.Provider == "" && model.cfg.Key == "" {
		return Config{}, fmt.Errorf("setup cancelled")
	}
	return model.cfg, nil
}
