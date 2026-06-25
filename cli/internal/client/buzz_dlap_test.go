package client

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestDlapError(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		wantErr  bool
		wantCode int
	}{
		{"ok", `{"response":{"code":"OK","user":{"id":"1"}}}`, false, 0},
		{"badrequest", `{"response":{"code":"BadRequest","message":"missing domainid"}}`, true, 400},
		{"unauth", `{"response":{"code":"NoAuthentication","message":"no auth"}}`, true, 401},
		{"forbidden", `{"response":{"code":"Forbidden","message":"nope"}}`, true, 403},
		{"notfound", `{"response":{"code":"NotFound"}}`, true, 404},
		{"non-envelope", `{"hello":"world"}`, false, 0},
		{"not-json", `plain text`, false, 0},
		{"empty", ``, false, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := dlapError("GET", "/cmd?cmd=x", []byte(c.body))
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.StatusCode != c.wantCode {
					t.Fatalf("status = %d, want %d", err.StatusCode, c.wantCode)
				}
			} else if err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}

func TestDLAPList(t *testing.T) {
	// array form
	body := json.RawMessage(`{"response":{"code":"OK","users":{"user":[{"id":"1"},{"id":"2"}]}}}`)
	got, err := DLAPList(body, "users", "user")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("array form: got %d, want 2", len(got))
	}
	// single-object form
	body = json.RawMessage(`{"response":{"code":"OK","users":{"user":{"id":"1"}}}}`)
	got, _ = DLAPList(body, "users", "user")
	if len(got) != 1 {
		t.Fatalf("single form: got %d, want 1", len(got))
	}
	// empty collection ({} with no inner array)
	body = json.RawMessage(`{"response":{"code":"OK","enrollments":{}}}`)
	got, _ = DLAPList(body, "enrollments", "enrollment")
	if len(got) != 0 {
		t.Fatalf("empty form: got %d, want 0", len(got))
	}
	if got == nil {
		t.Fatalf("empty form should return non-nil slice")
	}
}

func TestBuzzThrottleDelay(t *testing.T) {
	mk := func(rem, prov string) http.Header {
		h := http.Header{}
		if rem != "" {
			h.Set("X-Provisioned-Ms-Remaining", rem)
		}
		if prov != "" {
			h.Set("X-Provisioned-Ms", prov)
		}
		return h
	}
	if d := buzzThrottleDelay(mk("", "")); d != 0 {
		t.Fatalf("no headers should yield 0, got %v", d)
	}
	if d := buzzThrottleDelay(mk("500", "1000")); d != 0 {
		t.Fatalf("healthy budget should yield 0, got %v", d)
	}
	// remaining below floor → some positive delay
	if d := buzzThrottleDelay(mk("0", "1000")); d <= 0 {
		t.Fatalf("low budget should yield positive delay, got %v", d)
	}
	// negative remaining → larger delay, but capped at 5s
	d := buzzThrottleDelay(mk("-100000", "1000"))
	if d != 5*time.Second {
		t.Fatalf("huge deficit should cap at 5s, got %v", d)
	}
}
