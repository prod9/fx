package files

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPlanCleanup(t *testing.T) {
	now := time.Date(2026, 7, 9, 12, 0, 0, 0, time.UTC)
	old := now.Add(-48 * time.Hour) // past the dead timeout
	fresh := now.Add(-1 * time.Minute)
	dead := 24 * time.Hour

	cases := []struct {
		name       string
		rows       []cleanupRow
		storeKeys  []string
		wantDelete []string
		wantPrune  []int64
	}{
		{
			name:      "matched row and object — no action",
			rows:      []cleanupRow{{ID: 1, RemotePath: "drop/1/1", CreatedAt: old}},
			storeKeys: []string{"drop/1/1"},
		},
		{
			name:      "no records — no window, nothing reconciled",
			rows:      nil,
			storeKeys: []string{"drop/9/9"},
		},
		{
			name:      "young row, object not yet uploaded — in flight, left alone",
			rows:      []cleanupRow{{ID: 2, RemotePath: "drop/2/2", CreatedAt: fresh}},
			storeKeys: nil,
		},
		{
			name:      "old row, object never arrived — abandoned, pruned",
			rows:      []cleanupRow{{ID: 3, RemotePath: "drop/3/3", CreatedAt: old}},
			storeKeys: nil,
			wantPrune: []int64{3},
		},
		{
			name:       "in-window orphan — object above floor with no row, deleted",
			rows:       []cleanupRow{{ID: 10, RemotePath: "drop/o/10", CreatedAt: old}},
			storeKeys:  []string{"drop/o/10", "drop/o/11"},
			wantDelete: []string{"drop/o/11"},
		},
		{
			name:      "superseded orphan below floor — settled by earlier sweeps, skipped",
			rows:      []cleanupRow{{ID: 5, RemotePath: "drop/s/5", CreatedAt: old}},
			storeKeys: []string{"drop/s/2", "drop/s/5"},
		},
		{
			name:      "unparseable key — ignored",
			rows:      []cleanupRow{{ID: 4, RemotePath: "drop/4/4", CreatedAt: old}},
			storeKeys: []string{"drop/4/4", "weird/key/abc"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			plan := planCleanup(c.rows, c.storeKeys, now, dead)
			require.ElementsMatch(t, c.wantDelete, plan.ObjectsToDelete)
			require.ElementsMatch(t, c.wantPrune, plan.RowsToPrune)
		})
	}
}
