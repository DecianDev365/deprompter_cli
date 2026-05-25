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
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
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

func selectProvider() string {
	providers := []string{"groq", "gemini", "openrouter"}
	labels := []string{"Groq", "Google Gemini", "OpenRouter"}
	selected := 0

	fd := int(os.Stdin.Fd())
	var oldState syscall.Termios
	syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGETA, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)

	newState := oldState
	newState.Lflag &^= syscall.ECHO | syscall.ICANON
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0
	syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&newState)), 0, 0, 0)
	defer syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&oldState)), 0, 0, 0)

	printBold("Select your AI provider")
	fmt.Println()
	drawMenu(selected, labels)
	fmt.Printf("  %sUse ↑/↓ to navigate, Enter to select%s\n", colorGray, colorReset)

	menuLines := 4 + len(labels)

	for {
		key, err := readKeyRaw()
		if err != nil {
			break
		}

		switch key {
		case "up":
			selected = (selected - 1 + len(providers)) % len(providers)
		case "down":
			selected = (selected + 1) % len(providers)
		case "enter":
			fmt.Printf("\033[%dA\033[J", menuLines)
			return providers[selected]
		default:
			continue
		}

		fmt.Printf("\033[%dA\033[J", menuLines)
		printBold("Select your AI provider")
		fmt.Println()
		drawMenu(selected, labels)
		fmt.Printf("  %sUse ↑/↓ to navigate, Enter to select%s\n", colorGray, colorReset)
	}

	return providers[selected]
}

func drawMenu(selected int, labels []string) {
	for i, label := range labels {
		if i == selected {
			fmt.Printf("  %s > %s%s\n", colorOrange, label, colorReset)
		} else {
			fmt.Printf("     %s\n", label)
		}
	}
	fmt.Println()
}

func readKeyRaw() (string, error) {
	fd := int(os.Stdin.Fd())
	var buf [1]byte
	_, err := os.Stdin.Read(buf[:])
	if err != nil {
		return "", err
	}

	switch buf[0] {
	case '\x1b':
		var ts syscall.Termios
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCGETA, uintptr(unsafe.Pointer(&ts)), 0, 0, 0)
		ts.Cc[syscall.VMIN] = 0
		ts.Cc[syscall.VTIME] = 1
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&ts)), 0, 0, 0)

		extra := make([]byte, 2)
		n, _ := os.Stdin.Read(extra)

		ts.Cc[syscall.VMIN] = 1
		ts.Cc[syscall.VTIME] = 0
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), syscall.TIOCSETA, uintptr(unsafe.Pointer(&ts)), 0, 0, 0)

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
	}

	return "", nil
}
