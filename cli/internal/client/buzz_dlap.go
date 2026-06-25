// DLAP (Data Layer Access Protocol) response handling for Agilix Buzz.
//
// Buzz returns every response inside a {"response":{"code":...}} envelope and
// uses an HTTP 200 even for logical failures — the real status lives in
// response.code (e.g. "OK", "BadRequest", "Unauthorized", "Forbidden",
// "NotFound"). Without inspecting that field, every command would exit 0 on a
// logical error. dlapError translates a non-OK envelope into an *APIError so
// the existing classifyAPIError exit-code mapping applies uniformly across all
// generated endpoint commands.
//
// This file is hand-authored (not generated) and is the single chokepoint for
// DLAP envelope semantics shared by generated and novel commands.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// buzzThrottleDelay implements proactive Buzz API Time-Limiting adaptation. Buzz
// returns X-Provisioned-Ms-Remaining (the processing-time budget left before a
// 429) and X-Provisioned-Ms (the per-second recovery rate). Buzz's own guidance
// is to use these as a speedometer: slow down as remaining approaches zero so
// you stay just under the limit instead of cascading into Too Many Requests
// responses. When remaining drops below a small floor, we sleep just long enough
// for the budget to recover to that floor, bounded so a single call never stalls
// the CLI for long. Returns 0 when the headers are absent or the budget is
// healthy, so non-Buzz hosts and unthrottled calls are unaffected.
func buzzThrottleDelay(h http.Header) time.Duration {
	remStr := strings.TrimSpace(h.Get("X-Provisioned-Ms-Remaining"))
	if remStr == "" {
		return 0
	}
	rem, err := strconv.Atoi(remStr)
	if err != nil {
		return 0
	}
	const floor = 150 // keep at least this much budget in reserve
	if rem >= floor {
		return 0
	}
	prov := 1000
	if p, e := strconv.Atoi(strings.TrimSpace(h.Get("X-Provisioned-Ms"))); e == nil && p > 0 {
		prov = p
	}
	deficit := floor - rem // ms of budget to recover (rem may be negative)
	secs := float64(deficit) / float64(prov)
	d := time.Duration(secs * float64(time.Second))
	const maxDelay = 5 * time.Second
	if d > maxDelay {
		d = maxDelay
	}
	if d < 0 {
		d = 0
	}
	return d
}

// dlapEnvelope is the outer shape of every DLAP response.
type dlapEnvelope struct {
	Response json.RawMessage `json:"response"`
}

type dlapStatus struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	ErrorID string `json:"errorId"`
}

// dlapCodeToHTTP maps a DLAP response.code string to the closest HTTP status so
// the CLI's typed exit codes (auth=4xx, not-found=404, etc.) stay meaningful.
func dlapCodeToHTTP(code string) int {
	switch strings.ToLower(code) {
	case "ok":
		return 200
	case "unauthorized", "loginrequired", "sessionexpired", "invalidtoken", "noauthentication", "authenticationrequired":
		return 401
	case "forbidden", "accessdenied", "insufficientprivileges", "notallowed":
		return 403
	case "notfound", "nosuchentity", "entitynotfound":
		return 404
	case "conflict", "duplicate", "alreadyexists":
		return 409
	case "toomanyrequests", "ratelimited", "throttled":
		return 429
	case "serveroverwhelmed", "servererror", "internalerror":
		return 503
	default:
		// BadRequest and any other unrecognized logical error.
		return 400
	}
}

// dlapError inspects a textual JSON response body. If it is a DLAP envelope with
// a non-OK response.code, it returns an *APIError carrying the mapped HTTP
// status and the upstream message. It returns nil for OK responses, non-DLAP
// bodies, or bodies that do not parse as the envelope (those fall through to the
// normal success path unchanged).
func dlapError(method, displayPath string, body []byte) *APIError {
	trimmed := strings.TrimSpace(string(body))
	if !strings.HasPrefix(trimmed, "{") {
		return nil
	}
	var env dlapEnvelope
	if err := json.Unmarshal(body, &env); err != nil || len(env.Response) == 0 {
		return nil
	}
	var st dlapStatus
	if err := json.Unmarshal(env.Response, &st); err != nil || st.Code == "" {
		return nil
	}
	if strings.EqualFold(st.Code, "OK") {
		return nil
	}
	msg := st.Message
	if msg == "" {
		msg = "request failed"
	}
	detail := fmt.Sprintf("DLAP %s: %s", st.Code, msg)
	if st.ErrorID != "" {
		detail += fmt.Sprintf(" (errorId %s)", st.ErrorID)
	}
	return &APIError{
		Method:     method,
		Path:       displayPath,
		StatusCode: dlapCodeToHTTP(st.Code),
		Body:       detail,
	}
}

// DLAPList extracts the array at response.<plural>.<singular> from a DLAP
// response body, tolerating Buzz's XML→JSON quirks: an empty collection
// serializes as {} (no inner array key), and a single element may serialize as
// an object rather than a one-element array. Returns an empty slice (never nil)
// when no rows are present so callers can range safely and marshal [] not null.
//
// Example: DLAPList(body, "users", "user") for a ListUsers response shaped
// {"response":{"code":"OK","users":{"user":[...]}}}.
func DLAPList(body json.RawMessage, plural, singular string) ([]json.RawMessage, error) {
	out := []json.RawMessage{}
	var env dlapEnvelope
	if err := json.Unmarshal(body, &env); err != nil || len(env.Response) == 0 {
		return out, nil
	}
	// response -> { plural: { singular: [...] } }
	var resp map[string]json.RawMessage
	if err := json.Unmarshal(env.Response, &resp); err != nil {
		return out, nil
	}
	collection, ok := resp[plural]
	if !ok || len(collection) == 0 {
		return out, nil
	}
	var inner map[string]json.RawMessage
	if err := json.Unmarshal(collection, &inner); err != nil {
		return out, nil
	}
	items, ok := inner[singular]
	if !ok || len(items) == 0 {
		return out, nil
	}
	// Array form.
	if strings.HasPrefix(strings.TrimSpace(string(items)), "[") {
		var arr []json.RawMessage
		if err := json.Unmarshal(items, &arr); err != nil {
			return out, err
		}
		return arr, nil
	}
	// Single-object form.
	return []json.RawMessage{items}, nil
}

// DLAPInner returns the payload object inside the response envelope (everything
// under "response"), useful for single-entity reads like GetUser2 →
// response.user. Returns nil when the body is not a DLAP envelope.
func DLAPInner(body json.RawMessage) json.RawMessage {
	var env dlapEnvelope
	if err := json.Unmarshal(body, &env); err != nil || len(env.Response) == 0 {
		return nil
	}
	return env.Response
}

// RawUpload carries a raw (non-JSON) request body plus its content type. Used
// for DLAP PutResource-style uploads where the file bytes ARE the body and the
// entity/path are query-string parameters. doInternal sends RawUpload.Data
// verbatim with RawUpload.ContentType instead of JSON-marshaling.
type RawUpload struct {
	Data        []byte
	ContentType string
}

// PostRawResource uploads raw bytes to a DLAP command, passing params as the
// query string and the bytes as the request body with the given content type.
func (c *Client) PostRawResource(ctx context.Context, path string, params map[string]string, data []byte, contentType string) (json.RawMessage, int, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return c.do(ctx, "POST", path, params, RawUpload{Data: data, ContentType: contentType}, nil)
}
