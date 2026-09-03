package cmd

import (
	"fmt"
	"os"

	"onebuild/internal/selfupdate"
	"onebuild/internal/ui"
)

const Version = "1.0.0"

func Execute() {
	args := os.Args[1:]
	if len(args) == 0 {
		ui.ClearScreen()
		ui.Banner()
		printHelp()
		return
	}

	switch args[0] {
	case "build":
		ui.ClearScreen()
		checkForUpdateNotice()
		runBuild()
	case "history":
		ui.ClearScreen()
		runHistory(args[1:])
	case "auth":
		ui.ClearScreen()
		runAuth(args[1:])
	case "logout":
		ui.ClearScreen()
		runAuth([]string{"logout"})
	case "ios-cert":
		ui.ClearScreen()
		runIOSCert(args[1:])
	case "doctor":
		ui.ClearScreen()
		runDoctor()
	case "update":
		ui.ClearScreen()
		runUpdate()
	case "version", "-v", "--version":
		fmt.Println("OneBuild v" + Version)
	case "help", "-h", "--help":
		ui.ClearScreen()
		ui.Banner()
		printHelp()
	default:
		ui.Error("Unknown command: %s", args[0])
		printHelp()
		os.Exit(1)
	}
}

func checkForUpdateNotice() {
	latest, hasUpdate := selfupdate.CheckForUpdate(Version)
	if hasUpdate {
		ui.Warn("A newer version (%s) is available. Run 'onebuild update' to update.", latest)
		fmt.Println()
	}
}

func runUpdate() {
	ui.Banner()
	ui.Step("Checking for updates")

	spinner := ui.NewSpinner()
	spinner.Start("Checking the latest release...")
	newVersion, err := selfupdate.PerformUpdate(Version)
	spinner.Stop()

	if err != nil {
		ui.Error("Update failed: %s", err.Error())
		ui.Info("You can always download the latest release yourself from:")
		ui.Info("https://github.com/ghaderi0x/onebuild/releases/latest")
		os.Exit(1)
	}

	if newVersion == "" || "v"+Version == newVersion || Version == newVersion {
		ui.Success("You're already on the latest version (v%s).", Version)
		return
	}

	ui.Success("Updated to %s. Run 'onebuild version' to confirm.", newVersion)
}

func printHelp() {
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  onebuild build            Start the interactive build wizard")
	fmt.Println("  onebuild history          Show past builds")
	fmt.Println("  onebuild auth login       Save a GitHub token for future runs")
	fmt.Println("  onebuild auth logout      Remove the saved GitHub token")
	fmt.Println("  onebuild logout           Shortcut for 'onebuild auth logout'")
	fmt.Println("  onebuild auth status      Show who is currently logged in")
	fmt.Println("  onebuild ios-cert csr     Generate an Apple certificate request (no Mac needed)")
	fmt.Println("  onebuild ios-cert package Package a downloaded certificate into a .p12")
	fmt.Println("  onebuild ios-cert encode  Base64-encode a file (e.g. a provisioning profile)")
	fmt.Println("  onebuild doctor           Check your local environment")
	fmt.Println("  onebuild update           Update OneBuild to the latest version")
	fmt.Println("  onebuild version          Print the version number")
	fmt.Println()
	fmt.Println("Repo & issues: https://github.com/ghaderi0x/onebuild")
	fmt.Println()
}
