package platform

import (
	"encoding/json"
	"fmt"
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

// ghPRDetail represents the JSON structure from gh pr view
type ghPRDetail struct {
	Title     string `json:"title"`
	Body      string `json:"body"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Files     []struct {
		Path      string `json:"path"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
	} `json:"files"`
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

func parseGitHubMRDetail(number int, data []byte) (MRDetail, error) {
	var pr ghPRDetail
	if err := json.Unmarshal(data, &pr); err != nil {
		return MRDetail{}, err
	}

	files := make([]FileChange, len(pr.Files))
	for i, f := range pr.Files {
		files[i] = FileChange{
			Path:      f.Path,
			Additions: f.Additions,
			Deletions: f.Deletions,
		}
	}

	return MRDetail{
		Number:    number,
		Title:     pr.Title,
		Body:      pr.Body,
		Files:     files,
		Additions: pr.Additions,
		Deletions: pr.Deletions,
	}, nil
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

// GetMRDetail returns detailed information about a pull request
func (g *GitHub) GetMRDetail(number int) (MRDetail, error) {
	out, err := cmd.Run(g.repoPath, "gh", "pr", "view",
		fmt.Sprintf("%d", number),
		"--json", "title,body,files,additions,deletions",
	)
	if err != nil {
		return MRDetail{}, err
	}
	return parseGitHubMRDetail(number, out)
}

// ghCommit represents the JSON structure for commits from gh pr view
type ghCommit struct {
	OID             string `json:"oid"`
	MessageHeadline string `json:"messageHeadline"`
	Authors         []struct {
		Name string `json:"name"`
	} `json:"authors"`
	CommittedDate string `json:"committedDate"`
}

// GetMRCommits returns commits for a pull request
func (g *GitHub) GetMRCommits(number int) ([]Commit, error) {
	out, err := cmd.Run(g.repoPath, "gh", "pr", "view",
		fmt.Sprintf("%d", number),
		"--json", "commits",
	)
	if err != nil {
		return nil, err
	}

	var result struct {
		Commits []ghCommit `json:"commits"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}

	commits := make([]Commit, len(result.Commits))
	for i, c := range result.Commits {
		author := ""
		if len(c.Authors) > 0 {
			author = c.Authors[0].Name
		}
		commits[i] = Commit{
			SHA:     c.OID[:7], // Short SHA
			Message: c.MessageHeadline,
			Author:  author,
			Date:    formatDate(c.CommittedDate),
		}
	}
	return commits, nil
}

// formatDate formats an ISO date string to a shorter format
func formatDate(isoDate string) string {
	// Input: "2024-01-15T10:30:00Z", Output: "2024-01-15"
	if len(isoDate) >= 10 {
		return isoDate[:10]
	}
	return isoDate
}
