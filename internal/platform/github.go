package platform

import (
	"encoding/json"
	"strings"

	"gitHelper/internal/cmd"
)

// GitHub implements Platform for GitHub repositories
type GitHub struct {
	repoPath string
}

// NewGitHub creates a GitHub platform instance
func NewGitHub(repoPath string) *GitHub {
	return &GitHub{repoPath: repoPath}
}

// ghPR represents the JSON structure from gh pr list
type ghPR struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	HeadRefName string `json:"headRefName"`
	State       string `json:"state"`
	URL         string `json:"url"`
}

func parseGitHubMRs(data []byte) ([]MR, error) {
	var prs []ghPR
	if err := json.Unmarshal(data, &prs); err != nil {
		return nil, err
	}

	mrs := make([]MR, len(prs))
	for i, pr := range prs {
		mrs[i] = MR{
			Number: pr.Number,
			Title:  pr.Title,
			Branch: pr.HeadRefName,
			Status: strings.ToLower(pr.State),
			URL:    pr.URL,
		}
	}
	return mrs, nil
}

// ListMRs returns pull requests for the given author
func (g *GitHub) ListMRs(author string) ([]MR, error) {
	out, err := cmd.Run(g.repoPath, "gh", "pr", "list",
		"--author", author,
		"--json", "number,title,headRefName,state,url",
	)
	if err != nil {
		return nil, err
	}
	return parseGitHubMRs(out)
}

// GetRepoInfo returns repository information
func (g *GitHub) GetRepoInfo() (RepoInfo, error) {
	out, err := cmd.Run(g.repoPath, "gh", "repo", "view", "--json", "name,description,defaultBranchRef")
	if err != nil {
		return RepoInfo{}, err
	}

	var result struct {
		Name             string `json:"name"`
		Description      string `json:"description"`
		DefaultBranchRef struct {
			Name string `json:"name"`
		} `json:"defaultBranchRef"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return RepoInfo{}, err
	}

	return RepoInfo{
		Name:          result.Name,
		Description:   result.Description,
		Platform:      "github",
		DefaultBranch: result.DefaultBranchRef.Name,
	}, nil
}

// ListAuthors returns repository contributors
func (g *GitHub) ListAuthors() ([]Author, error) {
	out, err := cmd.Run(g.repoPath, "gh", "api", "repos/{owner}/{repo}/contributors", "--paginate", "-q", ".[].login")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	authors := make([]Author, 0, len(lines))
	for _, line := range lines {
		if line != "" {
			authors = append(authors, Author{
				Username: line,
				Name:     line,
			})
		}
	}
	return authors, nil
}
