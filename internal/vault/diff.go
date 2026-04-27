package vault

// DiffResult represents the comparison of a secret key between two environments.
type DiffResult struct {
	Path    string
	Key     string
	Status  DiffStatus
	SrcVal  string
	DstVal  string
}

// DiffStatus describes the outcome of comparing a single secret key.
type DiffStatus string

const (
	DiffStatusMatch   DiffStatus = "match"
	DiffStatusMissing DiffStatus = "missing"
	DiffStatusExtra   DiffStatus = "extra"
	DiffStatusMismatch DiffStatus = "mismatch"
)

// DiffSecrets compares two secret maps and returns a slice of DiffResult entries.
// srcSecrets is the source of truth; dstSecrets is the target environment.
func DiffSecrets(path string, srcSecrets, dstSecrets map[string]interface{}) []DiffResult {
	results := make([]DiffResult, 0)

	for k, sv := range srcSecrets {
		srcStr := toString(sv)
		if dv, ok := dstSecrets[k]; ok {
			dstStr := toString(dv)
			if srcStr == dstStr {
				results = append(results, DiffResult{Path: path, Key: k, Status: DiffStatusMatch, SrcVal: srcStr, DstVal: dstStr})
			} else {
				results = append(results, DiffResult{Path: path, Key: k, Status: DiffStatusMismatch, SrcVal: srcStr, DstVal: dstStr})
			}
		} else {
			results = append(results, DiffResult{Path: path, Key: k, Status: DiffStatusMissing, SrcVal: srcStr, DstVal: ""})
		}
	}

	for k, dv := range dstSecrets {
		if _, ok := srcSecrets[k]; !ok {
			results = append(results, DiffResult{Path: path, Key: k, Status: DiffStatusExtra, SrcVal: "", DstVal: toString(dv)})
		}
	}

	return results
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
