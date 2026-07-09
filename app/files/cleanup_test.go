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
			name:       "object with no row — orphan deleted",
			rows:       nil,
			storeKeys:  []string{"drop/9/9"},
			wantDelete: []string{"drop/9/9"},
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
			name: "superseded single-file: old object orphaned, newest kept",
			rows: []cleanupRow{{ID: 5, RemotePath: "drop/5/2", CreatedAt: old}},
			// two objects on disk, only the newest row survives
			storeKeys:  []string{"drop/5/1", "drop/5/2"},
			wantDelete: []string{"drop/5/1"},
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
