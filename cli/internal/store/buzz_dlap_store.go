// Hand-authored: DLAP read responses wrap the record in
// {"response":{"code":"OK","<entity>":{...,"id":...}}}. Incidental read-caching
// passes that whole envelope as the item, so the generic ID extractor finds no
// top-level id and skips caching. unwrapDLAPRecord descends to the inner record
// when the item is that shape; it is a no-op for already-flat sync rows.
package store

func unwrapDLAPRecord(obj map[string]any) map[string]any {
	resp, ok := obj["response"].(map[string]any)
	if !ok {
		return obj
	}
	for k, v := range resp {
		switch k {
		case "code", "message", "errorId":
			continue
		}
		if inner, ok := v.(map[string]any); ok {
			if _, has := inner["id"]; has {
				return inner
			}
		}
	}
	return obj
}
