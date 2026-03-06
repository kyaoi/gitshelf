package cli

import (
	"errors"
	"fmt"

	"github.com/kyaoi/gitshelf/internal/shelf"
	"github.com/spf13/cobra"
)

func newDoctorCommand(ctx *commandContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run integrity checks for .shelf",
		RunE: func(_ *cobra.Command, _ []string) error {
			report, err := shelf.RunDoctor(ctx.rootDir)
			if err == nil {
				fmt.Printf("%s %s\n", uiHeading("doctor:"), uiColor("問題は見つかりませんでした", "32"))
				return nil
			}
			if !errors.Is(err, shelf.ErrDoctorIssues) {
				return err
			}

			fmt.Printf("%s %s\n", uiHeading("doctor:"), uiColor(fmt.Sprintf("%d 件の問題を検出しました", len(report.Issues)), "31"))
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
	return cmd
}
