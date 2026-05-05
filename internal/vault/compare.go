package vault

import "fmt"

// CompareResult holds the outcome of comparing a single secret path across two environments.
type CompareResult struct {
	Path    string
	Status  string // "match", "mismatch", "missing_src", "missing_dst"
	Details string
}

// CompareSecrets reads all secrets under the given paths from src and dst clients
// and returns a per-path comparison result.
func CompareSecrets(src, dst KVClient, srcMount, dstMount, srcPath, dstPath string) ([]CompareResult, error) {
	srcPaths, err := ListSecrets(src, srcMount, srcPath)
	if err != nil {
		return nil, fmt.Errorf("list src: %w", err)
	}

	dstPaths, err := ListSecrets(dst, dstMount, dstPath)
	if err != nil {
		return nil, fmt.Errorf("list dst: %w", err)
	}

	dstSet := make(map[string]struct{}, len(dstPaths))
	for _, p := range dstPaths {
		dstSet[p] = struct{}{}
	}

	srcSet := make(map[string]struct{}, len(srcPaths))
	for _, p := range srcPaths {
		srcSet[p] = struct{}{}
	}

	var results []CompareResult

	for _, p := range srcPaths {
		if _, ok := dstSet[p]; !ok {
			results = append(results, CompareResult{Path: p, Status: "missing_dst"})
			continue
		}
		srcData, err := ReadSecret(src, srcMount, p)
		if err != nil {
			return nil, fmt.Errorf("read src %s: %w", p, err)
		}
		dstData, err := ReadSecret(dst, dstMount, p)
		if err != nil {
			return nil, fmt.Errorf("read dst %s: %w", p, err)
		}
		if diff := diffMaps(srcData, dstData); diff != "" {
			results = append(results, CompareResult{Path: p, Status: "mismatch", Details: diff})
		} else {
			results = append(results, CompareResult{Path: p, Status: "match"})
		}
	}

	for _, p := range dstPaths {
		if _, ok := srcSet[p]; !ok {
			results = append(results, CompareResult{Path: p, Status: "missing_src"})
		}
	}

	return results, nil
}

func diffMaps(a, b map[string]interface{}) string {
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return fmt.Sprintf("key %q missing in dst", k)
		}
		if fmt.Sprintf("%v", av) != fmt.Sprintf("%v", bv) {
			return fmt.Sprintf("key %q differs", k)
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			return fmt.Sprintf("key %q missing in src", k)
		}
	}
	return ""
}
