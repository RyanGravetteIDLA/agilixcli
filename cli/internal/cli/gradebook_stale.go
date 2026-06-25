// Stale-gradebook / ungraded-backlog finder for Agilix Buzz.
// pp:data-source live
package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newNovelGradebookStaleCmd(flags *rootFlags) *cobra.Command {
	var flagType, flagDomainid, flagEntityid string
	var maxPages, limit int

	cmd := &cobra.Command{
		Use:   "stale",
		Short: "Rank courses by ungraded backlog and oldest unscored work, from live enrollment metrics",
		Long: "Aggregates per-enrollment ungraded counts and oldest-unscored-work dates into a per-course " +
			"backlog ranking across a domain or course.\n\n" +
			"Use to find grading backlogs. For per-teacher engagement/inactivity use 'teacher activity'.",
		Example:     "  agilix-buzz-pp-cli gradebook stale --type domain --domainid 254591853 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would rank ungraded backlogs")
				return nil
			}
			scope, err := resolveScope(flagType, flagDomainid, flagEntityid)
			if err != nil {
				_ = cmd.Usage()
				return usageErr(err)
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			rows, scanned, err := fetchEnrollments(ctx, c, scope, maxPages)
			if err != nil {
				return err
			}
			now := nowUTC()
			type courseBacklog struct {
				CourseID       string `json:"courseid"`
				UngradedTotal  int    `json:"ungraded_total"`
				UngradedRed    int    `json:"ungraded_red"`
				Students       int    `json:"students_with_ungraded"`
				OldestUngraded int    `json:"oldest_ungraded_days"`
			}
			agg := map[string]*courseBacklog{}
			oldestStr := map[string]string{}
			for _, r := range rows {
				if r.UngradedTotal <= 0 && r.OldestWork == "" {
					continue
				}
				cb := agg[r.CourseID]
				if cb == nil {
					cb = &courseBacklog{CourseID: r.CourseID, OldestUngraded: -1}
					agg[r.CourseID] = cb
				}
				cb.UngradedTotal += r.UngradedTotal
				cb.UngradedRed += r.UngradedRed
				if r.UngradedTotal > 0 {
					cb.Students++
				}
				if r.OldestWork != "" {
					if cur, ok := oldestStr[r.CourseID]; !ok || r.OldestWork < cur {
						oldestStr[r.CourseID] = r.OldestWork
					}
				}
			}
			var out []courseBacklog
			for id, cb := range agg {
				if ow, ok := oldestStr[id]; ok {
					e := enrollRow{LastActed: ow}
					if dd := e.idleDays(now); dd >= 0 {
						cb.OldestUngraded = dd
					}
				}
				out = append(out, *cb)
			}
			sort.Slice(out, func(i, j int) bool {
				if out[i].UngradedTotal != out[j].UngradedTotal {
					return out[i].UngradedTotal > out[j].UngradedTotal
				}
				return out[i].OldestUngraded > out[j].OldestUngraded
			})
			if limit > 0 && len(out) > limit {
				out = out[:limit]
			}
			view := struct {
				Scope              string          `json:"scope"`
				ScopeID            string          `json:"scope_id"`
				Courses            []courseBacklog `json:"courses"`
				ScannedEnrollments int             `json:"scanned_enrollments"`
				Note               string          `json:"note,omitempty"`
			}{scope.kind, scope.id, out, scanned, ""}
			if len(out) == 0 {
				view.Courses = []courseBacklog{}
				view.Note = "no ungraded backlog found in scope (or no metrics returned)"
			}
			return flags.printJSON(cmd, view)
		},
	}
	cmd.Flags().StringVar(&flagType, "type", "domain", "Scope: domain or course")
	cmd.Flags().StringVar(&flagDomainid, "domainid", "", "Domain ID (with --type domain)")
	cmd.Flags().StringVar(&flagEntityid, "entityid", "", "Course/section ID (with --type course)")
	cmd.Flags().IntVar(&maxPages, "max-pages", 10, "Max enrollment pages to scan (1000 rows each)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Max courses to return (0 = all)")
	return cmd
}
