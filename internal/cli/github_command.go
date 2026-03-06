package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

type githubIssueRef struct {
	Owner  string
	Repo   string
	Number int
	URL    string
}

type githubIssuePayload struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	State   string `json:"state"`
	HTMLURL string `json:"html_url"`
}

func newGitHubCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "github",
		Short: "Manage GitHub links for tasks",
	}
	cmd.AddCommand(newGitHubLinkCommand(ctx))
	cmd.AddCommand(newGitHubUnlinkCommand(ctx))
	cmd.AddCommand(newGitHubShowCommand(ctx))
	return cmd
}

func newGitHubLinkCommand(ctx *commandContext) *cobra.Command {
	var urlValue string
	cmd := &cobra.Command{
		Use:   "link <id>",
		Short: "Attach a GitHub issue or pull request URL to a task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "GitHub URLを紐付けるタスクを選択", nil, true)
			if err != nil {
				return err
			}
			normalized, err := normalizeGitHubURL(urlValue)
			if err != nil {
				return err
			}
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "github-link"); err != nil {
					return err
				}
				_, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{AddGitHubURLs: []string{normalized}})
				if err != nil {
					return err
				}
				fmt.Printf("GitHub linked: %s\n", normalized)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&urlValue, "url", "", "GitHub issue or pull request URL")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

func newGitHubUnlinkCommand(ctx *commandContext) *cobra.Command {
	var urlValue string
	cmd := &cobra.Command{
		Use:   "unlink <id>",
		Short: "Remove a GitHub URL from a task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "GitHub URLを解除するタスクを選択", nil, true)
			if err != nil {
				return err
			}
			normalized, err := normalizeGitHubURL(urlValue)
			if err != nil {
				return err
			}
			return withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "github-unlink"); err != nil {
					return err
				}
				_, err := shelf.SetTask(ctx.rootDir, id, shelf.SetTaskInput{RemoveGitHubURLs: []string{normalized}})
				if err != nil {
					return err
				}
				fmt.Printf("GitHub unlinked: %s\n", normalized)
				return nil
			})
		},
	}
	cmd.Flags().StringVar(&urlValue, "url", "", "GitHub issue or pull request URL")
	_ = cmd.MarkFlagRequired("url")
	return cmd
}

func newGitHubShowCommand(ctx *commandContext) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show GitHub URLs linked to a task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			id, err := selectTaskIDIfMissing(ctx, args, "GitHub URLを表示するタスクを選択", nil, true)
			if err != nil {
				return err
			}
			task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
			if err != nil {
				return err
			}
			if asJSON {
				data, err := json.MarshalIndent(task.GitHubURLs, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
				return nil
			}
			if len(task.GitHubURLs) == 0 {
				fmt.Println("(none)")
				return nil
			}
			for _, item := range task.GitHubURLs {
				fmt.Println(item)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func newSyncCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Synchronize external integrations",
	}
	cmd.AddCommand(newSyncGitHubCommand(ctx))
	return cmd
}

func newSyncGitHubCommand(ctx *commandContext) *cobra.Command {
	var all bool
	cmd := &cobra.Command{
		Use:   "github [id]",
		Short: "Sync linked task metadata from GitHub",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			targets, err := resolveGitHubSyncTargets(ctx, args, all)
			if err != nil {
				return err
			}
			if len(targets) == 0 {
				fmt.Println("No GitHub-linked tasks.")
				return nil
			}
			type githubSyncUpdate struct {
				id     string
				title  string
				status shelf.Status
			}
			updates := make([]githubSyncUpdate, 0, len(targets))
			for _, task := range targets {
				if len(task.GitHubURLs) == 0 {
					continue
				}
				payload, err := fetchGitHubIssue(task.GitHubURLs[0])
				if err != nil {
					return err
				}
				updates = append(updates, githubSyncUpdate{
					id:     task.ID,
					title:  payload.Title,
					status: mapGitHubStateToStatus(payload.State),
				})
			}
			if len(updates) == 0 {
				fmt.Println("No GitHub-linked tasks.")
				return nil
			}
			if err := withWriteLock(ctx.rootDir, func() error {
				if err := prepareUndoSnapshot(ctx.rootDir, "sync-github"); err != nil {
					return err
				}
				for _, update := range updates {
					if _, err := shelf.SetTask(ctx.rootDir, update.id, shelf.SetTaskInput{
						Title:  &update.title,
						Status: &update.status,
					}); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
			fmt.Printf("GitHub synced: %d\n", len(updates))
			return nil
		},
	}
	cmd.Flags().BoolVar(&all, "all", false, "Sync all GitHub-linked tasks")
	return cmd
}

func resolveGitHubSyncTargets(ctx *commandContext, args []string, all bool) ([]shelf.Task, error) {
	if all {
		tasks, err := shelf.NewTaskStore(ctx.rootDir).List()
		if err != nil {
			return nil, err
		}
		result := make([]shelf.Task, 0)
		for _, task := range tasks {
			if len(task.GitHubURLs) > 0 {
				result = append(result, task)
			}
		}
		return result, nil
	}
	id, err := selectTaskIDIfMissing(ctx, args, "GitHub同期するタスクを選択", nil, true)
	if err != nil {
		return nil, err
	}
	task, err := shelf.EnsureTaskExists(ctx.rootDir, id)
	if err != nil {
		return nil, err
	}
	return []shelf.Task{task}, nil
}

func normalizeGitHubURL(value string) (string, error) {
	ref, err := parseGitHubIssueURL(value)
	if err != nil {
		return "", err
	}
	return ref.URL, nil
}

func parseGitHubIssueURL(value string) (githubIssueRef, error) {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return githubIssueRef{}, err
	}
	if !strings.EqualFold(parsed.Host, "github.com") {
		return githubIssueRef{}, fmt.Errorf("GitHub URL is required")
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) != 4 {
		return githubIssueRef{}, fmt.Errorf("unsupported GitHub URL: %s", value)
	}
	if parts[2] != "issues" && parts[2] != "pull" {
		return githubIssueRef{}, fmt.Errorf("unsupported GitHub URL: %s", value)
	}
	number, err := strconv.Atoi(parts[3])
	if err != nil {
		return githubIssueRef{}, fmt.Errorf("invalid GitHub number: %w", err)
	}
	return githubIssueRef{
		Owner:  parts[0],
		Repo:   parts[1],
		Number: number,
		URL:    "https://github.com/" + strings.Join(parts, "/"),
	}, nil
}

func fetchGitHubIssue(rawURL string) (githubIssuePayload, error) {
	ref, err := parseGitHubIssueURL(rawURL)
	if err != nil {
		return githubIssuePayload{}, err
	}
	baseURL := strings.TrimRight(os.Getenv("GITSHELF_GITHUB_API_URL"), "/")
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s/%s/issues/%d", baseURL, ref.Owner, ref.Repo, ref.Number), nil)
	if err != nil {
		return githubIssuePayload{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return githubIssuePayload{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return githubIssuePayload{}, fmt.Errorf("GitHub API returned %s", resp.Status)
	}
	var payload githubIssuePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return githubIssuePayload{}, err
	}
	return payload, nil
}

func mapGitHubStateToStatus(state string) shelf.Status {
	switch strings.ToLower(strings.TrimSpace(state)) {
	case "closed":
		return "done"
	default:
		return "open"
	}
}
