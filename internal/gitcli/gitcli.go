package gitcli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var defaultExcludes = []string{
	".dart_tool/",
	"build/",
	".gradle/",
	"**/Pods/",
	"**/.symlinks/",
	"**/DerivedData/",
	"node_modules/",
	".fvm/",
	".flutter-plugins",
	".flutter-plugins-dependencies",
	".packages",
	".DS_Store",
	"*.iml",
}

func Available() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func BuildAuthURL(cloneURL, token string) string {
	const prefix = "https://"
	if strings.HasPrefix(cloneURL, prefix) {
		return prefix + token + "@" + strings.TrimPrefix(cloneURL, prefix)
	}
	return cloneURL
}

func isRepo(dir string) bool {
	_, err := run(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func writeLocalExcludes(dir string) {
	gitDir := filepath.Join(dir, ".git", "info")
	os.MkdirAll(gitDir, 0755)
	excludePath := filepath.Join(gitDir, "exclude")
	content := strings.Join(defaultExcludes, "\n") + "\n"
	existing, _ := os.ReadFile(excludePath)
	combined := string(existing)
	if !strings.Contains(combined, "# onebuild-defaults") {
		combined += "\n# onebuild-defaults\n" + content
		os.WriteFile(excludePath, []byte(combined), 0644)
	}
}

func ensureIdentity(dir string) {
	if out, err := run(dir, "config", "user.email"); err != nil || strings.TrimSpace(out) == "" {
		run(dir, "config", "user.email", "onebuild@local")
	}
	if out, err := run(dir, "config", "user.name"); err != nil || strings.TrimSpace(out) == "" {
		run(dir, "config", "user.name", "OneBuild")
	}
}

func HeadCommitSHA(dir string) (string, error) {
	out, err := run(dir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func PushDirectory(dir, authURL, branch, commitMessage string) (string, error) {
	var log strings.Builder

	if !isRepo(dir) {
		out, err := run(dir, "init")
		log.WriteString(out)
		if err != nil {
			return log.String(), err
		}
	}

	writeLocalExcludes(dir)
	ensureIdentity(dir)

	out, err := run(dir, "add", "-A")
	log.WriteString(out)
	if err != nil {
		return log.String(), err
	}

	statusOut, _ := run(dir, "status", "--porcelain")
	if strings.TrimSpace(statusOut) != "" {
		out, err = run(dir, "commit", "-m", commitMessage)
		log.WriteString(out)
		if err != nil {
			return log.String(), err
		}
	}

	out, err = run(dir, "branch", "-M", branch)
	log.WriteString(out)
	if err != nil {
		return log.String(), err
	}

	if _, err := run(dir, "remote", "get-url", "origin"); err == nil {
		out, err = run(dir, "remote", "set-url", "origin", authURL)
	} else {
		out, err = run(dir, "remote", "add", "origin", authURL)
	}
	log.WriteString(out)
	if err != nil {
		return log.String(), err
	}
	defer run(dir, "remote", "remove", "origin")

	out, err = run(dir, "push", "-u", "origin", branch, "--force")
	log.WriteString(out)
	if err != nil {
		return log.String(), err
	}

	return log.String(), nil
}
