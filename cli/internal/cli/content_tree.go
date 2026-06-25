// Course content-structure viewer for Agilix Buzz.
// pp:data-source live
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNovelContentTreeCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree <courseid>",
		Short: "Render a course's content hierarchy with item type, points, and rollups",
		Long: "Reads a course's manifest (GetItemList) and renders the item hierarchy with type and points, " +
			"plus a summary rollup (counts by type, gradable count, total points).\n\n" +
			"Use to review one course's content. To compare two courses use 'content diff'.",
		Example:     "  agilix-buzz-pp-cli content tree 244050653 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would render content tree")
				return nil
			}
			if len(args) < 1 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("a <courseid> argument is required"))
			}
			courseID := args[0]
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			items, err := fetchItems(ctx, c, courseID)
			if err != nil {
				return err
			}
			sortItemsBySeq(items)
			view := struct {
				CourseID string         `json:"courseid"`
				Summary  contentSummary `json:"summary"`
				Items    []contentItem  `json:"items"`
				Note     string         `json:"note,omitempty"`
			}{courseID, summarize(items), items, ""}
			if len(items) == 0 {
				view.Items = []contentItem{}
				view.Note = "no items found; check the course id and that you have ReadCourse rights"
			}
			return flags.printJSON(cmd, view)
		},
	}
	return cmd
}
