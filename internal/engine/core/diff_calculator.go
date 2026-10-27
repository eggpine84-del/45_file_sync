package core

const (
	ActionCopy = "COPY"
	ActionSkip = "SKIP"
)

// FileMeta represents metadata of a single file for pure comparison.
type FileMeta struct {
	RelPath      string
	Size         int64
	ModTimeUnix  int64
	Hash         string
	IsDirectory  bool
}

// SyncTask defines a deterministic sync operation.
type SyncTask struct {
	SourcePath string
	TargetPath string
	RelPath    string
	Action     string
	Size       int64
}

// NeedsSync determines if a source file should be copied to target.
func NeedsSync(src, dst FileMeta) bool {
	if src.Size != dst.Size {
		return true
	}
	if src.ModTimeUnix > dst.ModTimeUnix {
		return true
	}
	if src.Hash != "" && dst.Hash != "" && src.Hash != dst.Hash {
		return true
	}
	return false
}

// ComputeSyncPlan generates a list of sync actions by comparing source and destination file maps.
func ComputeSyncPlan(srcMap map[string]FileMeta, dstMap map[string]FileMeta, srcRoot, dstRoot string) []SyncTask {
	tasks := make([]SyncTask, 0, len(srcMap))

	for relPath, srcMeta := range srcMap {
		if srcMeta.IsDirectory {
			continue
		}

		dstMeta, exists := dstMap[relPath]
		action := ActionCopy

		if exists && !NeedsSync(srcMeta, dstMeta) {
			action = ActionSkip
		}

		tasks = append(tasks, SyncTask{
			SourcePath: NormalizePath(srcRoot + PathSeparatorSlash + relPath),
			TargetPath: NormalizePath(dstRoot + PathSeparatorSlash + relPath),
			RelPath:    relPath,
			Action:     action,
			Size:       srcMeta.Size,
		})
	}

	return tasks
}
