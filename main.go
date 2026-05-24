package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "config" {
		_, err := RunSetup()
		if err != nil {
			printDividerError(err.Error())
			os.Exit(1)
		}
		return
	}

	provider := flag.String("provider", "", "AI provider to use (groq or gemini)")
	apiKey := flag.String("key", "", "API key for the provider")
	flag.Parse()

	if flag.NArg() != 1 {
		printDividerError("image path is required as a positional argument")
		flag.Usage()
		os.Exit(1)
	}
	imagePath := flag.Arg(0)

	cfg, err := LoadConfig()
	if err != nil {
		printDividerError(err.Error())
		os.Exit(1)
	}

	p := cfg.Provider
	if *provider != "" {
		p = *provider
	}

	k := cfg.Key
	if *apiKey != "" {
		k = *apiKey
	}

	if p == "" || k == "" {
		cfg, err = RunSetup()
		if err != nil {
			printDividerError(err.Error())
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
		printDividerError(err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "  %s%s%s\n", colorGray, "─────────────────────────────────────────", colorReset)

	running := true
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			if !running {
				fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 80))
				return
			}
			fmt.Fprintf(os.Stderr, "\r  %s%s%s Analyzing %s", colorOrange, frames[i], colorReset, imagePath)
			i = (i + 1) % len(frames)
			time.Sleep(80 * time.Millisecond)
		}
	}()

	prompt, err := GeneratePrompt(p, k, base64Img, mimeType)
	running = false
	time.Sleep(15 * time.Millisecond)

	if err != nil {
		fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 80))
		printDividerError(err.Error())
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "  %s%s✓ Prompt generated%s\n", colorBold, colorGreen, colorReset)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Printf("  %s%s%s\n", colorWhite, prompt, colorReset)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "  %s%s%s\n", colorGray, "─────────────────────────────────────────", colorReset)
	fmt.Fprintf(os.Stderr, "  %s✓ Copied to clipboard%s\n", colorGreen, colorReset)
	fmt.Fprintf(os.Stderr, "  %s%s%s\n", colorGray, "─────────────────────────────────────────", colorReset)

	cmd := exec.Command("pbcopy")
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Run()
}

func printDividerError(msg string) {
	fmt.Fprintf(os.Stderr, "  %s%s%s\n", colorGray, "─────────────────────────────────────────", colorReset)
	fmt.Fprintf(os.Stderr, "  %s✗ Error: %s%s\n", colorRed, msg, colorReset)
	fmt.Fprintf(os.Stderr, "  %s%s%s\n", colorGray, "─────────────────────────────────────────", colorReset)
}
