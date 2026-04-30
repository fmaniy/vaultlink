package vault

import (
	"context"
	"fmt"
	"time"
)

// WatchResult holds a snapshot of secrets at a point in time.
type WatchResult struct {
	Path      string
	Data      map[string]interface{}
	Timestamp time.Time
	Changed   bool
	Err       error
}

// SecretReader is the minimal interface needed to watch a secret.
type SecretReader interface {
	ReadSecret(mount, path string) (map[string]interface{}, error)
}

// WatchSecret polls a secret at the given interval and emits WatchResult
// values on the returned channel until ctx is cancelled.
func WatchSecret(ctx context.Context, client SecretReader, mount, path string, interval time.Duration) <-chan WatchResult {
	ch := make(chan WatchResult)

	go func() {
		defer close(ch)

		var prev map[string]interface{}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				data, err := client.ReadSecret(mount, path)
				result := WatchResult{
					Path:      fmt.Sprintf("%s/%s", mount, path),
					Timestamp: t,
					Err:       err,
				}
				if err == nil {
					result.Data = data
					result.Changed = secretChanged(prev, data)
					prev = data
				}
				ch <- result
			}
		}
	}()

	return ch
}

// secretChanged returns true if the two secret maps differ.
func secretChanged(prev, curr map[string]interface{}) bool {
	if prev == nil {
		return false
	}
	if len(prev) != len(curr) {
		return true
	}
	for k, v := range prev {
		if curr[k] != v {
			return true
		}
	}
	return false
}
