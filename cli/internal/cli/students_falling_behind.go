// At-risk / falling-behind student detector for Agilix Buzz.
// pp:data-source live
package cli

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

func newNovelStudentsFallingBehindCmd(flags *rootFlags) *cobra.Command {
	var flagType, flagDomainid, flagEntityid string
	var maxPages, limit, idleCutoff int
	var failThreshold float64

	cmd := &cobra.Command{
		Use:   "falling-behind",
		Short: "Rank at-risk students by idle days, low completion, pace, and grade below a threshold",
		Long: "Ranks individual students by a composite risk score combining days idle, completion ratio, " +
			"pace-late/past signals, failing flags, and final score below --fail-below.\n\n" +
			"Use to triage individual students. For section-level aggregate health use 'class engagement'.",
		Example:     "  agilix-buzz-pp-cli students falling-behind --type course --entityid 123456 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would rank at-risk students")
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
			type riskRow struct {
				UserID     string   `json:"userid"`
				Name       string   `json:"name"`
				CourseID   string   `json:"courseid"`
				Risk       float64  `json:"risk_score"`
				IdleDays   int      `json:"idle_days"`
				Completion float64  `json:"completion_ratio"`
				FinalScore float64  `json:"final_score"`
				Letter     string   `json:"final_letter,omitempty"`
				Signals    []string `json:"signals"`
			}
			var ranked []riskRow
			for _, r := range rows {
				if r.isTeacher() {
					continue
				}
				d := r.idleDays(now)
				risk := 0.0
				var signals []string
				if d < 0 {
					risk += 3
					signals = append(signals, "never-active")
				} else if d > idleCutoff {
					risk += float64(d) / 7.0
					signals = append(signals, fmt.Sprintf("idle-%dd", d))
				}
				comp := r.completionRatio()
				if r.Completable > 0 && comp < 0.5 {
					risk += (0.5 - comp) * 4
					signals = append(signals, "low-completion")
				}
				if r.PacePast > 0 {
					risk += 2
					signals = append(signals, "pace-past")
				} else if r.PaceLate > 0 {
					risk += 1
					signals = append(signals, "pace-late")
				}
				if r.Failing > 0 || r.Failed > 0 {
					risk += 2
					signals = append(signals, "failing")
				}
				if r.HasMetrics && r.FinalScore > 0 && r.FinalScore < failThreshold {
					risk += (failThreshold - r.FinalScore) / 20.0
					signals = append(signals, "low-grade")
				}
				if risk <= 0 {
					continue
				}
				ranked = append(ranked, riskRow{
					UserID: r.UserID, Name: r.displayName(), CourseID: r.CourseID,
					Risk: round1(risk), IdleDays: d, Completion: round1(comp),
					FinalScore: round1(r.FinalScore), Letter: r.FinalLetter, Signals: signals,
				})
			}
			sort.Slice(ranked, func(i, j int) bool { return ranked[i].Risk > ranked[j].Risk })
			if limit > 0 && len(ranked) > limit {
				ranked = ranked[:limit]
			}
			view := struct {
				Scope              string    `json:"scope"`
				ScopeID            string    `json:"scope_id"`
				AtRisk             []riskRow `json:"at_risk"`
				ScannedEnrollments int       `json:"scanned_enrollments"`
				Note               string    `json:"note,omitempty"`
			}{scope.kind, scope.id, ranked, scanned, ""}
			if len(ranked) == 0 {
				view.AtRisk = []riskRow{}
				view.Note = "no at-risk students detected in scope (or no metrics returned)"
			}
			return flags.printJSON(cmd, view)
		},
	}
	cmd.Flags().StringVar(&flagType, "type", "course", "Scope: course or domain")
	cmd.Flags().StringVar(&flagEntityid, "entityid", "", "Course/section ID (with --type course)")
	cmd.Flags().StringVar(&flagDomainid, "domainid", "", "Domain ID (with --type domain)")
	cmd.Flags().IntVar(&maxPages, "max-pages", 10, "Max enrollment pages to scan (1000 rows each)")
	cmd.Flags().IntVar(&limit, "limit", 50, "Max at-risk students to return (0 = all)")
	cmd.Flags().IntVar(&idleCutoff, "idle-days", 7, "Days idle before contributing to risk")
	cmd.Flags().Float64Var(&failThreshold, "fail-below", 70, "Final score below this counts toward risk")
	return cmd
}
