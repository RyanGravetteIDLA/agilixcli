// Generated from api.agilixbuzz.com/docs/all.md request-body shapes (hand-maintained).
// Maps DLAP commands to the request-envelope they require so the client can wrap
// flat flag-built bodies into the form the API accepts. Hand-authored support file.
package client

// dlapRequestEntity maps a DLAP command (lowercase) to the entity key used in the
// batch body envelope {"requests":{"<entity>":[ ... ]}}.
var dlapRequestEntity = map[string]string{
	"addgroupmembers":             "member",
	"calculateenrollmentscenario": "scenario",
	"copycourses":                 "course",
	"copyitems":                   "item",
	"copyresources":               "resource",
	"copywikipages":               "wikipage",
	"createcommandtokens":         "commandtoken",
	"createcourses":               "course",
	"createdomains":               "domain",
	"createenrollments":           "enrollment",
	"creategroups":                "group",
	"createobjectivesets":         "set",
	"createusers":                 "user",
	"createusers2":                "user",
	"deleteannouncements":         "announcement",
	"deleteblogs":                 "message",
	"deletecommandtokens":         "commandtoken",
	"deletegroups":                "group",
	"deleteitems":                 "item",
	"deleteobjectivemaps":         "map",
	"deleteobjectives":            "objective",
	"deleteobjectivesets":         "set",
	"deleteresources":             "resource",
	"deletesubscriptions":         "subscription",
	"deleteusers":                 "user",
	"deletewikipages":             "wikipage",
	"getdocumentinfo":             "document",
	"getiteminfo":                 "item",
	"getmanifestinfo":             "manifest",
	"getpeerresponseinfo":         "peerresponse",
	"getresourceinfo2":            "resource",
	"getstudentsubmissioninfo":    "submission",
	"getteacherresponseinfo":      "teacherresponse",
	"mergecourses":                "course",
	"putitemactivity":             "activity",
	"putitems":                    "item",
	"putitemstatus":               "status",
	"putobjectivemaps":            "map",
	"putobjectives":               "objective",
	"putquestions":                "question",
	"putresourcefolders":          "folder",
	"putteacherresponses":         "teacherresponse",
	"removegroupmembers":          "member",
	"restoreannouncements":        "announcement",
	"restoredocuments":            "document",
	"restoreitems":                "item",
	"restoremessages":             "message",
	"restorequestions":            "question",
	"restoreresources":            "resource",
	"restorewikipages":            "wikipage",
	"updateannouncementviewed":    "announcement",
	"updateblogviewed":            "message",
	"updatecommandtokens":         "commandtoken",
	"updatecourses":               "course",
	"updatedomains":               "domain",
	"updateenrollments":           "enrollment",
	"updategroups":                "group",
	"updatemanifestdata":          "manifest",
	"updatemessageviewed":         "message",
	"updateobjectivesets":         "set",
	"updaterights":                "rights",
	"updatesubscriptions":         "subscription",
	"updateusers":                 "user",
	"updatewikipageviewed":        "wikipage",
}

// dlapSingleRequest is the set of DLAP commands whose body is the single-wrap
// form {"request":{ ... }} rather than the batch {"requests":{...}} form.
var dlapSingleRequest = map[string]bool{
	"clearsecondfactorauthentication": true,
	"createbadge":                     true,
	"createdemocourse":                true,
	"createrole":                      true,
	"exportdata":                      true,
	"extendsession":                   true,
	"finishpasswordreset":             true,
	"forcepasswordchange":             true,
	"generatesubmission":              true,
	"getnextquestion":                 true,
	"getquestionscores":               true,
	"login2":                          true,
	"login3":                          true,
	"logout":                          true,
	"proxy":                           true,
	"putscodata":                      true,
	"resetlockout":                    true,
	"resetpassword":                   true,
	"saveattemptanswers":              true,
	"secondfactorauthenticate":        true,
	"setdatastreamconfiguration":      true,
	"setupsecondfactorauthentication": true,
	"submitattemptanswers":            true,
	"unproxy":                         true,
	"updatepassword":                  true,
	"updatepasswordquestionanswer":    true,
	"updaterole":                      true,
}

// dlapCmdFromPath extracts the lowercased DLAP command from a "/cmd?cmd=X" path.
func dlapCmdFromPath(path string) string {
	i := indexOf(path, "cmd=")
	if i < 0 {
		return ""
	}
	rest := path[i+4:]
	// stop at the next & or end
	for j := 0; j < len(rest); j++ {
		if rest[j] == '&' {
			rest = rest[:j]
			break
		}
	}
	return toLowerASCII(rest)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func toLowerASCII(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}

// wrapDLAPRequestBody wraps a flat, flag-built request body into the envelope
// the DLAP command requires: {"requests":{"<entity>":[ <flat> ]}} for batch
// commands, or {"request":{ <flat> }} for single-request commands. Bodies that
// are already wrapped (carry a "request", "requests", or "settings" key), are
// not maps, or belong to query-param-only commands pass through unchanged. This
// makes the generated flag path work for create/update/put commands without
// requiring the caller to hand-build the envelope via --stdin.
func wrapDLAPRequestBody(path string, body any) any {
	m, ok := body.(map[string]any)
	if !ok {
		return body
	}
	if _, has := m["requests"]; has {
		return body
	}
	if _, has := m["request"]; has {
		return body
	}
	if _, has := m["settings"]; has {
		return body
	}
	cmd := dlapCmdFromPath(path)
	if cmd == "" {
		return body
	}
	normalizeDLAPStatus(m)
	if ent, ok := dlapRequestEntity[cmd]; ok {
		return map[string]any{"requests": map[string]any{ent: []any{m}}}
	}
	if dlapSingleRequest[cmd] {
		return map[string]any{"request": m}
	}
	return body
}

// normalizeDLAPStatus lets callers pass a named EnrollmentStatus
// (none/active/inactive) for the "status" field instead of the numeric enum
// (0/1/10) DLAP requires. Numeric values and unrecognized strings pass through.
func normalizeDLAPStatus(m map[string]any) {
	s, ok := m["status"].(string)
	if !ok || s == "" {
		return
	}
	switch toLowerASCII(s) {
	case "none":
		m["status"] = "0"
	case "active":
		m["status"] = "1"
	case "inactive", "suspended":
		m["status"] = "10"
	}
}
