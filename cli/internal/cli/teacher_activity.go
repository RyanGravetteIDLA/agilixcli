// Teacher engagement & grading-activity report for Agilix Buzz.
// pp:data-source live
package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newNovelTeacherActivityCmd(flags *rootFlags) *cobra.Command {
	var flagDomainid string
	var maxPages, inactiveDays int
	var onlyInactive bool

	cmd := &cobra.Command{
		Use:   "activity",
		Short: "Per-teacher sections, days since last activity, ungraded backlog, oldest unscored work; flags inactive teachers",
		Long: "Aggregates teacher/grader enrollments across a domain into a per-teacher engagement report: " +
			"active section count, days since last activity, total ungraded items, oldest unscored work, and an " +
			"inactive flag.\n\nUse for teacher engagement and inactivity. For the ungraded pile by course use 'gradebook stale'.",
		Example:     "  agilix-buzz-pp-cli teacher activity --domainid 254591853 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would aggregate teacher activity")
				return nil
			}
			if flagDomainid == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--domainid is required"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			rows, scanned, err := fetchEnrollments(ctx, c, enrollScope{kind: "domain", id: flagDomainid, allDesc: true}, maxPages)
			if err != nil {
				return err
			}
			now := nowUTC()
			type teach struct {
				UserID         string `json:"userid"`
				Name           string `json:"name"`
				Sections       int    `json:"sections"`
				LastActiveDays int    `json:"last_active_days"`
				UngradedTotal  int    `json:"ungraded_total"`
				OldestUngraded int    `json:"oldest_ungraded_days"`
				Inactive       bool   `json:"inactive"`
			}
			agg := map[string]*teach{}
			oldest := map[string]string{}
			for _, r := range rows {
				if !r.isTeacher() {
					continue
				}
				t := agg[r.UserID]
				if t == nil {
					t = &teach{UserID: r.UserID, Name: r.displayName(), LastActiveDays: -1, OldestUngraded: -1}
					agg[r.UserID] = t
				}
				t.Sections++
				t.UngradedTotal += r.UngradedTotal
				d := r.idleDays(now)
				if d >= 0 && (t.LastActiveDays < 0 || d < t.LastActiveDays) {
					t.LastActiveDays = d
				}
				if r.OldestWork != "" {
					if cur, ok := oldest[r.UserID]; !ok || r.OldestWork < cur {
						oldest[r.UserID] = r.OldestWork
					}
				}
			}
			var out []teach
			for id, t := range agg {
				if ow, ok := oldest[id]; ok {
					e := enrollRow{LastActed: ow}
					if dd := e.idleDays(now); dd >= 0 {
						t.OldestUngraded = dd
					}
				}
				t.Inactive = t.LastActiveDays < 0 || t.LastActiveDays > inactiveDays
				if onlyInactive && !t.Inactive {
					continue
				}
				out = append(out, *t)
			}
			sort.Slice(out, func(i, j int) bool {
				if out[i].UngradedTotal != out[j].UngradedTotal {
					return out[i].UngradedTotal > out[j].UngradedTotal
				}
				return out[i].LastActiveDays > out[j].LastActiveDays
			})
			view := struct {
				Domain             string  `json:"domainid"`
				Teachers           []teach `json:"teachers"`
				ScannedEnrollments int     `json:"scanned_enrollments"`
				Note               string  `json:"note,omitempty"`
			}{flagDomainid, out, scanned, ""}
			if len(out) == 0 {
				view.Teachers = []teach{}
				view.Note = "no teacher enrollments found in scope (or no metrics returned)"
			}
			return flags.printJSON(cmd, view)
		},
	}
	cmd.Flags().StringVar(&flagDomainid, "domainid", "", "Domain ID to report on (walks descendant domains)")
	cmd.Flags().IntVar(&maxPages, "max-pages", 10, "Max enrollment pages to scan (1000 rows each)")
	cmd.Flags().IntVar(&inactiveDays, "inactive-days", 14, "Days without activity before a teacher is flagged inactive")
	cmd.Flags().BoolVar(&onlyInactive, "only-inactive", false, "Only return teachers flagged inactive")
	return cmd
}
