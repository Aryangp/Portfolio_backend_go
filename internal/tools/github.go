package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubClient handles HTTP communication with the GitHub REST API
type GitHubClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewGitHubClient creates a new GitHub client instance
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		baseURL: "https://api.github.com",
		token:   strings.TrimSpace(token),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GitHubUserProfile represents the sanitized user profile response
type GitHubUserProfile struct {
	Username    string `json:"username"`
	Name        string `json:"name"`
	Bio         string `json:"bio"`
	PublicRepos int    `json:"public_repos"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	ProfileURL  string `json:"profile_url"`
	Company     string `json:"company,omitempty"`
	Location    string `json:"location,omitempty"`
	Blog        string `json:"blog,omitempty"`
}

// GitHubRepositorySummary represents a summarized public repository
type GitHubRepositorySummary struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Stars       int      `json:"stars"`
	Forks       int      `json:"forks"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics,omitempty"`
	IsFork      bool     `json:"is_fork"`
	UpdatedAt   string   `json:"updated_at"`
}

// GitHubRepoDetails represents detailed information for a single repository
type GitHubRepoDetails struct {
	Name          string   `json:"name"`
	FullName      string   `json:"full_name"`
	Description   string   `json:"description"`
	URL           string   `json:"url"`
	Stars         int      `json:"stars"`
	Forks         int      `json:"forks"`
	OpenIssues    int      `json:"open_issues"`
	Language      string   `json:"language"`
	Topics        []string `json:"topics,omitempty"`
	DefaultBranch string   `json:"default_branch"`
	UpdatedAt     string   `json:"updated_at"`
	ReadmeExcerpt string   `json:"readme_excerpt,omitempty"`
}

// GetUserProfile fetches the public GitHub profile for the given username
func (c *GitHubClient) GetUserProfile(ctx context.Context, username string) (*GitHubUserProfile, error) {
	url := fmt.Sprintf("%s/users/%s", c.baseURL, username)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req, "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error (status %d): %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Login       string `json:"login"`
		Name        string `json:"name"`
		Bio         string `json:"bio"`
		PublicRepos int    `json:"public_repos"`
		Followers   int    `json:"followers"`
		Following   int    `json:"following"`
		HTMLURL     string `json:"html_url"`
		Company     string `json:"company"`
		Location    string `json:"location"`
		Blog        string `json:"blog"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse profile response: %w", err)
	}

	return &GitHubUserProfile{
		Username:    raw.Login,
		Name:        raw.Name,
		Bio:         raw.Bio,
		PublicRepos: raw.PublicRepos,
		Followers:   raw.Followers,
		Following:   raw.Following,
		ProfileURL:  raw.HTMLURL,
		Company:     raw.Company,
		Location:    raw.Location,
		Blog:        raw.Blog,
	}, nil
}

// GetRepositories fetches public repositories for the given username
func (c *GitHubClient) GetRepositories(ctx context.Context, username string, limit int, sort string) ([]GitHubRepositorySummary, error) {
	if limit <= 0 || limit > 30 {
		limit = 10
	}
	if sort == "" {
		sort = "updated"
	}

	url := fmt.Sprintf("%s/users/%s/repos?sort=%s&per_page=%d", c.baseURL, username, sort, limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req, "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error (status %d): %s", resp.StatusCode, string(body))
	}

	var rawRepos []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		HTMLURL     string   `json:"html_url"`
		Stargazers  int      `json:"stargazers_count"`
		Forks       int      `json:"forks_count"`
		Language    string   `json:"language"`
		Topics      []string `json:"topics"`
		Fork        bool     `json:"fork"`
		UpdatedAt   string   `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&rawRepos); err != nil {
		return nil, fmt.Errorf("failed to parse repos response: %w", err)
	}

	summaries := make([]GitHubRepositorySummary, 0, len(rawRepos))
	for _, r := range rawRepos {
		summaries = append(summaries, GitHubRepositorySummary{
			Name:        r.Name,
			Description: r.Description,
			URL:         r.HTMLURL,
			Stars:       r.Stargazers,
			Forks:       r.Forks,
			Language:    r.Language,
			Topics:      r.Topics,
			IsFork:      r.Fork,
			UpdatedAt:   r.UpdatedAt,
		})
	}

	return summaries, nil
}

