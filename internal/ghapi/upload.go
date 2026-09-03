package ghapi

import (
	"encoding/base64"
	"fmt"
	"os"

	"onebuild/internal/fsutil"
)

type blobRequest struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

type blobResponse struct {
	Sha string `json:"sha"`
}

type refResponse struct {
	Object struct {
		Sha string `json:"sha"`
	} `json:"object"`
}

type commitResponse struct {
	Sha  string `json:"sha"`
	Tree struct {
		Sha string `json:"sha"`
	} `json:"tree"`
}

type treeEntry struct {
	Path string `json:"path"`
	Mode string `json:"mode"`
	Type string `json:"type"`
	Sha  string `json:"sha"`
}

type createTreeRequest struct {
	BaseTree string      `json:"base_tree,omitempty"`
	Tree     []treeEntry `json:"tree"`
}

type treeResponse struct {
	Sha string `json:"sha"`
}

type createCommitRequest struct {
	Message string   `json:"message"`
	Tree    string   `json:"tree"`
	Parents []string `json:"parents"`
}

type createCommitResponse struct {
	Sha string `json:"sha"`
}

type updateRefRequest struct {
	Sha   string `json:"sha"`
	Force bool   `json:"force"`
}

type UploadProgress func(stage string, done, total int)

func (c *Client) UploadDirectoryAPI(owner, repo, branch, localDir, commitMessage string, progress UploadProgress) (string, error) {
	entries, skipped, err := fsutil.CollectFiles(localDir)
	if err != nil {
		return "", err
	}
	for _, s := range skipped {
		if progress != nil {
			progress("skip:"+s, 0, 0)
		}
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("no files found to upload in %s", localDir)
	}

	total := len(entries)
	treeEntries := make([]treeEntry, 0, total)

	for i, entry := range entries {
		content, err := os.ReadFile(entry.AbsPath)
		if err != nil {
			return "", fmt.Errorf("reading %s: %w", entry.RelPath, err)
		}
		encoded := base64.StdEncoding.EncodeToString(content)

		var blob blobResponse
		err = c.Post(fmt.Sprintf("/repos/%s/%s/git/blobs", owner, repo), blobRequest{
			Content:  encoded,
			Encoding: "base64",
		}, &blob)
		if err != nil {
			return "", fmt.Errorf("uploading blob for %s: %w", entry.RelPath, err)
		}

		treeEntries = append(treeEntries, treeEntry{
			Path: entry.RelPath,
			Mode: "100644",
			Type: "blob",
			Sha:  blob.Sha,
		})

		if progress != nil {
			progress("upload", i+1, total)
		}
	}

	var ref refResponse
	if err := c.Get(fmt.Sprintf("/repos/%s/%s/git/ref/heads/%s", owner, repo, branch), &ref); err != nil {
		return "", fmt.Errorf("reading base branch: %w", err)
	}
	baseCommitSha := ref.Object.Sha

	var baseCommit commitResponse
	if err := c.Get(fmt.Sprintf("/repos/%s/%s/git/commits/%s", owner, repo, baseCommitSha), &baseCommit); err != nil {
		return "", fmt.Errorf("reading base commit: %w", err)
	}

	var newTree treeResponse
	if err := c.Post(fmt.Sprintf("/repos/%s/%s/git/trees", owner, repo), createTreeRequest{
		BaseTree: baseCommit.Tree.Sha,
		Tree:     treeEntries,
	}, &newTree); err != nil {
		return "", fmt.Errorf("creating tree: %w", err)
	}

	var newCommit createCommitResponse
	if err := c.Post(fmt.Sprintf("/repos/%s/%s/git/commits", owner, repo), createCommitRequest{
		Message: commitMessage,
		Tree:    newTree.Sha,
		Parents: []string{baseCommitSha},
	}, &newCommit); err != nil {
		return "", fmt.Errorf("creating commit: %w", err)
	}

	if err := c.Patch(fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", owner, repo, branch), updateRefRequest{
		Sha:   newCommit.Sha,
		Force: false,
	}, nil); err != nil {
		return "", fmt.Errorf("updating branch ref: %w", err)
	}

	return newCommit.Sha, nil
}

func (c *Client) UpsertFile(owner, repo, branch, path, content, message string) (string, error) {
	type getContentResponse struct {
		Sha string `json:"sha"`
	}
	var existing getContentResponse
	getErr := c.Get(fmt.Sprintf("/repos/%s/%s/contents/%s?ref=%s", owner, repo, path, branch), &existing)

	type putContentRequest struct {
		Message string `json:"message"`
		Content string `json:"content"`
		Branch  string `json:"branch"`
		Sha     string `json:"sha,omitempty"`
	}
	type putContentResponse struct {
		Commit struct {
			Sha string `json:"sha"`
		} `json:"commit"`
	}
	req := putContentRequest{
		Message: message,
		Content: base64.StdEncoding.EncodeToString([]byte(content)),
		Branch:  branch,
	}
	if getErr == nil {
		req.Sha = existing.Sha
	}
	var resp putContentResponse
	if err := c.Put(fmt.Sprintf("/repos/%s/%s/contents/%s", owner, repo, path), req, &resp); err != nil {
		return "", err
	}
	return resp.Commit.Sha, nil
}
