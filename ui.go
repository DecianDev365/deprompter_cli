package main

import "fmt"

const (
	colorOrange = "\033[38;5;208m"
	colorWhite  = "\033[37m"
	colorGray   = "\033[2m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorReset  = "\033[0m"
)

func printOrange(text string) {
	fmt.Printf("  %s%s%s\n", colorOrange, text, colorReset)
}

func printGray(text string) {
	fmt.Printf("  %s%s%s\n", colorGray, text, colorReset)
}

func printGreen(text string) {
	fmt.Printf("  %s%s%s\n", colorGreen, text, colorReset)
}

func printRed(text string) {
	fmt.Printf("  %s%s%s\n", colorRed, text, colorReset)
}

func printDivider() {
	fmt.Printf("  %s%s%s\n", colorGray, "─────────────────────────────────────────", colorReset)
}

func printBold(text string) {
	fmt.Printf("  %s%s%s%s\n", colorBold, colorWhite, text, colorReset)
}
