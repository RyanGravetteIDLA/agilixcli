// DLAP batch partial-failure detection. Buzz returns an outer response.code of
// OK for a batch even when individual items in responses.response[] fail with
// their own non-OK code. detectDLAPPartialFailure surfaces those so mutate
// commands exit non-zero by default (callers opt out with --allow-partial-failure).
//
// Hand-authored support for the generated detectPartialFailure (helpers.go).
package cli

import (
	"fmt"
	"strings"
)

func detectDLAPPartialFailure(top map[string]any) *partialFailureReport {
	// Locate responses.response[] under response{}, tolerating the provenance
	// wrappers the output layer may add (data{} / results{}).
	find := func(m map[string]any) []any {
		resp, _ := m["response"].(map[string]any)
		if resp == nil {
			return nil
		}
		rr, _ := resp["responses"].(map[string]any)
		if rr == nil {
			return nil
		}
		a, _ := rr["response"].([]any)
		return a
	}
	arr := find(top)
	if arr == nil {
		if d, ok := top["data"].(map[string]any); ok {
			arr = find(d)
		}
	}
	if arr == nil {
		if r, ok := top["results"].(map[string]any); ok {
			arr = find(r)
		}
	}
	if len(arr) == 0 {
		return nil
	}
	failed := 0
	var firstMsg string
	for _, it := range arr {
		im, ok := it.(map[string]any)
		if !ok {
			continue
		}
		code, _ := im["code"].(string)
		if code != "" && !strings.EqualFold(code, "OK") {
			failed++
			if firstMsg == "" {
				firstMsg = code
				if m, _ := im["message"].(string); m != "" {
					firstMsg = code + ": " + m
				}
			}
		}
	}
	if failed == 0 {
		return nil
	}
	return &partialFailureReport{
		Field:   "dlap.responses",
		Message: fmt.Sprintf("%d of %d operation(s) failed (first: %s)", failed, len(arr), firstMsg),
		Code:    failed,
	}
}
