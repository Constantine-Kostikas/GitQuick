package platform

import (
	"encoding/json"
	"fmt"

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

// glabMRDetail represents the JSON structure from glab mr view
type glabMRDetail struct {
	IID         int    `json:"iid"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// glabDiffStats represents a file's diff statistics from GitLab API
type glabDiffStats struct {
	OldPath   string `json:"old_path"`
	NewPath   string `json:"new_path"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
}

func parseGitLabMRDetail(data []byte) (glabMRDetail, error) {
	var detail glabMRDetail
	if err := json.Unmarshal(data, &detail); err != nil {
		return glabMRDetail{}, err
	}
	return detail, nil
}

func parseGitLabDiffStats(data []byte) ([]FileChange, int, int, error) {
	var stats []glabDiffStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, 0, 0, err
	}

	files := make([]FileChange, len(stats))
	totalAdditions := 0
	totalDeletions := 0

	for i, s := range stats {
		// Use new_path as the file path (handles renames)
		path := s.NewPath
		if path == "" {
			path = s.OldPath
		}
		files[i] = FileChange{
			Path:      path,
			Additions: s.Additions,
			Deletions: s.Deletions,
		}
		totalAdditions += s.Additions
		totalDeletions += s.Deletions
	}

	return files, totalAdditions, totalDeletions, nil
}

// GetMRDetail returns detailed information about a merge request
func (g *GitLab) GetMRDetail(number int) (MRDetail, error) {
	// Get basic MR info using glab mr view
	mrOut, err := cmd.Run(g.repoPath, "glab", "mr", "view", fmt.Sprintf("%d", number), "-F", "json")
	if err != nil {
		return MRDetail{}, err
	}

	detail, err := parseGitLabMRDetail(mrOut)
	if err != nil {
		return MRDetail{}, err
	}

	result := MRDetail{
		Number: detail.IID,
		Title:  detail.Title,
		Body:   detail.Description,
	}

	// Get diff stats using GitLab API
	// The endpoint /projects/:id/merge_requests/:iid/changes returns file-level changes
	diffOut, err := cmd.Run(g.repoPath, "glab", "api", fmt.Sprintf("projects/:id/merge_requests/%d/changes", number))
	if err != nil {
		// If we can't get diff stats, return what we have
		return result, nil
	}

	// The changes endpoint returns the MR with a "changes" array
	var changesResponse struct {
		Changes []glabDiffStats `json:"changes"`
	}
	if err := json.Unmarshal(diffOut, &changesResponse); err != nil {
		// If parsing fails, return what we have
		return result, nil
	}

	// Convert changes to FileChange slice and compute totals
	files := make([]FileChange, len(changesResponse.Changes))
	for i, c := range changesResponse.Changes {
		path := c.NewPath
		if path == "" {
			path = c.OldPath
		}
		files[i] = FileChange{
			Path:      path,
			Additions: c.Additions,
			Deletions: c.Deletions,
		}
		result.Additions += c.Additions
		result.Deletions += c.Deletions
	}
	result.Files = files

	return result, nil
}
