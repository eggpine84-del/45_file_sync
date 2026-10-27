package shell_test

import (
	"file_sync/internal/engine/shell"
	"os"
	"path/filepath"
	"testing"
)

func TestE2E_VirtualSync(t *testing.T) {
	tempRoot := filepath.Join(os.TempDir(), "filesync_e2e_test")
	_ = os.RemoveAll(tempRoot)
	defer os.RemoveAll(tempRoot)

	// 1. 소스 디렉터리 및 단일 파일 생성
	srcDir := filepath.Join(tempRoot, "source_folder")
	subDir := filepath.Join(srcDir, "sub_module")
	_ = os.MkdirAll(subDir, 0755)

	file1 := filepath.Join(srcDir, "file1.txt")
	_ = os.WriteFile(file1, []byte("Content of file 1"), 0644)

	file2 := filepath.Join(subDir, "nested.go")
	_ = os.WriteFile(file2, []byte("package sub\nvar Value = 42"), 0644)

	standaloneFile := filepath.Join(tempRoot, "standalone.md")
	_ = os.WriteFile(standaloneFile, []byte("# Standalone File\nZero Emoji"), 0644)

	// 2. 대상 디렉터리 2개 생성
	dst1 := filepath.Join(tempRoot, "target_project_A")
	dst2 := filepath.Join(tempRoot, "target_project_B")
	_ = os.MkdirAll(dst1, 0755)
	_ = os.MkdirAll(dst2, 0755)

	preset := shell.Preset{
		ID:      "e2e-preset",
		Name:    "Virtual Test Preset",
		Sources: []string{srcDir, standaloneFile},
		Targets: []string{dst1, dst2},
	}

	service := shell.NewSyncService(nil)

	// 3. 1차 전체 동기화 실행
	var progressLogs []string
	result1, err := service.ExecutePreset(preset, func(cur, tot int, msg string) {
		progressLogs = append(progressLogs, msg)
	})

	if err != nil {
		t.Fatalf("1st sync failed: %v", err)
	}

	// 각 대상 폴더당 파일 3개(file1, nested, standalone) * 2개 대상 = 총 6개 복사되어야 함
	expectedCopies := 6
	if result1.TotalCopied != expectedCopies {
		t.Fatalf("expected %d copies in 1st sync, got %d", expectedCopies, result1.TotalCopied)
	}

	// 실제 대상 폴더 파일 무결성 대조
	checkFiles := []struct {
		targetDir string
		relPath   string
		expected  string
	}{
		{dst1, "file1.txt", "Content of file 1"},
		{dst1, filepath.Join("sub_module", "nested.go"), "package sub\nvar Value = 42"},
		{dst1, "standalone.md", "# Standalone File\nZero Emoji"},
		{dst2, "file1.txt", "Content of file 1"},
		{dst2, filepath.Join("sub_module", "nested.go"), "package sub\nvar Value = 42"},
		{dst2, "standalone.md", "# Standalone File\nZero Emoji"},
	}

	for _, cf := range checkFiles {
		fullPath := filepath.Join(cf.targetDir, cf.relPath)
		data, readErr := os.ReadFile(fullPath)
		if readErr != nil {
			t.Fatalf("file not found at target: %s (%v)", fullPath, readErr)
		}
		if string(data) != cf.expected {
			t.Fatalf("content mismatch for %s: expected '%s', got '%s'", fullPath, cf.expected, string(data))
		}
	}

	// 4. 2차 동기화 (내용 불변) -> 6개 모두 SKIP 되어야 함
	result2, err2 := service.ExecutePreset(preset, nil)
	if err2 != nil {
		t.Fatalf("2nd sync failed: %v", err2)
	}
	if result2.TotalCopied != 0 || result2.TotalSkipped != expectedCopies {
		t.Fatalf("expected 0 copied and %d skipped, got copied=%d, skipped=%d",
			expectedCopies, result2.TotalCopied, result2.TotalSkipped)
	}

	// 5. 3차 동기화: standalone.md만 내용 수정 -> standalone.md만 복사되고 나머지는 SKIP 되어야 함
	_ = os.WriteFile(standaloneFile, []byte("# Updated Standalone\nVersion 2.0"), 0644)
	result3, err3 := service.ExecutePreset(preset, nil)
	if err3 != nil {
		t.Fatalf("3rd sync failed: %v", err3)
	}

	// standalone.md가 대상 2곳에 각각 1개씩 복사(총 2개)되고, 나머지 4개는 SKIP
	if result3.TotalCopied != 2 || result3.TotalSkipped != 4 {
		t.Fatalf("expected 2 copied and 4 skipped on modification, got copied=%d, skipped=%d",
			result3.TotalCopied, result3.TotalSkipped)
	}

	// 수정된 내용이 대상에도 정상 반영되었는지 확인
	for _, targetDir := range []string{dst1, dst2} {
		updatedPath := filepath.Join(targetDir, "standalone.md")
		data, _ := os.ReadFile(updatedPath)
		if string(data) != "# Updated Standalone\nVersion 2.0" {
			t.Fatalf("updated content not reflected in %s", updatedPath)
		}
	}
}
