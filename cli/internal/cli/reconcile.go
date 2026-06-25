// Roster-vs-live enrollment reconciliation for Agilix Buzz (read-only).
// pp:data-source live
package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"agilix-buzz-pp-cli/internal/client"
	"github.com/spf13/cobra"
)

// idEntity is a set of identifier keys for one person plus a display label.
type idEntity struct {
	label string
	keys  []string
}

func newNovelReconcileCmd(flags *rootFlags) *cobra.Command {
	var flagDomainid string
	var maxPages int

	cmd := &cobra.Command{
		Use:   "reconcile <roster.csv>",
		Short: "Read-only diff of a roster CSV against live domain enrollments (missing / extra)",
		Long: "Compares a roster CSV (columns: any of userid, username, email, reference) against the live " +
			"enrollments in a domain and reports who is in the roster but not enrolled (missing) and who is " +
			"enrolled but not in the roster (extra). Writes nothing — purely diagnostic.",
		Example:     "  agilix-buzz-pp-cli reconcile roster.csv --domainid 254591853 --agent",
		Annotations: map[string]string{"mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if dryRunOK(flags) {
				fmt.Fprintln(cmd.OutOrStdout(), "would reconcile roster against live enrollments")
				return nil
			}
			if len(args) < 1 {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("a <roster.csv> path argument is required"))
			}
			if flagDomainid == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--domainid is required"))
			}
			roster, err := readRoster(args[0])
			if err != nil {
				return err
			}
			ctx, cancel := boundCtx(cmd.Context(), flags)
			defer cancel()
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			enrolled, scanned, err := fetchEnrolledIdentities(ctx, c, flagDomainid, maxPages)
			if err != nil {
				return err
			}
			enrolledKeys := keySet(enrolled)
			rosterKeys := keySet(roster)

			missing := []string{}
			for _, e := range roster {
				if !anyKeyIn(e.keys, enrolledKeys) {
					missing = append(missing, e.label)
				}
			}
			extra := []string{}
			for _, e := range enrolled {
				if !anyKeyIn(e.keys, rosterKeys) {
					extra = append(extra, e.label)
				}
			}
			view := struct {
				Domain         string   `json:"domainid"`
				RosterCount    int      `json:"roster_count"`
				EnrolledCount  int      `json:"enrolled_count"`
				Missing        []string `json:"missing_from_buzz"`
				Extra          []string `json:"extra_in_buzz"`
				ScannedEnrolls int      `json:"scanned_enrollments"`
			}{flagDomainid, len(roster), len(enrolled), missing, extra, scanned}
			return flags.printJSON(cmd, view)
		},
	}
	cmd.Flags().StringVar(&flagDomainid, "domainid", "", "Domain ID whose enrollments to reconcile against")
	cmd.Flags().IntVar(&maxPages, "max-pages", 10, "Max enrollment pages to scan (1000 rows each)")
	return cmd
}

func idKey(kind, val string) string { return kind + ":" + strings.ToLower(strings.TrimSpace(val)) }

func keySet(entities []idEntity) map[string]bool {
	s := map[string]bool{}
	for _, e := range entities {
		for _, k := range e.keys {
			s[k] = true
		}
	}
	return s
}

func anyKeyIn(keys []string, set map[string]bool) bool {
	for _, k := range keys {
		if set[k] {
			return true
		}
	}
	return false
}

// readRoster parses a roster CSV into one idEntity per row. Recognized header
// columns (case-insensitive): userid, username, email, reference.
func readRoster(path string) ([]idEntity, error) {
	// #nosec G304 -- path is the user's own roster CSV, passed explicitly as the
	// `reconcile <roster.csv>` positional argument and opened on the caller's
	// local machine. Reading the file the operator named is the command's
	// purpose, not an injection vector.
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening roster: %w", err)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parsing roster CSV: %w", err)
	}
	if len(records) < 2 {
		return nil, fmt.Errorf("roster CSV needs a header row and at least one data row")
	}
	col := map[string]int{}
	for i, h := range records[0] {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	idCols := []string{"userid", "username", "email", "reference"}
	hasAny := false
	for _, c := range idCols {
		if _, ok := col[c]; ok {
			hasAny = true
		}
	}
	if !hasAny {
		return nil, fmt.Errorf("roster CSV must have at least one of: userid, username, email, reference")
	}
	var out []idEntity
	for _, row := range records[1:] {
		e := idEntity{}
		for _, c := range idCols {
			if idx, ok := col[c]; ok && idx < len(row) {
				v := strings.TrimSpace(row[idx])
				if v == "" {
					continue
				}
				e.keys = append(e.keys, idKey(c, v))
				if e.label == "" {
					e.label = v
				}
			}
		}
		if len(e.keys) > 0 {
			out = append(out, e)
		}
	}
	return out, nil
}

// fetchEnrolledIdentities lists domain enrollments (with user data) and returns
// one idEntity per enrollment.
func fetchEnrolledIdentities(ctx context.Context, c *client.Client, domainID string, maxPages int) ([]idEntity, int, error) {
	const pageSize = 1000
	var out []idEntity
	scanned := 0
	lastID := ""
	for page := 0; page < maxPages; page++ {
		params := map[string]string{
			"domainid":                 domainID,
			"includedescendantdomains": "true",
			"select":                   "user",
			"limit":                    strconv.Itoa(pageSize),
		}
		if lastID != "" {
			params["query"] = "/id>" + lastID
		}
		body, err := c.Get(ctx, "/cmd?cmd=listenrollments", params)
		if err != nil {
			return out, scanned, err
		}
		raws, err := client.DLAPList(body, "enrollments", "enrollment")
		if err != nil {
			return out, scanned, err
		}
		if len(raws) == 0 {
			break
		}
		for _, raw := range raws {
			var m map[string]json.RawMessage
			if err := json.Unmarshal(raw, &m); err != nil {
				continue
			}
			scanned++
			lastID = rawStr(m, "id")
			e := idEntity{}
			if uid := rawStr(m, "userid"); uid != "" {
				e.keys = append(e.keys, idKey("userid", uid))
				e.label = "user " + uid
			}
			if u := rawObj(m, "user"); u != nil {
				if un := rawStr(u, "username"); un != "" {
					e.keys = append(e.keys, idKey("username", un))
					e.label = un
				}
				if em := rawStr(u, "email"); em != "" {
					e.keys = append(e.keys, idKey("email", em))
				}
				if ref := rawStr(u, "reference"); ref != "" {
					e.keys = append(e.keys, idKey("reference", ref))
				}
			}
			if len(e.keys) > 0 {
				out = append(out, e)
			}
		}
		if len(raws) < pageSize {
			break
		}
	}
	return out, scanned, nil
}
