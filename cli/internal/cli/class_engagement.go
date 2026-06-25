// Macro class/domain engagement analytics for Agilix Buzz.
// pp:data-source live
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newNovelClassEngagementCmd(flags *rootFlags) *cobra.Command {
	var flagType, flagDomainid, flagEntityid string
	var maxPages int
	var idleCutoff int

	cmd := &cobra.Command{
		Use:   "engagement",
		Short: "Macro engagement: active/idle counts, completion & submission rates, grade distribution, time-on-task",
		Long: "Aggregates live enrollment metrics into a class- or domain-level health view: active vs idle " +
			"students, completion and submission rates, an A–F grade-distribution histogram, and mean time-on-task.\n\n" +
			"Use for section/domain aggregate health. For individual at-risk students use 'students falling-behind'.",
		Example:     "  agilix-buzz-pp-cli class engagement --type domain --domainid 254591853 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would aggregate engagement metrics")
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
			var students []enrollRow
			for _, r := range rows {
				if !r.isTeacher() {
					students = append(students, r)
				}
			}
			now := nowUTC()
			active, idle, atRisk := 0, 0, 0
			var sumCompleted, sumCompletable, sumGradable, sumGraded float64
			var times []float64
			for _, s := range students {
				d := s.idleDays(now)
				if d >= 0 && d <= idleCutoff {
					active++
				} else {
					idle++
				}
				sumCompleted += float64(s.Completed)
				sumCompletable += float64(s.Completable)
				sumGradable += float64(s.Gradable)
				sumGraded += float64(s.Graded)
				if s.Seconds > 0 {
					times = append(times, float64(s.Seconds)/3600.0)
				}
				if s.Failing > 0 || s.PacePast > 0 || d < 0 || d > idleCutoff {
					atRisk++
				}
			}
			view := struct {
				Scope              string      `json:"scope"`
				ScopeID            string      `json:"scope_id"`
				Students           int         `json:"students"`
				Active             int         `json:"active"`
				Idle               int         `json:"idle"`
				CompletionRate     float64     `json:"completion_rate"`
				SubmissionRate     float64     `json:"submission_rate"`
				MeanTimeOnTaskHrs  float64     `json:"mean_time_on_task_hours"`
				GradeDistribution  interface{} `json:"grade_distribution"`
				AtRisk             int         `json:"at_risk"`
				ScannedEnrollments int         `json:"scanned_enrollments"`
				Note               string      `json:"note,omitempty"`
			}{
				Scope:              scope.kind,
				ScopeID:            scope.id,
				Students:           len(students),
				Active:             active,
				Idle:               idle,
				CompletionRate:     ratio(sumCompleted, sumCompletable),
				SubmissionRate:     ratio(sumGraded, sumGradable),
				MeanTimeOnTaskHrs:  round1(mean(times)),
				GradeDistribution:  gradeHistogram(students),
				AtRisk:             atRisk,
				ScannedEnrollments: scanned,
			}
			if len(students) == 0 {
				view.Note = "no student enrollments found in scope; check the id and that you have ReadGradebook rights"
			}
			return flags.printJSON(cmd, view)
		},
	}
	cmd.Flags().StringVar(&flagType, "type", "domain", "Scope: domain or course")
	cmd.Flags().StringVar(&flagDomainid, "domainid", "", "Domain ID (with --type domain)")
	cmd.Flags().StringVar(&flagEntityid, "entityid", "", "Course/section ID (with --type course)")
	cmd.Flags().IntVar(&maxPages, "max-pages", 10, "Max enrollment pages to scan (1000 rows each)")
	cmd.Flags().IntVar(&idleCutoff, "idle-days", 14, "Days since last activity before a student counts as idle")
	return cmd
}
