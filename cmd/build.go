package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	errcatalog "onebuild/internal/errors"
	"onebuild/internal/ghapi"
	"onebuild/internal/gitcli"
	"onebuild/internal/history"
	"onebuild/internal/ui"
	"onebuild/internal/workflow"
)

var slugPattern = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)

func slugify(name string) string {
	s := strings.TrimSpace(name)
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = slugPattern.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "flutter-app"
	}
	if len(s) > 90 {
		s = s[:90]
	}
	return s
}

var repoURLPattern = regexp.MustCompile(`github\.com[:/]+([^/]+)/([^/.]+)`)

func parseGitHubRepo(input string) (string, string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("empty repository reference")
	}
	if !strings.Contains(input, "github.com") {
		parts := strings.Split(input, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return parts[0], strings.TrimSuffix(parts[1], ".git"), nil
		}
		return "", "", fmt.Errorf("could not parse %q as owner/repo", input)
	}
	m := repoURLPattern.FindStringSubmatch(input)
	if len(m) != 3 {
		return "", "", fmt.Errorf("could not parse GitHub URL: %s", input)
	}
	return m[1], strings.TrimSuffix(m[2], ".git"), nil
}

func runBuild() {
	ui.Banner()

	token, err := ensureLoggedIn()
	if err != nil {
		ui.Error("Could not log in: %s", err.Error())
		os.Exit(1)
	}
	client := ghapi.NewClient(token)

	ui.Step("Project source")
	sourceChoice := ui.AskChoice("Where is your Flutter project?", []string{
		"A local folder on this computer",
		"An existing GitHub repository (already pushed)",
	})

	var owner, repoName, branch, headSha, repoHTMLURL string
	var localPath string
	usingLocalFolder := sourceChoice == 0

	appNameDefault := "my-flutter-app"

	if usingLocalFolder {
		localPath = ui.AskString("Path to your Flutter project folder", ".")
		info, statErr := os.Stat(localPath)
		if statErr != nil || !info.IsDir() {
			ui.Error("That folder could not be found: %s", localPath)
			os.Exit(1)
		}
		if _, err := os.Stat(filepath.Join(localPath, "pubspec.yaml")); err != nil {
			ui.Warn("No pubspec.yaml found in that folder. This doesn't look like a Flutter project root.")
			if !ui.AskYesNo("Continue anyway?", false) {
				os.Exit(1)
			}
		}
		abs, _ := filepath.Abs(localPath)
		appNameDefault = slugify(filepath.Base(abs))
	} else {
		raw := ui.AskString("GitHub repository URL (or owner/repo)", "")
		o, r, parseErr := parseGitHubRepo(raw)
		if parseErr != nil {
			ui.Error("%s", parseErr.Error())
			os.Exit(1)
		}
		owner, repoName = o, r
		repoInfo, getErr := client.GetRepo(owner, repoName)
		if getErr != nil {
			ui.Error("Could not access %s/%s: %s", owner, repoName, getErr.Error())
			os.Exit(1)
		}
		branch = repoInfo.DefaultBranch
		repoHTMLURL = repoInfo.HTMLURL
		appNameDefault = repoName
	}

	appName := ui.AskString("App name (used for labels and history)", appNameDefault)

	ui.Step("Build targets")
	labels := make([]string, len(workflow.TargetOrder))
	for i, key := range workflow.TargetOrder {
		labels[i] = workflow.TargetLabels[key]
	}
	chosenIdx := ui.AskMultiChoice("Which outputs do you want to build?", labels)
	var targets []string
	for _, idx := range chosenIdx {
		targets = append(targets, workflow.TargetOrder[idx])
	}

	opts := workflow.Options{
		Branch:         "main",
		FlutterVersion: "stable",
		Targets:        targets,
	}

	wantsIOSSigned := false
	for _, t := range targets {
		if t == workflow.TargetIOSSigned {
			wantsIOSSigned = true
		}
	}
	if wantsIOSSigned {
		ui.Step("iOS signing details")
		for {
			opts.IOSTeamID = ui.AskString("Your Apple Developer Team ID", "")
			if strings.TrimSpace(opts.IOSTeamID) != "" {
				break
			}
			ui.Warn("A Team ID is required for signed iOS builds (find it at developer.apple.com/account under Membership).")
		}
		methodIdx := ui.AskChoice("Export method for the signed .ipa", []string{
			"ad-hoc (install on registered test devices)",
			"app-store (for App Store Connect upload)",
			"development",
			"enterprise",
		})
		methods := []string{"ad-hoc", "app-store", "development", "enterprise"}
		opts.IOSExportMethod = methods[methodIdx]
	}

	if usingLocalFolder {
		ui.Step("Repository setup")
		defaultRepoName := slugify(appName)
		var repoDisplayName string
		for {
			repoDisplayName = ui.AskString("Name for the new GitHub repository", defaultRepoName)
			if !client.RepoExists(mustLogin(client), repoDisplayName) {
				break
			}
			ui.Warn("A repository named %q already exists on your account.", repoDisplayName)
			defaultRepoName = repoDisplayName + "-2"
		}
		private := ui.AskYesNo("Make the repository private?", true)

		spinner := ui.NewSpinner()
		spinner.Start("Creating repository on GitHub...")
		repo, createErr := client.CreateRepo(repoDisplayName, private, "Created with OneBuild")
		spinner.Stop()
		if createErr != nil {
			ui.Error("Could not create repository: %s", createErr.Error())
			os.Exit(1)
		}
		owner = mustLogin(client)
		repoName = repoDisplayName
		branch = repo.DefaultBranch
		repoHTMLURL = repo.HTMLURL
		opts.Branch = branch

		workflowYAML := workflow.Generate(opts)
		workflowDir := filepath.Join(localPath, ".github", "workflows")
		if err := os.MkdirAll(workflowDir, 0755); err != nil {
			ui.Error("Could not create .github/workflows in your project: %s", err.Error())
			os.Exit(1)
		}
		workflowFile := filepath.Join(workflowDir, "onebuild.yml")
		if err := os.WriteFile(workflowFile, []byte(workflowYAML), 0644); err != nil {
			ui.Error("Could not write workflow file: %s", err.Error())
			os.Exit(1)
		}
		ui.Info("Added .github/workflows/onebuild.yml to your project")

		if gitcli.Available() {
			ui.Info("git found — uploading with git")
			authURL := gitcli.BuildAuthURL(repo.CloneURL, token)
			spinner.Start("Pushing your project to GitHub...")
			_, pushErr := gitcli.PushDirectory(localPath, authURL, branch, "OneBuild: initial upload")
			spinner.Stop()
			if pushErr != nil {
				ui.Warn("git upload failed, falling back to the GitHub API upload method")
				headSha, err = uploadViaAPI(client, owner, repoName, branch, localPath)
			} else {
				headSha, err = gitcli.HeadCommitSHA(localPath)
			}
		} else {
			headSha, err = uploadViaAPI(client, owner, repoName, branch, localPath)
		}
		if err != nil {
			ui.Error("Upload failed: %s", err.Error())
			os.Exit(1)
		}
	} else {
		opts.Branch = branch
		workflowYAML := workflow.Generate(opts)
		spinner := ui.NewSpinner()
		spinner.Start("Adding the CI workflow to your repository...")
		sha, upErr := client.UpsertFile(owner, repoName, branch, workflow.WorkflowPath, workflowYAML, "OneBuild: add/update CI workflow")
		spinner.Stop()
		if upErr != nil {
			ui.Error("Could not add workflow file: %s", upErr.Error())
			os.Exit(1)
		}
		headSha = sha
	}

	if wantsIOSSigned {
		ensureIOSSecrets(client, owner, repoName)
	}

	entryID := history.NewID()
	entry := history.Entry{
		ID:        entryID,
		AppName:   appName,
		RepoURL:   repoHTMLURL,
		Status:    "running",
		Targets:   targets,
		CreatedAt: time.Now(),
	}
	_ = history.Add(entry)

	ui.Step("GitHub Actions")
	spinner := ui.NewSpinner()
	spinner.Start("Waiting for the workflow to start...")
	run, findErr := client.FindRunForCommit(owner, repoName, headSha, "OneBuild", 120*time.Second)
	spinner.Stop()
	if findErr != nil {
		ui.Warn("Could not automatically detect the workflow run: %s", findErr.Error())
		ui.Info("You can check it manually here: %s/actions", repoHTMLURL)
		history.Update(entryID, func(e *history.Entry) { e.Status = "unknown" })
		return
	}

	ui.Success("Workflow started: %s", run.HTMLURL)
	history.Update(entryID, func(e *history.Entry) { e.RunURL = run.HTMLURL })

	timeoutMinutes := 20 * len(targets)
	if timeoutMinutes < 25 {
		timeoutMinutes = 25
	}
	if timeoutMinutes > 150 {
		timeoutMinutes = 150
	}

	spinner.Start("Building on GitHub Actions, this can take a while...")
	started := time.Now()
	finishedRun, waitErr := client.WaitForRun(owner, repoName, run.ID, time.Duration(timeoutMinutes)*time.Minute, func(r *ghapi.WorkflowRun) {
		elapsed := time.Since(started).Round(time.Second)
		spinner.UpdateMessage(fmt.Sprintf("Building on GitHub Actions... status: %s (%s elapsed)", r.Status, elapsed))
	})
	spinner.Stop()

	if waitErr != nil {
		ui.Warn("%s", waitErr.Error())
		ui.Info("Check progress directly: %s", run.HTMLURL)
		history.Update(entryID, func(e *history.Entry) { e.Status = "unknown" })
		return
	}

	jobs, _ := client.ListJobs(owner, repoName, finishedRun.ID)
	ui.Step("Results")
	successCount, failCount := 0, 0
	for _, j := range jobs {
		icon := "•"
		switch j.Conclusion {
		case "success":
			icon = "✔"
			successCount++
		case "failure":
			icon = "✖"
			failCount++
		case "cancelled":
			icon = "⊘"
		case "skipped":
			icon = "–"
		}
		fmt.Printf("  %s %s  (%s)\n", icon, j.Name, j.HTMLURL)
	}

	overallStatus := "partial"
	if failCount == 0 && successCount > 0 {
		overallStatus = "success"
	} else if successCount == 0 {
		overallStatus = "failed"
	}

	outDir := filepath.Join(mustOutputDir(), fmt.Sprintf("%s-%s", slugify(appName), entryID))
	os.MkdirAll(outDir, 0755)

	var artifactRecords []history.ArtifactRecord
	if successCount > 0 {
		ui.Step("Downloading artifacts")
		artifacts, _ := client.ListArtifacts(owner, repoName, finishedRun.ID)
		for _, art := range artifacts {
			spinner.Start(fmt.Sprintf("Downloading %s...", art.Name))
			data, dErr := client.DownloadArtifactZip(art)
			spinner.Stop()
			if dErr != nil {
				ui.Warn("Could not download %s: %s", art.Name, dErr.Error())
				continue
			}
			targetDir := filepath.Join(outDir, art.Name)
			os.MkdirAll(targetDir, 0755)
			names, exErr := ghapi.ExtractZip(data, func(name string, content []byte) error {
				dest := filepath.Join(targetDir, name)
				if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
					return err
				}
				return os.WriteFile(dest, content, 0644)
			})
			if exErr != nil {
				ui.Warn("Could not extract %s: %s", art.Name, exErr.Error())
				continue
			}
			ui.Success("%s saved to %s", art.Name, targetDir)
			note := ""
			if len(names) > 0 {
				note = names[0]
			}
			artifactRecords = append(artifactRecords, history.ArtifactRecord{
				Target:    art.Name,
				Name:      art.Name,
				LocalPath: targetDir,
				Note:      note,
			})
		}
	}

	type jobDiagnosis struct {
		JobName     string
		JobURL      string
		Annotations []ghapi.Annotation
		LogTail     []string
	}
	var diagnoses []jobDiagnosis

	if failCount > 0 {
		ui.Step("Failure details")

		annotationsByJob := map[string][]ghapi.Annotation{}
		checkRuns, checksErr := client.ListCheckRunsForRef(owner, repoName, headSha)
		if checksErr == nil {
			for _, cr := range checkRuns {
				if cr.Conclusion != "failure" || cr.AnnotationsCount == 0 {
					continue
				}
				anns, annErr := client.ListCheckRunAnnotations(owner, repoName, cr.ID)
				if annErr == nil && len(anns) > 0 {
					annotationsByJob[cr.Name] = anns
				}
			}
		}

		for _, j := range jobs {
			if j.Conclusion != "failure" {
				continue
			}
			spinner.Start(fmt.Sprintf("Fetching logs for %s...", j.Name))
			jobLog, logErr := client.DownloadJobLogsText(owner, repoName, j.ID)
			spinner.Stop()

			var tail []string
			if logErr == nil {
				tail = truncateLines(jobLog, 150)
			}

			diag := jobDiagnosis{
				JobName:     j.Name,
				JobURL:      j.HTMLURL,
				Annotations: annotationsByJob[j.Name],
				LogTail:     tail,
			}
			diagnoses = append(diagnoses, diag)

			fmt.Println()
			ui.Warn("Job failed: %s", j.Name)
			fmt.Printf("     %s\n", j.HTMLURL)

			if len(diag.Annotations) > 0 {
				fmt.Println("     GitHub annotations:")
				for _, a := range diag.Annotations {
					fmt.Printf("       [%s] %s:%d  %s\n", a.AnnotationLevel, a.Path, a.StartLine, a.Message)
				}
			}

			if logErr != nil {
				ui.Warn("Could not fetch the log for this job: %s", logErr.Error())
			} else {
				fmt.Println("     Last lines of the log:")
				start := len(tail) - 25
				if start < 0 {
					start = 0
				}
				for _, l := range tail[start:] {
					fmt.Printf("       %s\n", l)
				}
			}
		}

		if ui.AskYesNo("Save a PDF of these failures?", false) {
			sections := []errcatalog.Section{
				{Heading: "Failed jobs", Lines: failedJobLines(jobs)},
			}
			for _, diag := range diagnoses {
				var lines []string
				lines = append(lines, diag.JobURL, "")
				if len(diag.Annotations) > 0 {
					lines = append(lines, "GitHub annotations:")
					for _, a := range diag.Annotations {
						lines = append(lines, fmt.Sprintf("- [%s] %s:%d %s", a.AnnotationLevel, a.Path, a.StartLine, a.Message))
					}
					lines = append(lines, "")
				}
				lines = append(lines, "Log tail:")
				lines = append(lines, diag.LogTail...)
				sections = append(sections, errcatalog.Section{Heading: "Job: " + diag.JobName, Lines: lines})
			}
			pdfBytes := errcatalog.GeneratePDF("OneBuild failure report - "+appName, sections)
			pdfPath := filepath.Join(desktopDir(), fmt.Sprintf("onebuild-error-report-%s.pdf", entryID))
			if err := os.WriteFile(pdfPath, pdfBytes, 0644); err != nil {
				ui.Error("Could not write PDF: %s", err.Error())
			} else {
				ui.Success("PDF saved to %s", pdfPath)
			}
		}
	}

	history.Update(entryID, func(e *history.Entry) {
		e.Status = overallStatus
		e.Artifacts = artifactRecords
	})

	fmt.Println()
	ui.Success("Done. Repository: %s", repoHTMLURL)
	if len(artifactRecords) > 0 {
		ui.Success("Output folder: %s", outDir)
	}
}

