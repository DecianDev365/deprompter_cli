package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func selectProvider() string {
	providers := []string{"groq", "gemini", "openrouter"}
	labels := []string{"Groq", "Google Gemini", "OpenRouter"}

	reader := bufio.NewReader(os.Stdin)

	for {
		printBold("Select your AI provider")
		fmt.Println()
		for i, label := range labels {
			fmt.Printf("  %d. %s\n", i+1, label)
		}
		fmt.Println()
		fmt.Printf("  %s›%s ", colorOrange, colorReset)

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			return providers[0]
		case "2":
			return providers[1]
		case "3":
			return providers[2]
		}
	}
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
	reader := bufio.NewReader(os.Stdin)
	b, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	switch b {
	case '\x1b':
		extra := make([]byte, 2)
		n, _ := reader.Read(extra)
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
