package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type promptResultMsg struct {
	prompt string
	err    error
}

type mainModel struct {
	spinner  spinner.Model
	loading  bool
	done     bool
	err      error
	prompt   string
	imagePath string
	provider string
	apiKey   string
	base64Img string
	mimeType string
}

func newMainModel(provider, apiKey, base64Img, mimeType, imagePath string) mainModel {
	s := spinner.New()
	s.Style = spinnerStyle
	s.Spinner = spinner.Dot

	return mainModel{
		spinner:   s,
		loading:   true,
		imagePath: imagePath,
		provider:  provider,
		apiKey:    apiKey,
		base64Img: base64Img,
		mimeType:  mimeType,
	}
}

func (m mainModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			prompt, err := GeneratePrompt(m.provider, m.apiKey, m.base64Img, m.mimeType)
			return promptResultMsg{prompt, err}
		},
	)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case promptResultMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.prompt = msg.prompt
			m.done = true
		}
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m mainModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n%s\n\n", errorStyle.Render("  ✗ Error: "+m.err.Error()))
	}
	if m.done {
		return fmt.Sprintf("\n  %s\n\n  %s\n\n", successStyle.Render("✓ Prompt generated"), mutedStyle.Render("✓ Copied to clipboard"))
	}
	return fmt.Sprintf("\n  %s %s\n\n", m.spinner.View(), mutedStyle.Render("Analyzing "+m.imagePath))
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "config":
			_, err := RunSetup()
			if err != nil {
				os.Exit(1)
			}
			return
		case "history":
			ShowHistory()
			return
		}
	}

	provider := flag.String("provider", "", "AI provider to use (groq or gemini)")
	flag.Parse()

	apiKeyEnv := os.Getenv("DEPROMPT_API_KEY")

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: image path is required as a positional argument\n")
		flag.Usage()
		os.Exit(1)
	}
	imagePath := flag.Arg(0)

	cfg, err := LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	p := cfg.Provider
	if *provider != "" {
		p = *provider
	}

	k := cfg.Key
	if apiKeyEnv != "" {
		k = apiKeyEnv
	}

	if p == "" || k == "" {
		cfg, err = RunSetup()
		if err != nil {
			os.Exit(1)
		}
		if p == "" {
			p = cfg.Provider
		}
		if k == "" {
			k = cfg.Key
		}
	}

	base64Img, mimeType, err := EncodeImage(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding image: %v\n", err)
		os.Exit(1)
	}

	model := newMainModel(p, k, base64Img, mimeType, imagePath)
	program := tea.NewProgram(model, tea.WithOutput(os.Stderr))

	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	m := finalModel.(mainModel)
	if m.err != nil {
		os.Exit(1)
	}

	AppendHistory(HistoryEntry{
		Prompt:    m.prompt,
		ImagePath: m.imagePath,
		Provider:  m.provider,
		Timestamp: time.Now(),
	})

	fmt.Println(m.prompt)

	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(m.prompt)
	cmd.Run()
}
