package selfupdate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const releasesURL = "https://api.github.com/repos/ghaderi0x/onebuild/releases/latest"

type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

func fetchLatestRelease() (*Release, error) {
	client := &http.Client{Timeout: 4 * time.Second}
	req, err := http.NewRequest(http.MethodGet, releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, err
	}
	return &rel, nil
}

func assetNameForPlatform() (string, bool) {
	switch runtime.GOOS + "/" + runtime.GOARCH {
	case "darwin/amd64":
		return "onebuild-macos-intel", true
	case "darwin/arm64":
		return "onebuild-macos-arm64", true
	case "linux/amd64":
		return "onebuild-linux-amd64", true
	case "linux/arm64":
		return "onebuild-linux-arm64", true
	case "windows/amd64":
		return "onebuild-windows-amd64.exe", true
	default:
		return "", false
	}
}

func parseVersion(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	nums := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

func isNewer(latest, current string) bool {
	lv, cv := parseVersion(latest), parseVersion(current)
	if lv == nil || cv == nil {
		return strings.TrimPrefix(latest, "v") != current
	}
	for i := 0; i < len(lv) || i < len(cv); i++ {
		var l, c int
		if i < len(lv) {
			l = lv[i]
		}
		if i < len(cv) {
			c = cv[i]
		}
		if l != c {
			return l > c
		}
	}
	return false
}

func CheckForUpdate(currentVersion string) (latestVersion string, hasUpdate bool) {
	rel, err := fetchLatestRelease()
	if err != nil || rel == nil || rel.TagName == "" {
		return "", false
	}
	if isNewer(rel.TagName, currentVersion) {
		return rel.TagName, true
	}
	return "", false
}

func downloadAsset(url string) ([]byte, error) {
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func PerformUpdate(currentVersion string) (string, error) {
	rel, err := fetchLatestRelease()
	if err != nil {
		return "", fmt.Errorf("could not check the latest release: %w", err)
	}
	if !isNewer(rel.TagName, currentVersion) {
		return rel.TagName, nil
	}

	assetName, ok := assetNameForPlatform()
	if !ok {
		return "", fmt.Errorf("no prebuilt binary for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	var downloadURL string
	for _, a := range rel.Assets {
		if a.Name == assetName {
			downloadURL = a.BrowserDownloadURL
			break
		}
	}
	if downloadURL == "" {
		return "", fmt.Errorf("release %s has no asset named %s", rel.TagName, assetName)
	}

	data, err := downloadAsset(downloadURL)
	if err != nil {
		return "", fmt.Errorf("downloading new version: %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine the running executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", fmt.Errorf("could not resolve executable path: %w", err)
	}

	dir := filepath.Dir(exePath)
	tmpPath := filepath.Join(dir, ".onebuild-update-tmp")
	oldPath := filepath.Join(dir, ".onebuild-old")

	if err := os.WriteFile(tmpPath, data, 0755); err != nil {
		return "", fmt.Errorf("writing new binary: %w", err)
	}

	os.Remove(oldPath)
	if err := os.Rename(exePath, oldPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("could not replace the running binary (try running with elevated permissions): %w", err)
	}
	if err := os.Rename(tmpPath, exePath); err != nil {
		os.Rename(oldPath, exePath)
		return "", fmt.Errorf("could not move the new binary into place: %w", err)
	}
	os.Remove(oldPath)

	return rel.TagName, nil
}
