package platform

import (
	"encoding/json"

	"gitHelper/internal/cmd"
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
	out, err := cmd.Run(g.repoPath, "glab", "mr", "list", "-F", "json", "--author", author)
	if err != nil {
		return nil, err
	}
	return parseGitLabMRs(out)
}

// GetRepoInfo returns repository information
func (g *GitLab) GetRepoInfo() (RepoInfo, error) {
	out, err := cmd.Run(g.repoPath, "glab", "repo", "view", "-F", "json")
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

// ListAuthors returns repository members (uses members API for proper usernames)
func (g *GitLab) ListAuthors() ([]Author, error) {
	out, err := cmd.Run(g.repoPath, "glab", "api", "projects/:id/members/all")
	if err != nil {
		return nil, err
	}

	var members []struct {
		Username string `json:"username"`
		Name     string `json:"name"`
	}
	if err := json.Unmarshal(out, &members); err != nil {
		return nil, err
	}

	authors := make([]Author, len(members))
	for i, m := range members {
		authors[i] = Author{
			Username: m.Username,
			Name:     m.Name,
		}
	}
	return authors, nil
}
