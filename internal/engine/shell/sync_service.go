package shell

import (
	"errors"
	"file_sync/internal/engine/core"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	ErrSyncAlreadyRunning = errors.New("sync is already running")
)

// ProgressCallback receives real-time progress updates during sync.
type ProgressCallback func(current, total int, message string)

// SyncResult summarizes the outcome of a batch sync operation.
type SyncResult struct {
	TotalCopied  int      `json:"total_copied"`
	TotalSkipped int      `json:"total_skipped"`
	TotalErrors  int      `json:"total_errors"`
	Logs         []string `json:"logs"`
}

// SyncService orchestrates batch file synchronization with concurrency locks.
type SyncService struct {
	adapter *FSAdapter
	mu      sync.Mutex
	running bool
}

// NewSyncService creates an orchestrated sync service instance.
func NewSyncService(adapter *FSAdapter) *SyncService {
	if adapter == nil {
		adapter = NewFSAdapter()
	}
	return &SyncService{
		adapter: adapter,
	}
}

// IsRunning reports whether a synchronization is currently in progress.
func (s *SyncService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// ExecutePreset executes synchronization for all source-target pairs defined in the preset.
func (s *SyncService) ExecutePreset(preset Preset, callback ProgressCallback) (*SyncResult, error) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil, ErrSyncAlreadyRunning
	}
	s.running = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	if err := core.ValidateBatch(preset.Sources, preset.Targets); err != nil {
		return nil, err
	}

	result := &SyncResult{
		Logs: make([]string, 0),
	}

	var allTasks []core.SyncTask

	for _, src := range preset.Sources {
		srcInfo, statErr := os.Stat(core.NormalizePath(src))
		if statErr != nil {
			msg := fmt.Sprintf("Source not found: %s", src)
			result.Logs = append(result.Logs, msg)
			result.TotalErrors++
			continue
		}

		srcMap, scanErr := s.adapter.ScanPath(src)
		if scanErr != nil {
			result.Logs = append(result.Logs, scanErr.Error())
			result.TotalErrors++
			continue
		}

		for _, dst := range preset.Targets {
			dstMap, _ := s.adapter.ScanPath(dst)
			if dstMap == nil {
				dstMap = make(map[string]core.FileMeta)
			}

			var tasks []core.SyncTask
			if srcInfo.IsDir() {
				tasks = core.ComputeSyncPlan(srcMap, dstMap, src, dst)
			} else {
				// Single file source
				fileName := filepath.Base(core.NormalizePath(src))
				dstMapKey := fileName
				dstMeta, exists := dstMap[dstMapKey]
				srcMeta := srcMap[fileName]

				action := core.ActionCopy
				if exists && !core.NeedsSync(srcMeta, dstMeta) {
					action = core.ActionSkip
				}

				tasks = append(tasks, core.SyncTask{
					SourcePath: core.NormalizePath(src),
					TargetPath: core.NormalizePath(filepath.Join(dst, fileName)),
					RelPath:    fileName,
					Action:     action,
					Size:       srcMeta.Size,
				})
			}
			allTasks = append(allTasks, tasks...)
		}
	}

	total := len(allTasks)
	for i, task := range allTasks {
		current := i + 1

		if task.Action == core.ActionSkip {
			result.TotalSkipped++
			if callback != nil {
				callback(current, total, "")
			}
			continue
		}

		copyErr := s.adapter.CopyFileAtomic(task.SourcePath, task.TargetPath)
		if copyErr != nil {
			result.TotalErrors++
			msg := fmt.Sprintf("[ERROR] %s: %v", task.RelPath, copyErr)
			result.Logs = append(result.Logs, msg)
			if callback != nil {
				callback(current, total, msg)
			}
		} else {
			result.TotalCopied++
			msg := fmt.Sprintf("[COPIED] %s", task.RelPath)
			result.Logs = append(result.Logs, msg)
			if callback != nil {
				callback(current, total, msg)
			}
		}
	}

	return result, nil
}
