package ghapi

import "fmt"

type CheckRun struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Status          string `json:"status"`
	Conclusion      string `json:"conclusion"`
	HTMLURL         string `json:"html_url"`
	AnnotationsCount int   `json:"-"`
	Output          struct {
		AnnotationsCount int `json:"annotations_count"`
	} `json:"output"`
}

type listCheckRunsResponse struct {
	TotalCount int        `json:"total_count"`
	CheckRuns  []CheckRun `json:"check_runs"`
}

func (c *Client) ListCheckRunsForRef(owner, repo, ref string) ([]CheckRun, error) {
	var resp listCheckRunsResponse
	if err := c.Get(fmt.Sprintf("/repos/%s/%s/commits/%s/check-runs?per_page=100", owner, repo, ref), &resp); err != nil {
		return nil, err
	}
	for i := range resp.CheckRuns {
		resp.CheckRuns[i].AnnotationsCount = resp.CheckRuns[i].Output.AnnotationsCount
	}
	return resp.CheckRuns, nil
}

type Annotation struct {
	Path            string `json:"path"`
	StartLine       int    `json:"start_line"`
	EndLine         int    `json:"end_line"`
	AnnotationLevel string `json:"annotation_level"`
	Title           string `json:"title"`
	Message         string `json:"message"`
}

func (c *Client) ListCheckRunAnnotations(owner, repo string, checkRunID int64) ([]Annotation, error) {
	var annotations []Annotation
	if err := c.Get(fmt.Sprintf("/repos/%s/%s/check-runs/%d/annotations?per_page=100", owner, repo, checkRunID), &annotations); err != nil {
		return nil, err
	}
	return annotations, nil
}
