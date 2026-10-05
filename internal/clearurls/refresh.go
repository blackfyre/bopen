package clearurls

import (
	"context"
	"time"
)

// Updater fetches the list into the cache; Update satisfies it.
type Updater func(ctx context.Context, c Cache, f Fetcher, now time.Time) (Meta, error)

// StartRefresh starts a background update when ClearURLs is enabled and the
// cache described by meta is missing or stale at now. It returns whether an update started;
// done, when not nil, runs after it finishes. Nothing touches the network
// when enabled is false.
func StartRefresh(enabled bool, meta Meta, c Cache, f Fetcher, now time.Time, update Updater, done func(Meta, error)) bool {
	if !enabled || !meta.NeedsRefresh(now) {
		return false
	}
	go func() {
		m, err := update(context.Background(), c, f, now)
		if done != nil {
			done(m, err)
		}
	}()
	return true
}
