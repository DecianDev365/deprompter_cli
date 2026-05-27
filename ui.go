package main

import "github.com/charmbracelet/lipgloss"

var (
	orange    = lipgloss.Color("#FF8800")
	green     = lipgloss.Color("#00FF88")
	red       = lipgloss.Color("#FF4444")
	gray      = lipgloss.Color("#888888")
	darkGray  = lipgloss.Color("#444444")

	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(orange)

	successStyle = lipgloss.NewStyle().Bold(true).Foreground(green)

	errorStyle = lipgloss.NewStyle().Bold(true).Foreground(red)

	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))

	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(orange)

	itemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#DDDDDD"))

	mutedStyle = lipgloss.NewStyle().Foreground(gray)

	helpStyle = lipgloss.NewStyle().Foreground(darkGray)

	spinnerStyle = lipgloss.NewStyle().Foreground(orange)
)
