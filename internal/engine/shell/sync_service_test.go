package shell_test

import (
	"file_sync/internal/engine/shell"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSyncService_Execute(t *testing.T) {
	tempDir := t.TempDir()

	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "target")

	_ = os.MkdirAll(srcDir, 0755)
	_ = os.MkdirAll(dstDir, 0755)

	file1 := filepath.Join(srcDir, "doc.txt")
	_ = os.WriteFile(file1, []byte("Hello World"), 0644)

	preset := shell.Preset{
		ID:      "test-preset",
		Name:    "Test",
		Sources: []string{srcDir},
		Targets: []string{dstDir},
	}

	service := shell.NewSyncService(nil)
	var logs []string
	result, err := service.ExecutePreset(preset, func(cur, tot int, msg string) {
		logs = append(logs, msg)
	})

	if err != nil {
		t.Fatalf("sync execution failed: %v", err)
	}

	if result.TotalCopied != 1 {
		t.Fatalf("expected 1 file copied, got %d", result.TotalCopied)
	}

	expectedLog := "[COPIED] doc.txt"
	if len(logs) == 0 || logs[0] != expectedLog {
		t.Fatalf("expected log %q, got %q", expectedLog, logs[0])
	}

	// Second run -> Should skip without polluting logs
	var secondLogs []string
	result2, err2 := service.ExecutePreset(preset, func(cur, tot int, msg string) {
		if msg != "" {
			secondLogs = append(secondLogs, msg)
		}
	})
	if err2 != nil {
		t.Fatalf("second sync failed: %v", err2)
	}

	if result2.TotalSkipped != 1 {
		t.Fatalf("expected 1 file skipped on second run, got %d", result2.TotalSkipped)
	}
	if len(secondLogs) != 0 {
		t.Fatalf("expected 0 skip log messages, got %d: %v", len(secondLogs), secondLogs)
	}
}

func TestSyncService_MutexLock(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	dstDir := filepath.Join(tempDir, "target")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.MkdirAll(dstDir, 0755)

	_ = os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("data"), 0644)

	preset := shell.Preset{
		ID:      "test",
		Sources: []string{srcDir},
		Targets: []string{dstDir},
	}

	service := shell.NewSyncService(nil)

	var wg sync.WaitGroup
	var errAlreadyRunning error

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = service.ExecutePreset(preset, func(cur, tot int, msg string) {
			time.Sleep(50 * time.Millisecond)
		})
	}()

	time.Sleep(10 * time.Millisecond)

	_, errSecond := service.ExecutePreset(preset, nil)
	if errSecond != nil {
		errAlreadyRunning = errSecond
	}

	wg.Wait()

	if errAlreadyRunning != shell.ErrSyncAlreadyRunning {
		t.Fatalf("expected ErrSyncAlreadyRunning on concurrent execution, got: %v", errAlreadyRunning)
	}
}

func TestWatchDaemon_DynamicSubdirectory(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "source")
	_ = os.MkdirAll(srcDir, 0755)

	preset := shell.Preset{
		ID:      "watch-test",
		Sources: []string{srcDir},
		Targets: []string{filepath.Join(tempDir, "target")},
	}

	triggered := make(chan bool, 5)
	daemon := shell.NewWatchDaemon()
	err := daemon.Start(preset, func() {
		triggered <- true
	})
	if err != nil {
		t.Fatalf("failed to start daemon: %v", err)
	}
	defer daemon.Stop()

	// Wait 50ms for initial watch setup
	time.Sleep(50 * time.Millisecond)

	// Create a new subfolder after watch has started
	subDir := filepath.Join(srcDir, "new_subfolder")
	_ = os.MkdirAll(subDir, 0755)

	// Wait 100ms for dynamic registration
	time.Sleep(100 * time.Millisecond)

	// Drain any creation triggers from the subfolder itself
	for len(triggered) > 0 {
		<-triggered
	}

	// Create a new file inside the newly created subfolder
	newFile := filepath.Join(subDir, "inside_sub.txt")
	_ = os.WriteFile(newFile, []byte("sub-data"), 0644)

	// Verify that trigger is fired for the file inside dynamic subfolder
	select {
	case <-triggered:
		// Success!
	case <-time.After(800 * time.Millisecond):
		t.Fatalf("timeout: watch daemon failed to detect change in dynamically created subfolder")
	}
}

