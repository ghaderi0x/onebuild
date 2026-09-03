package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"onebuild/internal/config"
)

type ArtifactRecord struct {
	Target    string `json:"target"`
	Name      string `json:"name"`
	LocalPath string `json:"local_path"`
	Note      string `json:"note"`
}

type Entry struct {
	ID        string           `json:"id"`
	AppName   string           `json:"app_name"`
	RepoURL   string           `json:"repo_url"`
	RunURL    string           `json:"run_url"`
	Status    string           `json:"status"`
	Targets   []string         `json:"targets"`
	Artifacts []ArtifactRecord `json:"artifacts"`
	CreatedAt time.Time        `json:"created_at"`
}

func filePath() (string, error) {
	dir, err := config.BaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "history.json"), nil
}

func Load() ([]Entry, error) {
	fp, err := filePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func save(entries []Entry) error {
	fp, err := filePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fp, data, 0600)
}

func Add(entry Entry) error {
	entries, err := Load()
	if err != nil {
		return err
	}
	entries = append([]Entry{entry}, entries...)
	return save(entries)
}

func Update(id string, mutate func(*Entry)) error {
	entries, err := Load()
	if err != nil {
		return err
	}
	for i := range entries {
		if entries[i].ID == id {
			mutate(&entries[i])
			break
		}
	}
	return save(entries)
}

func NewID() string {
	t := time.Now()
	return fmt.Sprintf("%s-%03d", t.Format("20060102-150405"), t.Nanosecond()/1e6)
}
