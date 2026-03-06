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
		asJSON bool
	)

	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run integrity checks for .shelf",
		Example: "  shelf doctor\n" +
			"  shelf doctor --fix\n" +
			"  shelf doctor --json",
		RunE: func(_ *cobra.Command, _ []string) error {
			var (
				report shelf.DoctorReport
				err    error
				fixed  int
			)
			if fix {
				rawReport, fixedCount, runErr := shelf.RunDoctorWithFix(ctx.rootDir)
				report = rawReport
				fixed = fixedCount
				err = runErr
			} else {
				rawReport, runErr := shelf.RunDoctor(ctx.rootDir)
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
				payload := map[string]any{
					"ok":          err == nil,
					"fixed_count": fixed,
					"issue_count": len(issues),
					"issues":      issues,
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
			return err
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "Apply safe automatic fixes before running checks")
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
	case strings.Contains(msg, "invalid toml"), strings.Contains(msg, "failed to parse"), strings.Contains(msg, "failed to read file"):
		return "ファイルを手動で修正し、必要なら `shelf doctor --fix` を再実行してください"
	default:
		return ""
	}
}
