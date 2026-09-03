package ghapi

import "fmt"

type Repo struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
	Private       bool   `json:"private"`
}

type createRepoRequest struct {
	Name        string `json:"name"`
	Private     bool   `json:"private"`
	AutoInit    bool   `json:"auto_init"`
	Description string `json:"description"`
}

func (c *Client) CreateRepo(name string, private bool, description string) (*Repo, error) {
	var repo Repo
	req := createRepoRequest{Name: name, Private: private, AutoInit: true, Description: description}
	if err := c.Post("/user/repos", req, &repo); err != nil {
		return nil, err
	}
	return &repo, nil
}

func (c *Client) GetRepo(owner, name string) (*Repo, error) {
	var repo Repo
	if err := c.Get(fmt.Sprintf("/repos/%s/%s", owner, name), &repo); err != nil {
		return nil, err
	}
	return &repo, nil
}

func (c *Client) RepoExists(owner, name string) bool {
	_, err := c.GetRepo(owner, name)
	return err == nil
}

type secretsListResponse struct {
	TotalCount int      `json:"total_count"`
	Secrets    []Secret `json:"secrets"`
}

type Secret struct {
	Name string `json:"name"`
}

func (c *Client) ListSecretNames(owner, repo string) ([]string, error) {
	var resp secretsListResponse
	if err := c.Get(fmt.Sprintf("/repos/%s/%s/actions/secrets", owner, repo), &resp); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(resp.Secrets))
	for _, s := range resp.Secrets {
		names = append(names, s.Name)
	}
	return names, nil
}

func (c *Client) HasAllSecrets(owner, repo string, required []string) ([]string, error) {
	existing, err := c.ListSecretNames(owner, repo)
	if err != nil {
		return nil, err
	}
	have := map[string]bool{}
	for _, n := range existing {
		have[n] = true
	}
	var missing []string
	for _, n := range required {
		if !have[n] {
			missing = append(missing, n)
		}
	}
	return missing, nil
}
