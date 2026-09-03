package cmd

import (
	"net/http"
	"runtime"
	"time"

	"onebuild/internal/config"
	"onebuild/internal/gitcli"
	"onebuild/internal/ui"
)

func runDoctor() {
	ui.Banner()
	ui.Step("Environment check")

	ui.Info("OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)

	if gitcli.Available() {
		ui.Success("git is installed (will be used for faster uploads)")
	} else {
		ui.Warn("git was not found (OneBuild will upload via the GitHub API instead, which works fine but is a bit slower for large projects)")
	}

	if _, err := config.BaseDir(); err == nil {
		ui.Success("Local config directory is writable (~/.onebuild)")
	} else {
		ui.Error("Could not create ~/.onebuild: %s", err.Error())
	}

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get("https://api.github.com")
	if err != nil {
		ui.Error("Could not reach api.github.com: %s", err.Error())
	} else {
		resp.Body.Close()
		ui.Success("api.github.com is reachable")
	}

	if config.HasSession() {
		ui.Success("A GitHub session is saved")
	} else {
		ui.Warn("Not logged in yet. Run: onebuild auth login")
	}
}
