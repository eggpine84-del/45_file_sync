package core_test

import (
	"file_sync/internal/engine/core"
	"testing"
)

func TestNeedsSync(t *testing.T) {
	base := core.FileMeta{
		RelPath:     "file.txt",
		Size:        100,
		ModTimeUnix: 1000,
		Hash:        "hash1",
	}

	// Identical -> Skip
	if core.NeedsSync(base, base) {
		t.Fatal("expected identical files not to need sync")
	}

	// Size difference -> Copy
	diffSize := base
	diffSize.Size = 200
	if !core.NeedsSync(diffSize, base) {
		t.Fatal("expected different size to need sync")
	}

	// Newer source -> Copy
	newerSrc := base
	newerSrc.ModTimeUnix = 2000
	if !core.NeedsSync(newerSrc, base) {
		t.Fatal("expected newer source to need sync")
	}

	// Different hash -> Copy
	diffHash := base
	diffHash.Hash = "hash2"
	if !core.NeedsSync(diffHash, base) {
		t.Fatal("expected different hash to need sync")
	}
}

func TestComputeSyncPlan(t *testing.T) {
	srcMap := map[string]core.FileMeta{
		"new.txt":   {RelPath: "new.txt", Size: 50, ModTimeUnix: 1000},
		"same.txt":  {RelPath: "same.txt", Size: 100, ModTimeUnix: 1000},
		"sub/a.txt": {RelPath: "sub/a.txt", Size: 20, ModTimeUnix: 1000},
	}

	dstMap := map[string]core.FileMeta{
		"same.txt": {RelPath: "same.txt", Size: 100, ModTimeUnix: 1000},
	}

	plan := core.ComputeSyncPlan(srcMap, dstMap, "C:/source", "D:/target")
	if len(plan) != 3 {
		t.Fatalf("expected 3 tasks in sync plan, got %d", len(plan))
	}

	copyCount := 0
	skipCount := 0
	for _, task := range plan {
		if task.Action == core.ActionCopy {
			copyCount++
		} else if task.Action == core.ActionSkip {
			skipCount++
		}
	}

	if copyCount != 2 {
		t.Fatalf("expected 2 COPY tasks, got %d", copyCount)
	}
	if skipCount != 1 {
		t.Fatalf("expected 1 SKIP task, got %d", skipCount)
	}
}
