// Course content diff for Agilix Buzz.
// pp:data-source live
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNovelContentDiffCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <courseA> <courseB>",
		Short: "Diff two courses' content: added/removed/changed items and points drift",
		Long: "Compares two courses' manifests by item title and reports items only in A, only in B, items whose " +
			"type or points changed, and the total-points drift between them.\n\n" +
			"Use to compare a master against a derived copy, or a course across terms. To view one course use 'content tree'.",
		Example:     "  agilix-buzz-pp-cli content diff 244050653 244050999 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would diff two courses' content")
				return nil
			}
			if len(args) < 2 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("two course id arguments are required: <courseA> <courseB>"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			itemsA, err := fetchItems(ctx, c, args[0])
			if err != nil {
				return fmt.Errorf("fetching course A: %w", err)
			}
			itemsB, err := fetchItems(ctx, c, args[1])
			if err != nil {
				return fmt.Errorf("fetching course B: %w", err)
			}
			// Index by title (stable across DerivativeSiblingCopy, which reassigns ids).
			idx := func(items []contentItem) map[string]contentItem {
				m := map[string]contentItem{}
				for _, it := range items {
					key := it.Title
					if key == "" {
						key = it.ID
					}
					m[key] = it
				}
				return m
			}
			a, b := idx(itemsA), idx(itemsB)
			type change struct {
				Title string  `json:"title"`
				Field string  `json:"field"`
				A     string  `json:"a"`
				B     string  `json:"b"`
				_     float64 `json:"-"`
			}
			onlyA := []string{}
			onlyB := []string{}
			changed := []change{}
			for k, ia := range a {
				ib, ok := b[k]
				if !ok {
					onlyA = append(onlyA, k)
					continue
				}
				if ia.Type != ib.Type {
					changed = append(changed, change{Title: k, Field: "type", A: ia.Type, B: ib.Type})
				}
				if ia.MaxPoints != ib.MaxPoints {
					changed = append(changed, change{Title: k, Field: "max_points",
						A: fmt.Sprintf("%g", ia.MaxPoints), B: fmt.Sprintf("%g", ib.MaxPoints)})
				}
			}
			for k := range b {
				if _, ok := a[k]; !ok {
					onlyB = append(onlyB, k)
				}
			}
			sumA, sumB := summarize(itemsA), summarize(itemsB)
			view := struct {
				CourseA     string   `json:"course_a"`
				CourseB     string   `json:"course_b"`
				OnlyInA     []string `json:"only_in_a"`
				OnlyInB     []string `json:"only_in_b"`
				Changed     []change `json:"changed"`
				ItemCountA  int      `json:"item_count_a"`
				ItemCountB  int      `json:"item_count_b"`
				PointsA     float64  `json:"points_a"`
				PointsB     float64  `json:"points_b"`
				PointsDrift float64  `json:"points_drift"`
				Identical   bool     `json:"identical"`
			}{
				CourseA: args[0], CourseB: args[1],
				OnlyInA: onlyA, OnlyInB: onlyB, Changed: changed,
				ItemCountA: sumA.TotalItems, ItemCountB: sumB.TotalItems,
				PointsA: sumA.TotalPoints, PointsB: sumB.TotalPoints,
				PointsDrift: round1(sumB.TotalPoints - sumA.TotalPoints),
				Identical:   len(onlyA) == 0 && len(onlyB) == 0 && len(changed) == 0,
			}
			return flags.printJSON(cmd, view)
		},
	}
	return cmd
}
