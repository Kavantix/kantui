package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Kavantix/kantui/internal/model"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	program := tea.NewProgram(model.New())

	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	slog.Info("Starting")

	_, err = program.Run()
	if err != nil {
		slog.Error("Running program failed: ", slog.String("error", err.Error()))
	}
}
