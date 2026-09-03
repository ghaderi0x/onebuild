package cmd

import (
	"fmt"
	"os"

	"onebuild/internal/config"
	"onebuild/internal/ghapi"
	"onebuild/internal/ui"
)

func runAuth(args []string) {
	if len(args) == 0 {
		ui.Warn("Specify a subcommand: login, logout, or status")
		return
	}
	switch args[0] {
	case "login":
		if err := loginFlow(); err != nil {
			ui.Error("Login failed: %s", err.Error())
			os.Exit(1)
		}
	case "logout":
		if err := config.ClearSession(); err != nil {
			ui.Error("Could not clear saved session: %s", err.Error())
			os.Exit(1)
		}
		ui.Success("Logged out. The saved token was removed from this machine.")
	case "status":
		printAuthStatus()
	default:
		ui.Warn("Unknown auth subcommand: %s", args[0])
	}
}

func printAuthStatus() {
	if !config.HasSession() {
		ui.Info("Not logged in. Run: onebuild auth login")
		return
	}
	sess, err := config.LoadSession()
	if err != nil {
		ui.Error("Could not read saved session: %s", err.Error())
		return
	}
	ui.Success("Logged in as %s (@%s)", sess.Name, sess.Login)
}

func loginFlow() error {
	ui.Step("Connect your GitHub account")
	fmt.Println("  OneBuild needs a GitHub Personal Access Token to create repositories")
	fmt.Println("  and run GitHub Actions on your behalf. It is stored encrypted on")
	fmt.Println("  this machine only, and never leaves your computer except to talk")
	fmt.Println("  directly to api.github.com.")
	fmt.Println()
	fmt.Println("  Create one at: https://github.com/settings/tokens/new")
	fmt.Println("  Required scopes (classic token): repo, workflow")
	fmt.Println()

	token := ui.AskSecret("Paste your GitHub token")
	if token == "" {
		return fmt.Errorf("no token provided")
	}

	spinner := ui.NewSpinner()
	spinner.Start("Verifying token with GitHub...")
	client := ghapi.NewClient(token)
	user, err := client.CurrentUser()
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("token could not be verified: %w", err)
	}

	if err := config.SaveSession(user.Login, user.Name, token); err != nil {
		return fmt.Errorf("could not save session: %w", err)
	}

	ui.Success("Logged in as %s (@%s). You won't need to do this again.", user.Name, user.Login)
	return nil
}

func ensureLoggedIn() (string, error) {
	if config.HasSession() {
		token, err := config.LoadToken()
		if err == nil {
			return token, nil
		}
		ui.Warn("Saved session could not be used (%s). Please log in again.", err.Error())
	}
	if err := loginFlow(); err != nil {
		return "", err
	}
	return config.LoadToken()
}
