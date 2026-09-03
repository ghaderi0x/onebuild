package cmd

import (
	"fmt"

	"onebuild/internal/history"
	"onebuild/internal/ui"
)

func runHistory(args []string) {
	entries, err := history.Load()
	if err != nil {
		ui.Error("Could not read history: %s", err.Error())
		return
	}
	if len(entries) == 0 {
		ui.Info("No builds yet. Run: onebuild build")
		return
	}

	ui.Banner()
	ui.Step("Build history")

	for i, e := range entries {
		statusIcon := "•"
		switch e.Status {
		case "success":
			statusIcon = "✔"
		case "failed":
			statusIcon = "✖"
		case "partial":
			statusIcon = "◐"
		case "running":
			statusIcon = "…"
		}
		fmt.Printf("\n  %d. [%s] %s\n", i+1, statusIcon, e.AppName)
		fmt.Printf("     Repo:   %s\n", e.RepoURL)
		if e.RunURL != "" {
			fmt.Printf("     Run:    %s\n", e.RunURL)
		}
		fmt.Printf("     Date:   %s\n", e.CreatedAt.Format("2006-01-02 15:04"))
		if len(e.Artifacts) > 0 {
			fmt.Println("     Artifacts:")
			for _, a := range e.Artifacts {
				loc := a.LocalPath
				if loc == "" {
					loc = "(not downloaded)"
				}
				fmt.Printf("       - %s: %s\n", a.Target, loc)
			}
		}
	}
	fmt.Println()
}
