package cli

import (
	"encoding/json"
	"errors"
	"fmt"

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
				payload := map[string]any{
					"ok":          err == nil,
					"fixed_count": fixed,
					"issue_count": len(report.Issues),
					"issues":      report.Issues,
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
			}
			return err
		},
	}
	cmd.Flags().BoolVar(&fix, "fix", false, "Apply safe automatic fixes before running checks")
	cmd.Flags().BoolVar(&asJSON, "json", false, "Output as JSON")
	return cmd
}