// GetRepoDetails fetches metadata and README excerpt for a specific repository
func (c *GitHubClient) GetRepoDetails(ctx context.Context, username, repoName string) (*GitHubRepoDetails, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", c.baseURL, username, repoName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req, "application/vnd.github.v3+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github api error (status %d): %s", resp.StatusCode, string(body))
	}

	var raw struct {
		Name          string   `json:"name"`
		FullName      string   `json:"full_name"`
		Description   string   `json:"description"`
		HTMLURL       string   `json:"html_url"`
		Stargazers    int      `json:"stargazers_count"`
		Forks         int      `json:"forks_count"`
		OpenIssues    int      `json:"open_issues_count"`
		Language      string   `json:"language"`
		Topics        []string `json:"topics"`
		DefaultBranch string   `json:"default_branch"`
		UpdatedAt     string   `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse repo details: %w", err)
	}

	details := &GitHubRepoDetails{
		Name:          raw.Name,
		FullName:      raw.FullName,
		Description:   raw.Description,
		URL:           raw.HTMLURL,
		Stars:         raw.Stargazers,
		Forks:         raw.Forks,
		OpenIssues:    raw.OpenIssues,
		Language:      raw.Language,
		Topics:        raw.Topics,
		DefaultBranch: raw.DefaultBranch,
		UpdatedAt:     raw.UpdatedAt,
	}

	// Attempt to fetch README excerpt
	readmeReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/repos/%s/%s/readme", c.baseURL, username, repoName), nil)
	if err == nil {
		c.setHeaders(readmeReq, "application/vnd.github.raw")
		readmeResp, readmeErr := c.httpClient.Do(readmeReq)
		if readmeErr == nil && readmeResp.StatusCode == http.StatusOK {
			defer readmeResp.Body.Close()
			readmeBytes, _ := io.ReadAll(io.LimitReader(readmeResp.Body, 4000))
			details.ReadmeExcerpt = string(readmeBytes)
		}
	}

	return details, nil
}

func (c *GitHubClient) setHeaders(req *http.Request, accept string) {
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "portfolio-backend-chatbot")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// -----------------------------------------------------------------------------
// Dynamic Tool Implementations
// -----------------------------------------------------------------------------

// GitHubUserProfileTool fetches Aryan's live profile
type GitHubUserProfileTool struct {
	client          *GitHubClient
	defaultUsername string
}

func (t *GitHubUserProfileTool) Name() string {
	return "get_github_user_profile"
}

func (t *GitHubUserProfileTool) Description() string {
	return "Fetch live public GitHub profile statistics for Aryan Gupta (e.g. bio, follower count, total public repositories, profile link)."
}

func (t *GitHubUserProfileTool) Declaration() FunctionDeclaration {
	return FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &Schema{
			Type: TypeObject,
			Properties: map[string]*Schema{
				"username": {
					Type:        TypeString,
					Description: "Optional GitHub username. Defaults to Aryan Gupta's username ('Aryangp').",
				},
			},
		},
	}
}

func (t *GitHubUserProfileTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	username := t.defaultUsername
	if u, ok := args["username"].(string); ok && strings.TrimSpace(u) != "" {
		username = strings.TrimSpace(u)
	}
	return t.client.GetUserProfile(ctx, username)
}

// GitHubRepositoriesTool fetches list of public repositories
type GitHubRepositoriesTool struct {
	client          *GitHubClient
	defaultUsername string
}

func (t *GitHubRepositoriesTool) Name() string {
	return "get_github_repositories"
}

func (t *GitHubRepositoriesTool) Description() string {
	return "Fetch live list of public GitHub repositories for Aryan Gupta, including star counts, forks, programming language, topics, and descriptions."
}

func (t *GitHubRepositoriesTool) Declaration() FunctionDeclaration {
	return FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &Schema{
			Type: TypeObject,
			Properties: map[string]*Schema{
				"limit": {
					Type:        TypeInteger,
					Description: "Number of repositories to return (default: 10, max: 30).",
				},
				"sort": {
					Type:        TypeString,
					Description: "Sort order: 'updated', 'stars', 'pushed', or 'created' (default: 'updated').",
					Enum:        []string{"updated", "stars", "pushed", "created"},
				},
				"username": {
					Type:        TypeString,
					Description: "Optional GitHub username. Defaults to Aryan Gupta's username ('Aryangp').",
				},
			},
		},
	}
}

func (t *GitHubRepositoriesTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	username := t.defaultUsername
	if u, ok := args["username"].(string); ok && strings.TrimSpace(u) != "" {
		username = strings.TrimSpace(u)
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	} else if l, ok := args["limit"].(int); ok {
		limit = l
	}

	sort := "updated"
	if s, ok := args["sort"].(string); ok && strings.TrimSpace(s) != "" {
		sort = strings.TrimSpace(s)
	}

	return t.client.GetRepositories(ctx, username, limit, sort)
}

// GitHubRepoDetailsTool fetches details and README for a specific repository
type GitHubRepoDetailsTool struct {
	client          *GitHubClient
	defaultUsername string
}

func (t *GitHubRepoDetailsTool) Name() string {
	return "get_github_repo_details"
}

func (t *GitHubRepoDetailsTool) Description() string {
	return "Fetch in-depth details, commit information, and README excerpt for a specific GitHub repository of Aryan Gupta."
}

func (t *GitHubRepoDetailsTool) Declaration() FunctionDeclaration {
	return FunctionDeclaration{
		Name:        t.Name(),
		Description: t.Description(),
		Parameters: &Schema{
			Type:     TypeObject,
			Required: []string{"repo_name"},
			Properties: map[string]*Schema{
				"repo_name": {
					Type:        TypeString,
					Description: "The exact name of the GitHub repository (e.g., 'gpzer', 'nlp-indexing', 'portfolio-backend').",
				},
				"username": {
					Type:        TypeString,
					Description: "Optional GitHub username. Defaults to Aryan Gupta's username ('Aryangp').",
				},
			},
		},
	}
}

func (t *GitHubRepoDetailsTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	username := t.defaultUsername
	if u, ok := args["username"].(string); ok && strings.TrimSpace(u) != "" {
		username = strings.TrimSpace(u)
	}

	repoName, ok := args["repo_name"].(string)
	if !ok || strings.TrimSpace(repoName) == "" {
		return nil, fmt.Errorf("repo_name parameter is required")
	}

	return t.client.GetRepoDetails(ctx, username, strings.TrimSpace(repoName))
}

// NewGitHubTools creates and returns the collection of GitHub tools for registration
func NewGitHubTools(client *GitHubClient, defaultUsername string) []Tool {
	if defaultUsername == "" {
		defaultUsername = "Aryangp"
	}
	return []Tool{
		&GitHubUserProfileTool{client: client, defaultUsername: defaultUsername},
		&GitHubRepositoriesTool{client: client, defaultUsername: defaultUsername},
		&GitHubRepoDetailsTool{client: client, defaultUsername: defaultUsername},
	}
}
