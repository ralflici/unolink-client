package main

import (
	"fmt"

	"unolink-client/cmd"

	"github.com/charmbracelet/lipgloss"
	// "github.com/charmbracelet/lipgloss/table"
)

type Log struct {
    logType string
    logStyle lipgloss.Style
}

func (l Log) Log(msg string) {
    fmt.Println(l.logStyle.Render(l.logType + msg))
}

var (
    Info    Log
    Debug   Log
    Warn    Log
    Error   Log
)

func initLogs() {
    Info.logType = "INFO: "
    Info.logStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

    Debug.logType = "DEBUG: "
    Debug.logStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

    Warn.logType = "WARN: "
    Warn.logStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

    Error.logType = "ERROR: "
    Error.logStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
}

func main() {
    initLogs()
    cmd.Execute()
}
