package ghapi

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type WorkflowRun struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HeadSha    string `json:"head_sha"`
	HTMLURL    string `json:"html_url"`
	CreatedAt  string `json:"created_at"`
}

type listRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

func (c *Client) FindRunForCommit(owner, repo, headSha, workflowName string, timeout time.Duration) (*WorkflowRun, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		var resp listRunsResponse
		err := c.Get(fmt.Sprintf("/repos/%s/%s/actions/runs?per_page=30", owner, repo), &resp)
		if err == nil {
			for _, run := range resp.WorkflowRuns {
				if run.HeadSha == headSha && (workflowName == "" || run.Name == workflowName) {
					return &run, nil
				}
			}
		}
		time.Sleep(3 * time.Second)
	}
	return nil, fmt.Errorf("timed out waiting for a workflow run to start for commit %s", headSha)
}

func (c *Client) GetRun(owner, repo string, runID int64) (*WorkflowRun, error) {
	var run WorkflowRun
	if err := c.Get(fmt.Sprintf("/repos/%s/%s/actions/runs/%d", owner, repo, runID), &run); err != nil {
		return nil, err
	}
	return &run, nil
}

func (c *Client) WaitForRun(owner, repo string, runID int64, timeout time.Duration, onTick func(*WorkflowRun)) (*WorkflowRun, error) {
	deadline := time.Now().Add(timeout)
	interval := 6 * time.Second
	for time.Now().Before(deadline) {
		run, err := c.GetRun(owner, repo, runID)
		if err == nil {
			if onTick != nil {
				onTick(run)
			}
			if run.Status == "completed" {
				return run, nil
			}
		}
		time.Sleep(interval)
	}
	return nil, fmt.Errorf("timed out waiting for workflow run to finish")
}

type Job struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	HTMLURL    string `json:"html_url"`
	Steps      []Step `json:"steps"`
}

type Step struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Number     int    `json:"number"`
}

type listJobsResponse struct {
	Jobs []Job `json:"jobs"`
}

func (c *Client) ListJobs(owner, repo string, runID int64) ([]Job, error) {
	var resp listJobsResponse
	if err := c.Get(fmt.Sprintf("/repos/%s/%s/actions/runs/%d/jobs?per_page=100", owner, repo, runID), &resp); err != nil {
		return nil, err
	}
	return resp.Jobs, nil
}

type Artifact struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	SizeInBytes        int64  `json:"size_in_bytes"`
	Expired            bool   `json:"expired"`
	ArchiveDownloadURL string `json:"archive_download_url"`
}

type listArtifactsResponse struct {
	Artifacts []Artifact `json:"artifacts"`
}

func (c *Client) ListArtifacts(owner, repo string, runID int64) ([]Artifact, error) {
	var resp listArtifactsResponse
	if err := c.Get(fmt.Sprintf("/repos/%s/%s/actions/runs/%d/artifacts?per_page=100", owner, repo, runID), &resp); err != nil {
		return nil, err
	}
	return resp.Artifacts, nil
}

func (c *Client) downloadRaw(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) DownloadArtifactZip(artifact Artifact) ([]byte, error) {
	return c.downloadRaw(artifact.ArchiveDownloadURL)
}

func (c *Client) DownloadJobLogsText(owner, repo string, jobID int64) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/actions/jobs/%d/logs", apiBase, owner, repo, jobID)
	data, err := c.downloadRaw(url)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) DownloadRunLogsText(owner, repo string, runID int64) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/actions/runs/%d/logs", apiBase, owner, repo, runID)
	data, err := c.downloadRaw(url)
	if err != nil {
		return "", err
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			continue
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			continue
		}
		buf.WriteString("\n--- ")
		buf.WriteString(f.Name)
		buf.WriteString(" ---\n")
		buf.Write(content)
	}
	return buf.String(), nil
}

func ExtractZip(data []byte, writeFile func(name string, content []byte) error) ([]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		cleanName := filepath.ToSlash(filepath.Clean(f.Name))
		if cleanName == "." || strings.HasPrefix(cleanName, "../") || strings.Contains(cleanName, "/../") || filepath.IsAbs(cleanName) {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		if err := writeFile(cleanName, content); err != nil {
			return nil, err
		}
		names = append(names, cleanName)
	}
	return names, nil
}
