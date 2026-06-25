// Hand-authored to close a coverage gap against the complete DLAP reference
// (api.agilixbuzz.com/docs/all.md): GetDomainSettings, a current (non-deprecated)
// command the generated surface missed. Follows the generated endpoint pattern.
package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newDomainsGetSettingsCmd(flags *rootFlags) *cobra.Command {
	var flagDomainid string
	var flagPath string
	var flagIncludeSource bool

	cmd := &cobra.Command{
		Use:   "get-settings",
		Short: "Retrieve domain-hierarchy-merged settings for an application (no auth required).",
		Long: "Loads the settings resource at --path from --domainid and each ancestor domain, merging them " +
			"(child-most wins). Requires no authentication. Returns an empty settings element if the path is not found.",
		Example:     "  agilix-buzz-pp-cli domains get-settings --domainid 254591853 --path AgilixBuzzSettings.xml --json",
		Annotations: map[string]string{"pp:endpoint": "domains.get-settings", "pp:method": "GET", "pp:path": "/cmd?cmd=getdomainsettings", "mcp:read-only": "true"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && cmd.Flags().NFlag() == 0 {
				return cmd.Help()
			}
			if flagDomainid == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--domainid is required"))
			}
			if flagPath == "" {
				_ = cmd.Usage()
				return usageErr(fmt.Errorf("--path is required (e.g. AgilixBuzzSettings.xml)"))
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			path := "/cmd?cmd=getdomainsettings"
			params := map[string]string{}
			if flagDomainid != "" {
				params["domainid"] = formatCLIParamValue(flagDomainid)
			}
			if flagPath != "" {
				params["path"] = formatCLIParamValue(flagPath)
			}
			if flagIncludeSource {
				params["includesource"] = "true"
			}
			data, prov, err := resolveReadWithStrategy(cmd.Context(), c, flags, "auto", "domains", false, path, params, nil, cmd.ErrOrStderr())
			if err != nil {
				return classifyAPIError(err, flags)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				var countItems []json.RawMessage
				_ = json.Unmarshal(data, &countItems)
				printProvenance(cmd, len(countItems), prov)
			}
			if flags.asJSON || (!isTerminal(cmd.OutOrStdout()) && !flags.csv && !flags.quiet && !flags.plain) {
				filtered := data
				if flags.selectFields != "" {
					filtered = filterFields(filtered, flags.selectFields)
				} else if flags.compact {
					filtered = compactFields(filtered)
				}
				wrapped, wrapErr := wrapWithProvenance(filtered, prov)
				if wrapErr != nil {
					return wrapErr
				}
				return printOutput(cmd.OutOrStdout(), wrapped, true)
			}
			if wantsHumanTable(cmd.OutOrStdout(), flags) {
				var items []map[string]any
				if json.Unmarshal(data, &items) == nil && len(items) > 0 {
					if err := printAutoTable(cmd.OutOrStdout(), items); err != nil {
						return err
					}
					if len(items) >= 25 {
						fmt.Fprintf(os.Stderr, "\nShowing %d results. To narrow: add --limit, --json --select, or filter flags.\n", len(items))
					}
					return nil
				}
			}
			return printOutputWithFlags(cmd.OutOrStdout(), data, flags)
		},
	}
	cmd.Flags().StringVar(&flagDomainid, "domainid", "", "Domain ID to load settings for (required)")
	cmd.Flags().StringVar(&flagPath, "path", "", "Settings resource path, e.g. AgilixBuzzSettings.xml (required)")
	cmd.Flags().BoolVar(&flagIncludeSource, "includesource", false, "Include source-domainid attributes showing where each node was set")
	return cmd
}
