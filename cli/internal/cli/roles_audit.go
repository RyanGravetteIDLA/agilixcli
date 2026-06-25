// Tree-wide role audit for Agilix Buzz.
// pp:data-source live
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"agilix-buzz-pp-cli/internal/client"
	"github.com/spf13/cobra"
)

func newNovelRolesAuditCmd(flags *rootFlags) *cobra.Command {
	var flagUserid, flagType string

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "List every enrollment, role, and entity a user holds across the whole domain tree",
		Long: "Reads all of a user's enrollments (ListUserEnrollments, all statuses) and reports every entity " +
			"they are bound to with its role and privileges — the 'what does this person have, everywhere' view " +
			"that GetEffectiveRights only answers one entity at a time.",
		Example:     "  agilix-buzz-pp-cli roles audit --type user --userid 267876066 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would audit a user's roles")
				return nil
			}
			if t := strings.ToLower(strings.TrimSpace(flagType)); t != "" && t != "user" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--type %q is not supported; only --type user is available in this version", flagType))
			}
			if flagUserid == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--userid is required"))
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			rows, err := fetchUserEnrollments(ctx, c, flagUserid)
			if err != nil {
				return err
			}
			type assign struct {
				EntityID   string `json:"entityid"`
				CourseID   string `json:"courseid,omitempty"`
				DomainID   string `json:"domainid,omitempty"`
				RoleID     string `json:"roleid,omitempty"`
				Privileges string `json:"privileges,omitempty"`
				Status     string `json:"status,omitempty"`
			}
			out := make([]assign, 0, len(rows))
			roleCounts := map[string]int{}
			for _, m := range rows {
				a := assign{
					EntityID:   rawStr(m, "entityid"),
					CourseID:   rawStr(m, "courseid"),
					DomainID:   rawStr(m, "domainid"),
					RoleID:     rawStr(m, "roleid"),
					Privileges: rawStr(m, "privileges"),
					Status:     rawStr(m, "status"),
				}
				if a.EntityID == "" {
					a.EntityID = a.CourseID
				}
				if a.RoleID != "" {
					roleCounts[a.RoleID]++
				}
				out = append(out, a)
			}
			view := struct {
				UserID        string         `json:"userid"`
				Assignments   []assign       `json:"assignments"`
				RoleUsage     map[string]int `json:"role_usage"`
				TotalEntities int            `json:"total_entities"`
				Note          string         `json:"note,omitempty"`
			}{flagUserid, out, roleCounts, len(out), ""}
			if len(out) == 0 {
				view.Assignments = []assign{}
				view.Note = "no enrollments found for this user (or insufficient rights to read them)"
			}
			return flags.printJSON(cmd, view)
		},
	}
	cmd.Flags().StringVar(&flagType, "type", "user", "Audit type (only 'user' is supported in this version)")
	cmd.Flags().StringVar(&flagUserid, "userid", "", "User ID to audit (supports extended IDs like domainid//username)")
	return cmd
}

// fetchUserEnrollments lists all enrollments for a user (all statuses).
func fetchUserEnrollments(ctx context.Context, c *client.Client, userID string) ([]map[string]json.RawMessage, error) {
	body, err := c.Get(ctx, "/cmd?cmd=listuserenrollments", map[string]string{
		"userid":    userID,
		"allstatus": "true",
		"select":    "course",
	})
	if err != nil {
		return nil, err
	}
	raws, err := client.DLAPList(body, "enrollments", "enrollment")
	if err != nil {
		return nil, err
	}
	out := make([]map[string]json.RawMessage, 0, len(raws))
	for _, r := range raws {
		var m map[string]json.RawMessage
		if err := json.Unmarshal(r, &m); err == nil {
			out = append(out, m)
		}
	}
	return out, nil
}
