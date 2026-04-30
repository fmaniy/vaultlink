package vault

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeWatchClient struct {
	calls   int
	results []map[string]interface{}
	err     error
}

func (f *fakeWatchClient) ReadSecret(_, _ string) (map[string]interface{}, error) {
	if f.err != nil {
		return nil, f.err
	}
	idx := f.calls
	if idx >= len(f.results) {
		idx = len(f.results) - 1
	}
	f.calls++
	return f.results[idx], nil
}

func TestWatchSecret_EmitsResults(t *testing.T) {
	client := &fakeWatchClient{
		results: []map[string]interface{}{
			{"key": "v1"},
			{"key": "v2"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer cancel()

	ch := WatchSecret(ctx, client, "secret", "app/config", 100*time.Millisecond)

	var results []WatchResult
	for r := range ch {
		results = append(results, r)
	}

	if len(results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(results))
	}
	if results[0].Changed {
		t.Error("first result should not be marked changed")
	}
	if !results[1].Changed {
		t.Error("second result should be marked changed")
	}
}

func TestWatchSecret_PropagatesError(t *testing.T) {
	client := &fakeWatchClient{err: errors.New("vault unavailable")}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	ch := WatchSecret(ctx, client, "secret", "app/config", 80*time.Millisecond)

	r := <-ch
	if r.Err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestSecretChanged_NoPrev(t *testing.T) {
	if secretChanged(nil, map[string]interface{}{"k": "v"}) {
		t.Error("should not be changed when prev is nil")
	}
}

func TestSecretChanged_Mismatch(t *testing.T) {
	prev := map[string]interface{}{"k": "old"}
	curr := map[string]interface{}{"k": "new"}
	if !secretChanged(prev, curr) {
		t.Error("expected changed=true")
	}
}

func TestSecretChanged_Equal(t *testing.T) {
	prev := map[string]interface{}{"k": "same"}
	curr := map[string]interface{}{"k": "same"}
	if secretChanged(prev, curr) {
		t.Error("expected changed=false")
	}
}