func mustLogin(client *ghapi.Client) string {
	user, err := client.CurrentUser()
	if err != nil {
		ui.Error("Could not confirm your GitHub username: %s", err.Error())
		os.Exit(1)
	}
	return user.Login
}

func mustOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	dir := filepath.Join(home, "OneBuild-output")
	os.MkdirAll(dir, 0755)
	return dir
}

func desktopDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return mustOutputDir()
	}
	dir := filepath.Join(home, "Desktop")
	if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
		return mustOutputDir()
	}
	return dir
}

func uploadViaAPI(client *ghapi.Client, owner, repo, branch, localPath string) (string, error) {
	spinner := ui.NewSpinner()
	spinner.Start("Uploading files via the GitHub API...")
	sha, err := client.UploadDirectoryAPI(owner, repo, branch, localPath, "OneBuild: initial upload", func(stage string, done, total int) {
		if strings.HasPrefix(stage, "skip:") {
			return
		}
		if total > 0 {
			spinner.UpdateMessage(fmt.Sprintf("Uploading files via the GitHub API... (%d/%d)", done, total))
		}
	})
	spinner.Stop()
	return sha, err
}

func ensureIOSSecrets(client *ghapi.Client, owner, repo string) {
	ui.Step("iOS signing secrets")
	for {
		missing, err := client.HasAllSecrets(owner, repo, workflow.IOSSignedRequiredSecrets)
		if err != nil {
			ui.Warn("Could not check repository secrets: %s", err.Error())
			return
		}
		if len(missing) == 0 {
			ui.Success("All required secrets are present")
			return
		}
		ui.Warn("This repository is missing %d required secret(s) for signed iOS builds:", len(missing))
		for _, m := range missing {
			fmt.Printf("     - %s\n", m)
		}
		fmt.Println()
		fmt.Println("  Add them at:")
		fmt.Printf("  https://github.com/%s/%s/settings/secrets/actions/new\n", owner, repo)
		fmt.Println()
		fmt.Println("  IOS_CERTIFICATE_BASE64            base64 of your exported .p12 certificate")
		fmt.Println("  IOS_CERTIFICATE_PASSWORD          the password you set when exporting the .p12")
		fmt.Println("  IOS_PROVISIONING_PROFILE_BASE64   base64 of your .mobileprovision file")
		fmt.Println("  KEYCHAIN_PASSWORD                 any password, used only for a temporary CI keychain")
		fmt.Println()
		fmt.Println("  Tip: onebuild ios-cert csr / package can generate and package your")
		fmt.Println("  certificate without a Mac. onebuild ios-cert encode can base64-encode")
		fmt.Println("  your provisioning profile. Run 'onebuild help' for details.")
		fmt.Println()
		fmt.Println("  Or manually: on macOS,  base64 -i certificate.p12 | pbcopy")
		fmt.Println("               on Linux,  base64 -w0 certificate.p12")
		fmt.Println()

		choice := ui.AskChoice("What would you like to do?", []string{
			"I've added them, check again",
			"Skip the signed iOS build and continue",
			"Cancel the whole build",
		})
		if choice == 0 {
			continue
		} else if choice == 1 {
			return
		}
		os.Exit(1)
	}
}

func failedJobLines(jobs []ghapi.Job) []string {
	var lines []string
	for _, j := range jobs {
		if j.Conclusion == "failure" {
			lines = append(lines, fmt.Sprintf("- %s (%s)", j.Name, j.HTMLURL))
		}
	}
	if len(lines) == 0 {
		lines = append(lines, "(none)")
	}
	return lines
}

func truncateLines(text string, maxLines int) []string {
	lines := strings.Split(text, "\n")
	if len(lines) <= maxLines {
		return lines
	}
	return lines[len(lines)-maxLines:]
}
