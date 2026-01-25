package platform

import (
	"encoding/json"
	"os/exec"
)

// GitLab implements Platform for GitLab repositories
type GitLab struct {
	repoPath string
}

// NewGitLab creates a GitLab platform instance
func NewGitLab(repoPath string) *GitLab {
	return &GitLab{repoPath: repoPath}
}

// glabMR represents the JSON structure from glab mr list
type glabMR struct {
	IID          int    `json:"iid"`
	Title        string `json:"title"`
	SourceBranch string `json:"source_branch"`
	State        string `json:"state"`
	WebURL       string `json:"web_url"`
}

func parseGitLabMRs(data []byte) ([]MR, error) {
	var gMRs []glabMR
	if err := json.Unmarshal(data, &gMRs); err != nil {
		return nil, err
	}

	mrs := make([]MR, len(gMRs))
	for i, mr := range gMRs {
		status := mr.State
		if status == "opened" {
			status = "open"
		}
		mrs[i] = MR{
			Number: mr.IID,
			Title:  mr.Title,
			Branch: mr.SourceBranch,
			Status: status,
			URL:    mr.WebURL,
		}
	}
	return mrs, nil
}

// ListMRs returns merge requests for the given author
func (g *GitLab) ListMRs(author string) ([]MR, error) {
	args := []string{"mr", "list", "-F", "json"}
	if author == "@me" {
		args = append(args, "--author", "@me")
	} else {
		args = append(args, "--author", author)
	}

	cmd := exec.Command("glab", args...)
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseGitLabMRs(out)
}

// GetRepoInfo returns repository information
func (g *GitLab) GetRepoInfo() (RepoInfo, error) {
	cmd := exec.Command("glab", "repo", "view", "-F", "json")
	cmd.Dir = g.repoPath
	out, err := cmd.Output()
	if err != nil {
		return RepoInfo{}, err
	}

	var result struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		DefaultBranch string `json:"default_branch"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return RepoInfo{}, err
	}

	return RepoInfo{
		Name:          result.Name,
		Description:   result.Description,
		Platform:      "gitlab",
		DefaultBranch: result.DefaultBranch,
	}, nil
}
