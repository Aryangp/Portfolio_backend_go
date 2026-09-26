package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGitHubTools_Execution(t *testing.T) {
	mux := http.NewServeMux()

	// Mock /users/Aryangp
	mux.HandleFunc("/users/Aryangp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"login": "Aryangp",
			"name": "Aryan Gupta",
			"bio": "Software Engineer",
			"public_repos": 12,
			"followers": 25,
			"following": 10,
			"html_url": "https://github.com/Aryangp"
		}`))
	})

	// Mock /users/Aryangp/repos
	mux.HandleFunc("/users/Aryangp/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"name": "gpzer",
				"description": "Programming language",
				"html_url": "https://github.com/Aryangp/gpzer",
				"stargazers_count": 15,
				"forks_count": 2,
				"language": "Python",
				"topics": ["compiler"],
				"fork": false,
				"updated_at": "2026-03-01T12:00:00Z"
			}
		]`))
	})

	// Mock /repos/Aryangp/gpzer
	mux.HandleFunc("/repos/Aryangp/gpzer", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name": "gpzer",
			"full_name": "Aryangp/gpzer",
			"description": "Programming language",
			"html_url": "https://github.com/Aryangp/gpzer",
			"stargazers_count": 15,
			"forks_count": 2,
			"open_issues_count": 0,
			"language": "Python",
			"topics": ["compiler"],
			"default_branch": "main",
			"updated_at": "2026-03-01T12:00:00Z"
		}`))
	})

	// Mock /repos/Aryangp/gpzer/readme
	mux.HandleFunc("/repos/Aryangp/gpzer/readme", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("# gpzer\nDynamic language"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	client := &GitHubClient{
		baseURL:    server.URL,
		httpClient: server.Client(),
	}

	tools := NewGitHubTools(client, "Aryangp")
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}

	reg := NewRegistry()
	reg.RegisterAll(tools...)

	ctx := context.Background()

	// 1. Test get_github_user_profile
	profileRes, err := reg.Execute(ctx, "get_github_user_profile", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error fetching profile: %v", err)
	}
	profile, ok := profileRes.(*GitHubUserProfile)
	if !ok || profile.Username != "Aryangp" || profile.PublicRepos != 12 {
		t.Errorf("unexpected profile result: %+v", profileRes)
	}

	// 2. Test get_github_repositories
	reposRes, err := reg.Execute(ctx, "get_github_repositories", map[string]any{"limit": float64(5)})
	if err != nil {
		t.Fatalf("unexpected error fetching repos: %v", err)
	}
	repos, ok := reposRes.([]GitHubRepositorySummary)
	if !ok || len(repos) != 1 || repos[0].Name != "gpzer" {
		t.Errorf("unexpected repos result: %+v", reposRes)
	}

	// 3. Test get_github_repo_details
	repoRes, err := reg.Execute(ctx, "get_github_repo_details", map[string]any{"repo_name": "gpzer"})
	if err != nil {
		t.Fatalf("unexpected error fetching repo details: %v", err)
	}
	repoDetails, ok := repoRes.(*GitHubRepoDetails)
	if !ok || repoDetails.Name != "gpzer" || repoDetails.ReadmeExcerpt != "# gpzer\nDynamic language" {
		t.Errorf("unexpected repo details result: %+v", repoRes)
	}

	// 4. Test missing required arg for repo details
	_, err = reg.Execute(ctx, "get_github_repo_details", map[string]any{})
	if err == nil {
		t.Fatalf("expected error when repo_name is missing, got nil")
	}
}
