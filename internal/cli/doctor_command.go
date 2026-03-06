package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newDoctorCommand(ctx *commandContext) *cobra.Command {
	var (
		fix    bool
		strict bool
		asJSON bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run integrity checks for .shelf",
		Example: "  shelf doctor\n" +
			"  shelf doctor --fix\n" +
			"  shelf doctor --strict\n" +
			"  shelf doctor --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				report shelf.DoctorReport
				err    error
				fixed  int
			)
			opts := shelf.DoctorOptions{Strict: strict}
			if fix {
				lockErr := withWriteLock(ctx.rootDir, func() error {
					rawReport, fixedCount, runErr := shelf.RunDoctorWithFixOptions(ctx.rootDir, opts)
					report = rawReport
					fixed = fixedCount
					err = runErr
					return nil
				})
				if lockErr != nil {
					return lockErr
				}
			} else {
				rawReport, runErr := shelf.RunDoctorWithOptions(ctx.rootDir, opts)
				report = rawReport
				err = runErr
			}

			if err != nil && !errors.Is(err, shelf.ErrDoctorIssues) {
				return err
			}

			if asJSON {
				type issueWithAdvice struct {
					Path    string `json:"path"`
					TaskID  string `json:"task_id,omitempty"`
					Message string `json:"message"`
					Advice  string `json:"advice,omitempty"`
				}
				issues := make([]issueWithAdvice, 0, len(report.Issues))
				for _, issue := range report.Issues {
					issues = append(issues, issueWithAdvice{
						Path:    issue.Path,
						TaskID:  issue.TaskID,
						Message: issue.Message,
						Advice:  buildDoctorAdvice(issue),
					})
				}
				warnings := make([]issueWithAdvice, 0, len(report.Warnings))
				for _, warning := range report.Warnings {
					warnings = append(warnings, issueWithAdvice{
						Path:    warning.Path,
						TaskID:  warning.TaskID,
						Message: warning.Message,
						Advice:  buildDoctorAdvice(warning),
					})
				}
				payload := map[string]any{
					"ok":            err == nil,
					"fixed_count":   fixed,
					"issue_count":   len(issues),
					"warning_count": len(warnings),
					"issues":        issues,
					"warnings":      warnings,
				}
				data, marshalErr := json.MarshalIndent(payload, "", "  ")
				if marshalErr != nil {
					return marshalErr
				}
				fmt.Println(string(data))
				return err
			}

			if err == nil {
				if fix {
					fmt.Printf("%s %s (fixed=%d)\n", uiHeading("doctor:"), uiColor("問題は見つかりませんでした", "32"), fixed)
				} else {
					fmt.Printf("%s %s\n", uiHeading("doctor:"), uiColor("問題は見つかりませんでした", "32"))
				}
				if len(report.Warnings) > 0 {
					fmt.Printf("%s %s\n", uiHeading("doctor:"), uiColor(fmt.Sprintf("%d 件の警告を検出しました", len(report.Warnings)), "33"))
					for _, warning := range report.Warnings {
						if warning.TaskID != "" {
							fmt.Printf("- %s (%s): %s\n", warning.Path, uiShortID(shelf.ShortID(warning.TaskID)), warning.Message)
						} else {
							fmt.Printf("- %s: %s\n", warning.Path, warning.Message)
						}
						if advice := buildDoctorAdvice(warning); advice != "" {
							fmt.Printf("  hint: %s\n", advice)
						}
					}
				}
				return nil
			}

			fmt.Printf("%s %s\n", uiHeading("doctor:"), uiColor(fmt.Sprintf("%d 件の問題を検出しました", len(report.Issues)), "31"))
			if fix {
				fmt.Printf("fixed=%d\n", fixed)
			}
			for _, issue := range report.Issues {
				if issue.TaskID != "" {
					fmt.Printf("- %s (%s): %s\n", issue.Path, uiShortID(shelf.ShortID(issue.TaskID)), issue.Message)
				} else {
					fmt.Printf("- %s: %s\n", issue.Path, issue.Message)
				}
				if advice := buildDoctorAdvice(issue); advice != "" {
					fmt.Printf("  hint: %s\n", advice)
				}
			}
			if len(report.Warnings) > 0 {
				fmt.Printf("%s %s\n", uiHeading("doctor:"), uiColor(fmt.Sprintf("%d 件の警告を検出しました", len(report.Warnings)), "33"))
				for _, warning := range report.Warnings {
					if warning.TaskID != "" {
						fmt.Printf("- %s (%s): %s\n", warning.Path, uiShortID(shelf.ShortID(warning.TaskID)), warning.Message)
					} else {
						fmt.Printf("- %s: %s\n", warning.Path, warning.Message)
					}
					if advice := buildDoctorAdvice(warning); advice != "" {
						fmt.Printf("  hint: %s\n", advice)
					}
				}
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "Apply safe automatic fixes before running checks")
	cmd.Flags().BoolVar(&strict, "strict", false, "Enable stricter warnings (e.g. todo without due_on)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}

func buildDoctorAdvice(issue shelf.DoctorIssue) string {
	msg := strings.ToLower(issue.Message)
	switch {
	case strings.Contains(msg, "unknown kind"):
		return "config の kinds を確認するか、`shelf set <id> --kind <known-kind>` で修正してください"
	case strings.Contains(msg, "unknown status"):
		return "config の statuses を確認するか、`shelf set <id> --status <known-status>` で修正してください"
	case strings.Contains(msg, "parent does not exist"):
		return "`shelf mv <id> --parent root` で root に戻すか、正しい parent ID に付け替えてください"
	case strings.Contains(msg, "parent cycle detected"):
		return "`shelf mv` で循環しない親子関係へ移動してください"
	case strings.Contains(msg, "edge destination does not exist"):
		return "壊れたリンクを `shelf unlink --from <id> --to <id> --type <depends_on|related>` で削除してください"
	case strings.Contains(msg, "source task does not exist"):
		return "対応する edge file を削除するか、source task を復元してください"
	case strings.Contains(msg, "duplicate edge found"):
		return "`shelf doctor --fix` で重複 edge を正規化できます"
	case strings.Contains(msg, "invalid github_url"), strings.Contains(msg, "github_url is not canonical"):
		return "`shelf github unlink` / `shelf github link` で正しい GitHub issue / pull request URL に付け替えてください"
	case strings.Contains(msg, "todo task has no due_on"):
		return "`shelf set <id> --due today` や `snooze --to` で期限を設定してください"
	case strings.Contains(msg, "invalid toml"), strings.Contains(msg, "failed to parse"), strings.Contains(msg, "failed to read file"):
		return "ファイルを手動で修正し、必要なら `shelf doctor --fix` を再実行してください"
	default:
		return ""
	}
}
