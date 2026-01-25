package platform

// MR represents a merge/pull request
type MR struct {
	Number int
	Title  string
	Branch string
	Status string // "open", "draft", "merged", "closed"
	URL    string
}

// Author represents a repository contributor
type Author struct {
	Username string
	Name     string
}

// RepoInfo contains repository metadata
type RepoInfo struct {
	Name          string
	Description   string
	Platform      string // "github" or "gitlab"
	DefaultBranch string
}

// Platform abstracts GitHub/GitLab operations
type Platform interface {
	ListMRs(author string) ([]MR, error)
	GetRepoInfo() (RepoInfo, error)
	ListAuthors() ([]Author, error)
}
