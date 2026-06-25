// DLAP content-upload helper. Several DLAP "put*" commands (putannouncement,
// putmessage, putblog, putwikipage, putstudentsubmission, putteacherresponse,
// putworkinprogress, putpeerresponse, putattemptfile, putmessagepart) take the
// resource CONTENT as the raw POST body (json/xml/zip/text) with id/metadata
// params in the QUERY STRING — exactly like putresource (see files put). The
// generated command scaffold instead serialized ids into a JSON body, which the
// server rejects. runRawContentPut ports the proven files_put.go flow so these
// siblings work: query-string params + raw content from --file/--stdin via
// PostRawResource, with identical partial-failure / output handling.
//
// Hand-authored support for the generated put* commands.
package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

// runRawContentPut executes a DLAP content-upload command. params holds the
// query-string id/metadata arguments; content is read from filePath (--file) or
// stdin (--stdin). contentType overrides the auto-detected MIME type.
func runRawContentPut(cmd *cobra.Command, flags *rootFlags, resource, path string, params map[string]string, filePath string, stdinBody bool, contentType string) error {
	if filePath == "" && !stdinBody {
		_ = cmd.Usage()
		return usageErr(fmt.Errorf("provide resource content via --file <path> or --stdin"))
	}
	if flags.dryRun {
		src := filePath
		if src == "" {
			src = "<stdin>"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "would POST %s?%s  (content from %s)\n(dry run - no request sent)\n", path, encodeParamsForDisplay(params), src)
		return nil
	}

	ct := contentType
	var content []byte
	if filePath != "" {
		b, rerr := os.ReadFile(filePath) // #nosec G304 -- user-named upload file, opened locally; that is the command's purpose
		if rerr != nil {
			return fmt.Errorf("reading --file: %w", rerr)
		}
		content = b
		if ct == "" {
			ct = mime.TypeByExtension(filepath.Ext(filePath))
		}
	} else {
		b, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			return fmt.Errorf("reading stdin: %w", rerr)
		}
		content = b
	}

	c, err := flags.newClient()
	if err != nil {
		return err
	}
	data, statusCode, err := c.PostRawResource(cmd.Context(), path, params, content, ct)
	if err != nil {
		return classifyAPIError(err, flags)
	}

	// Partial-failure detection mirrors files_put.go: surface a non-OK inner
	// DLAP code as a non-zero exit unless --allow-partial-failure downgrades it.
	var partialFailure *partialFailureReport
	if statusCode >= 200 && statusCode < 300 {
		partialFailure = detectPartialFailure(data)
		if partialFailure != nil {
			fmt.Fprintf(os.Stderr, "warning: partial failure detected in %s response: %s\n", resource, partialFailure.Message)
			if len(partialFailure.ResourceNames) > 0 {
				fmt.Fprintf(os.Stderr, "         succeeded: %d operation(s)\n", len(partialFailure.ResourceNames))
			}
		}
	}
	if statusCode >= 200 && statusCode < 300 && (partialFailure == nil || flags.allowPartialFailure) {
		writeMutationResponseToStore(cmd.Context(), resource, data, "")
	}

	if wantsHumanTable(cmd.OutOrStdout(), flags) {
		var items []map[string]any
		if json.Unmarshal(data, &items) == nil && len(items) > 0 {
			if err := printAutoTable(cmd.OutOrStdout(), items); err != nil {
				fmt.Fprintf(os.Stderr, "warning: table rendering failed, falling back to JSON: %v\n", err)
			} else {
				if partialFailure != nil && !flags.allowPartialFailure {
					return partialFailureErr(fmt.Errorf("partial failure in %s response: %s", resource, partialFailure.Message))
				}
				return nil
			}
		} else {
			var wrapped struct {
				Data []map[string]any `json:"data"`
			}
			if json.Unmarshal(data, &wrapped) == nil && len(wrapped.Data) > 0 {
				if err := printAutoTable(cmd.OutOrStdout(), wrapped.Data); err != nil {
					fmt.Fprintf(os.Stderr, "warning: table rendering failed, falling back to JSON: %v\n", err)
				} else {
					if partialFailure != nil && !flags.allowPartialFailure {
						return partialFailureErr(fmt.Errorf("partial failure in %s response: %s", resource, partialFailure.Message))
					}
					return nil
				}
			}
		}
	}

	if flags.asJSON || (!isTerminal(cmd.OutOrStdout()) && !flags.csv && !flags.quiet && !flags.plain) {
		if flags.quiet {
			if partialFailure != nil && !flags.allowPartialFailure {
				return partialFailureErr(fmt.Errorf("partial failure in %s response: %s", resource, partialFailure.Message))
			}
			return nil
		}
		envelope := map[string]any{
			"action":   "post",
			"resource": resource,
			"path":     path,
			"status":   statusCode,
			"success":  statusCode >= 200 && statusCode < 300 && (partialFailure == nil || flags.allowPartialFailure),
		}
		if partialFailure != nil {
			envelope["partial_failure"] = partialFailure
		}
		filtered := data
		if flags.selectFields != "" {
			filtered = filterFields(filtered, flags.selectFields)
		} else if flags.compact {
			filtered = compactFields(filtered)
		}
		if len(filtered) > 0 {
			var parsed any
			if err := json.Unmarshal(filtered, &parsed); err == nil {
				envelope["data"] = parsed
			}
		}
		envelopeJSON, err := json.Marshal(envelope)
		if err != nil {
			return err
		}
		if perr := printOutput(cmd.OutOrStdout(), json.RawMessage(envelopeJSON), true); perr != nil {
			return perr
		}
		if partialFailure != nil && !flags.allowPartialFailure {
			return partialFailureErr(fmt.Errorf("partial failure in %s response: %s", resource, partialFailure.Message))
		}
		return nil
	}

	if perr := printOutputWithFlags(cmd.OutOrStdout(), data, flags); perr != nil {
		return perr
	}
	if partialFailure != nil && !flags.allowPartialFailure {
		return partialFailureErr(fmt.Errorf("partial failure in %s response: %s", resource, partialFailure.Message))
	}
	return nil
}

// putParam adds k=v to params when v is non-empty (keeps query strings tidy).
func putParam(params map[string]string, k, v string) {
	if v != "" {
		params[k] = v
	}
}

// encodeParamsForDisplay renders params as a stable k=v&... string for --dry-run.
func encodeParamsForDisplay(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, "&")
}
