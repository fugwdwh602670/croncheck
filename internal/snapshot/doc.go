// Package snapshot provides point-in-time capture of all monitored job states.
//
// A snapshot records the health, missed count, and last-seen time for every
// known job at the moment of capture. Snapshots are stored in a bounded
// in-memory ring buffer and can be retrieved by ID or listed newest-first.
//
// # Ring Buffer Behaviour
//
// When the buffer is full, the oldest snapshot is silently evicted to make
// room for the newest one. The capacity is set once at construction time via
// [New] and cannot be changed afterwards.
//
// # Concurrency
//
// All exported methods on Store are safe for concurrent use by multiple
// goroutines.
//
// # Typical usage
//
//	store := snapshot.New(20)
//	snap := store.Capture("manual-2024-06-01", jobSource)
//
//	// Retrieve a specific snapshot by its ID.
//	if s, ok := store.Get("manual-2024-06-01"); ok {
//		fmt.Println(s.CapturedAt)
//	}
//
//	// List all retained snapshots, newest first.
//	for _, s := range store.List() {
//		fmt.Println(s.ID, s.CapturedAt)
//	}
package snapshot
