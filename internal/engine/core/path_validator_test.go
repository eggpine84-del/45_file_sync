package core_test

import (
	"file_sync/internal/engine/core"
	"testing"
)

func TestPathValidator_Success(t *testing.T) {
	src := "C:/Projects/source"
	dst := "D:/Backup/target"

	err := core.ValidatePathPair(src, dst)
	if err != nil {
		t.Fatalf("expected valid path pair, got error: %v", err)
	}
}

func TestPathValidator_Empty(t *testing.T) {
	err := core.ValidatePathPair("", "D:/target")
	if err == nil {
		t.Fatal("expected error for empty source path, got nil")
	}

	err = core.ValidatePathPair("C:/source", "")
	if err == nil {
		t.Fatal("expected error for empty target path, got nil")
	}
}

func TestPathValidator_Identical(t *testing.T) {
	path := "C:/Projects/source"
	err := core.ValidatePathPair(path, path)
	if err == nil {
		t.Fatal("expected error for identical paths, got nil")
	}
}

func TestPathValidator_Nesting(t *testing.T) {
	parent := "C:/Projects/workspace"
	child := "C:/Projects/workspace/nested/target"

	// Target in Source
	err := core.ValidatePathPair(parent, child)
	if err == nil {
		t.Fatal("expected error when target is inside source, got nil")
	}

	// Source in Target
	err = core.ValidatePathPair(child, parent)
	if err == nil {
		t.Fatal("expected error when source is inside target, got nil")
	}
}

func TestPathValidator_Batch(t *testing.T) {
	sources := []string{"C:/Src1", "C:/Src2"}
	targets := []string{"D:/Dst1", "D:/Dst2"}

	err := core.ValidateBatch(sources, targets)
	if err != nil {
		t.Fatalf("expected batch validation to pass, got: %v", err)
	}

	// Empty sources
	err = core.ValidateBatch([]string{}, targets)
	if err == nil {
		t.Fatal("expected error for empty sources batch, got nil")
	}

	// Nesting collision in batch
	collisionTargets := []string{"D:/Dst1", "C:/Src1/backup"}
	err = core.ValidateBatch(sources, collisionTargets)
	if err == nil {
		t.Fatal("expected error for nested path collision in batch, got nil")
	}
}
