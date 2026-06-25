// Shared helpers for the Agilix Buzz activity-analytics commands
// (class engagement, students falling-behind, teacher activity, gradebook
// stale). These read live enrollment data with select=metrics,user and
// aggregate in Go — the DLAP API returns per-enrollment rows, never the
// class/domain rollups these commands compute.
//
// Hand-authored. Robust to Buzz's XML→JSON quirks: numeric fields may arrive as
// JSON strings, nested objects (metrics, user) may be absent, and an empty
// collection serializes as {}.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"agilix-buzz-pp-cli/internal/client"
)

// enrollRow is a flattened enrollment + metrics + user record.
type enrollRow struct {
	EnrollmentID string `json:"enrollmentid"`
	UserID       string `json:"userid"`
	CourseID     string `json:"courseid"`
	DomainID     string `json:"domainid"`
	Status       string `json:"status"`
	Privileges   string `json:"privileges,omitempty"`
	RoleID       string `json:"roleid,omitempty"`
	FirstActed   string `json:"firstactivitydate,omitempty"`
	LastActed    string `json:"lastactivitydate,omitempty"`
	// user
	FirstName string `json:"firstname,omitempty"`
	LastName  string `json:"lastname,omitempty"`
	Email     string `json:"email,omitempty"`
	Username  string `json:"username,omitempty"`
	// metrics
	HasMetrics    bool    `json:"-"`
	FinalScore    float64 `json:"finalscore"`
	FinalLetter   string  `json:"finalletter,omitempty"`
	Completed     int     `json:"completed"`
	Completable   int     `json:"completable"`
	Gradable      int     `json:"gradable"`
	Graded        int     `json:"graded"`
	Possible      float64 `json:"possible"`
	UngradedTotal int     `json:"ungradedtotal"`
	UngradedRed   int     `json:"ungradedred"`
	OldestWork    string  `json:"oldestworkitem,omitempty"`
	Seconds       int64   `json:"seconds"`
	Late          int     `json:"late"`
	PaceLate      int     `json:"pacelate"`
	PacePast      int     `json:"pacepast"`
	Failing       int     `json:"failing"`
	Failed        int     `json:"failed"`
}

func (e enrollRow) displayName() string {
	n := strings.TrimSpace(e.FirstName + " " + e.LastName)
	if n == "" {
		n = e.Username
	}
	if n == "" {
		n = "user " + e.UserID
	}
	return n
}

// isTeacher reports whether the enrollment privileges indicate a teacher/grader
// (any grading or course-control bit). DLAP privileges is a bitmask string.
func (e enrollRow) isTeacher() bool {
	p, _ := strconv.ParseInt(e.Privileges, 10, 64)
	// GradeAssignment|GradeExam|GradeForum|SetupGradebook|UpdateCourse bits all
	// live in the high word (0x8F060000 in the documented teacher example).
	const teacherMask = int64(0x8F000000)
	return p&teacherMask != 0
}

func (e enrollRow) completionRatio() float64 {
	if e.Completable <= 0 {
		return 0
	}
	return float64(e.Completed) / float64(e.Completable)
}

// idleDays returns whole days since last activity, or -1 if never active. Buzz
// represents "never active" as the min date 0001-01-01T00:00:00Z rather than an
// empty string, so any parsed date before ~1990 is treated as never-active.
func (e enrollRow) idleDays(now time.Time) int {
	if e.LastActed == "" {
		return -1
	}
	t := parseBuzzTime(e.LastActed)
	if t.IsZero() || t.Year() < 1990 {
		return -1
	}
	d := now.Sub(t).Hours() / 24
	if d < 0 {
		return 0
	}
	return int(d)
}

// parseBuzzTime parses Buzz's timestamp variants, returning the zero time on
// failure.
func parseBuzzTime(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05.999Z",
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

// --- defensive JSON extraction over Buzz's mixed string/number/nested shape ---

func rawStr(m map[string]json.RawMessage, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return s
	}
	return strings.Trim(string(v), `"`)
}

func rawNum(m map[string]json.RawMessage, key string) float64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	var f float64
	if err := json.Unmarshal(v, &f); err == nil {
		return f
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		f, _ = strconv.ParseFloat(strings.TrimSpace(s), 64)
		return f
	}
	return 0
}

func rawInt(m map[string]json.RawMessage, key string) int { return int(rawNum(m, key)) }
func rawObj(m map[string]json.RawMessage, key string) map[string]json.RawMessage {
	v, ok := m[key]
	if !ok {
		return nil
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(v, &out); err != nil {
		return nil
	}
	return out
}

// parseEnrollRow flattens one raw enrollment object (with optional nested
// metrics/enrollmentmetrics and user nodes) into an enrollRow.
func parseEnrollRow(raw json.RawMessage) enrollRow {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return enrollRow{}
	}
	e := enrollRow{
		EnrollmentID: rawStr(m, "id"),
		UserID:       rawStr(m, "userid"),
		CourseID:     rawStr(m, "courseid"),
		DomainID:     rawStr(m, "domainid"),
		Status:       rawStr(m, "status"),
		Privileges:   rawStr(m, "privileges"),
		RoleID:       rawStr(m, "roleid"),
		FirstActed:   rawStr(m, "firstactivitydate"),
		LastActed:    rawStr(m, "lastactivitydate"),
	}
	if e.EnrollmentID == "" {
		e.EnrollmentID = rawStr(m, "enrollmentid")
	}
	// nested user
	if u := rawObj(m, "user"); u != nil {
		e.FirstName = rawStr(u, "firstname")
		e.LastName = rawStr(u, "lastname")
		e.Email = rawStr(u, "email")
		e.Username = rawStr(u, "username")
	}
	// nested metrics (key may be "metrics" or "enrollmentmetrics")
	met := rawObj(m, "metrics")
	if met == nil {
		met = rawObj(m, "enrollmentmetrics")
	}
	if met != nil {
		e.HasMetrics = true
		e.FinalScore = rawNum(met, "finalscore")
		e.FinalLetter = rawStr(met, "finalletter")
		if e.FinalLetter == "" {
			e.FinalLetter = rawStr(met, "letter")
		}
		e.Completed = rawInt(met, "completed")
		e.Completable = rawInt(met, "completable")
		e.Gradable = rawInt(met, "gradable")
		e.Graded = rawInt(met, "graded")
		e.Possible = rawNum(met, "possible")
		e.UngradedTotal = rawInt(met, "ungradedtotal")
		e.UngradedRed = rawInt(met, "ungradedred")
		e.OldestWork = rawStr(met, "oldestworkitem")
		e.Seconds = int64(rawNum(met, "seconds"))
		e.Late = rawInt(met, "late")
		e.PaceLate = rawInt(met, "pacelate")
		e.PacePast = rawInt(met, "pacepast")
		e.Failing = rawInt(met, "failing")
		e.Failed = rawInt(met, "failed")
	}
	return e
}

