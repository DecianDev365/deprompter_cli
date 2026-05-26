package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
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

func RunSetup() (Config, error) {
	reader := bufio.NewReader(os.Stdin)

	printOrange("██████  ███████ ██████  ██████   ██████  ███    ███ ██████  ████████")
	printOrange("██   ██ ██      ██   ██ ██   ██ ██    ██ ████  ████ ██   ██    ██")
	printOrange("██   ██ █████   ██████  ██████  ██    ██ ██ ████ ██ ██████     ██")
	printOrange("██   ██ ██      ██      ██   ██ ██    ██ ██  ██  ██ ██         ██")
	printOrange("██████  ███████ ██      ██   ██  ██████  ██      ██ ██         ██")
	fmt.Println()
	printGray("reverse engineer any image into a prompt")
	printDivider()
	printBold("Setup")
	printDivider()
	fmt.Println()
	provider := selectProvider()

	fmt.Printf("  %s✓%s %sProvider   %s%s%s\n", colorGreen, colorReset, colorGray, colorReset, colorOrange, provider)
	fmt.Println()
	printBold("API Key")
	fmt.Printf("  %s›%s ", colorOrange, colorReset)
	var key string
	for {
		key, _ = reader.ReadString('\n')
		key = strings.TrimSpace(key)
		if key == "" {
			fmt.Printf("  %sAPI key cannot be empty. Please enter your API key.%s\n", colorRed, colorReset)
			fmt.Printf("  %s›%s ", colorOrange, colorReset)
			continue
		}
		break
	}

	fmt.Printf("  %s✓%s %sAPI Key    %s%s%s\n", colorGreen, colorReset, colorGray, colorReset, colorOrange, "saved")
	fmt.Println()

	cfg := Config{Provider: provider, Key: key}
	if err := SaveConfig(cfg); err != nil {
		return Config{}, err
	}

	printDivider()
	fmt.Printf("  %s%s✓ You're all set!%s\n", colorBold, colorGreen, colorReset)
	fmt.Println()
	fmt.Printf("  %sConfig saved to %s%s%s\n", colorGray, colorOrange, configPath(), colorReset)
	fmt.Printf("  %sRun deprompt image.png to get started%s\n", colorGray, colorReset)
	printDivider()

	return cfg, nil
}

var stdinReader = bufio.NewReader(os.Stdin)

func rawMode() (func(), error) {
	var oldState syscall.Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCGETA, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0); err != 0 {
		return nil, fmt.Errorf("terminal not available")
	}

	newState := oldState
	newState.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	newState.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	newState.Cflag &^= syscall.CSIZE | syscall.PARENB
	newState.Cflag |= syscall.CS8
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCSETA, uintptr(unsafe.Pointer(&newState)), 0, 0, 0); err != 0 {
		return nil, fmt.Errorf("failed to set raw mode")
	}

	return func() {
		syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCSETA, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)
		fmt.Print("\033[?25h")
	}, nil
}

func selectProvider() string {
	providers := []string{"groq", "gemini", "openrouter"}
	labels := []string{"Groq", "Google Gemini", "OpenRouter"}
	selected := 0

	restore, err := rawMode()
	if err != nil {
		fmt.Printf("  %s%s%s\n", colorRed, "error: terminal required for interactive setup", colorReset)
		os.Exit(1)
	}
	defer restore()

	printBold("Select your AI provider")
	fmt.Println()
	drawMenu(selected, labels)
	fmt.Print("\033[?25l")

	for {
		key, err := readKeyRaw()
		if err != nil {
			continue
		}

		switch key {
		case "up":
			if selected > 0 {
				selected--
			}
		case "down":
			if selected < len(labels)-1 {
				selected++
			}
		case "enter":
			fmt.Print("\033[?25h")
			return providers[selected]
		case "ctrl_c":
			fmt.Print("\033[?25h")
			os.Exit(1)
		}

		fmt.Printf("\033[%dA", len(labels)+1)
		drawMenu(selected, labels)
	}
}

func drawMenu(selected int, labels []string) {
	for i, label := range labels {
		fmt.Print("\033[K")
		if i == selected {
			fmt.Printf("  %s> %s%s\n", colorOrange, label, colorReset)
		} else {
			fmt.Printf("    %s\n", label)
		}
	}
	fmt.Print("\033[K")
	fmt.Println()
}

func readKeyRaw() (string, error) {
	b, err := stdinReader.ReadByte()
	if err != nil {
		return "", err
	}

	switch b {
	case '\x1b':
		extra := make([]byte, 2)
		n, _ := stdinReader.Read(extra)
		if n >= 2 && extra[0] == '[' {
			switch extra[1] {
			case 'A':
				return "up", nil
			case 'B':
				return "down", nil
			}
		}
		return "", nil

	case '\r', '\n':
		return "enter", nil

	case '\x03':
		return "ctrl_c", nil
	}

	return "", nil
}