// enrollScope identifies what set of enrollments to pull.
type enrollScope struct {
	kind      string // "domain" or "course"
	id        string
	allDesc   bool
	allStatus bool
}

// fetchEnrollments pulls enrollments with metrics+user for a domain or course
// scope, paging by id cursor. It is bounded by maxPages to respect the live
// timeout; each page is up to 1000 rows.
func fetchEnrollments(ctx context.Context, c *client.Client, scope enrollScope, maxPages int) ([]enrollRow, int, error) {
	const pageSize = 1000
	var rows []enrollRow
	scanned := 0
	lastID := ""
	cmd := "listenrollments"
	plural, singular := "enrollments", "enrollment"
	if scope.kind == "course" {
		cmd = "listentityenrollments"
	}
	for page := 0; page < maxPages; page++ {
		params := map[string]string{
			"select": "metrics,user",
			"limit":  strconv.Itoa(pageSize),
		}
		if scope.kind == "course" {
			params["entityid"] = scope.id
		} else {
			params["domainid"] = scope.id
			if scope.allDesc {
				params["includedescendantdomains"] = "true"
			}
		}
		if scope.allStatus {
			params["allstatus"] = "true"
		}
		if lastID != "" {
			params["query"] = "/id>" + lastID
		}
		body, err := c.Get(ctx, "/cmd?cmd="+cmd, params)
		if err != nil {
			return rows, scanned, err
		}
		items, err := client.DLAPList(body, plural, singular)
		if err != nil {
			return rows, scanned, fmt.Errorf("parsing enrollments: %w", err)
		}
		if len(items) == 0 {
			break
		}
		for _, it := range items {
			r := parseEnrollRow(it)
			scanned++
			rows = append(rows, r)
			lastID = r.EnrollmentID
		}
		if len(items) < pageSize {
			break
		}
	}
	return rows, scanned, nil
}

// gradeBucket buckets a graded enrollment into A/B/C/D/F, or "—" when the
// enrollment has no grade yet. A letter is authoritative; otherwise a score is
// only meaningful when the enrollment actually has graded work (graded > 0) —
// otherwise a fresh, never-graded enrollment (finalscore 0, no letter) would be
// miscounted as an F.
func gradeBucket(letter string, score float64, graded int) string {
	l := strings.ToUpper(strings.TrimSpace(letter))
	if l != "" {
		switch l[0] {
		case 'A', 'B', 'C', 'D', 'F':
			return string(l[0])
		}
	}
	if graded <= 0 {
		return "—"
	}
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// gradeHistogram returns an ordered A→F(+—) distribution.
func gradeHistogram(rows []enrollRow) []struct {
	Bucket string `json:"bucket"`
	Count  int    `json:"count"`
} {
	order := []string{"A", "B", "C", "D", "F", "—"}
	counts := map[string]int{}
	for _, r := range rows {
		counts[gradeBucket(r.FinalLetter, r.FinalScore, r.Graded)]++
	}
	out := make([]struct {
		Bucket string `json:"bucket"`
		Count  int    `json:"count"`
	}, 0, len(order))
	for _, b := range order {
		if counts[b] > 0 {
			out = append(out, struct {
				Bucket string `json:"bucket"`
				Count  int    `json:"count"`
			}{b, counts[b]})
		}
	}
	return out
}

func median(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sort.Float64s(xs)
	n := len(xs)
	if n%2 == 1 {
		return xs[n/2]
	}
	return (xs[n/2-1] + xs[n/2]) / 2
}

func round1(f float64) float64 { return math.Round(f*10) / 10 }

func ratio(num, den float64) float64 {
	if den <= 0 {
		return 0
	}
	return round1(num / den)
}

func mean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	var s float64
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}

func nowUTC() time.Time { return time.Now().UTC() }

// resolveScope validates --type plus the matching id flag.
func resolveScope(typ, domainID, entityID string) (enrollScope, error) {
	switch strings.ToLower(strings.TrimSpace(typ)) {
	case "", "domain":
		if domainID == "" {
			return enrollScope{}, fmt.Errorf("--domainid is required for --type domain")
		}
		return enrollScope{kind: "domain", id: domainID, allDesc: true}, nil
	case "course", "section", "entity":
		if entityID == "" {
			return enrollScope{}, fmt.Errorf("--entityid is required for --type course")
		}
		return enrollScope{kind: "course", id: entityID}, nil
	default:
		return enrollScope{}, fmt.Errorf("--type must be domain or course, got %q", typ)
	}
}
