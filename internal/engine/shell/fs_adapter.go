package shell

import (
	"errors"
	"file_sync/internal/engine/core"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	TempFileSuffix       = ".tmp_sync"
	DefaultDirPermission = 0755
)

// FSAdapter handles actual filesystem operations including scanning and atomic copy.
type FSAdapter struct{}

// NewFSAdapter creates a new filesystem adapter.
func NewFSAdapter() *FSAdapter {
	return &FSAdapter{}
}

// ScanPath scans a file or directory and returns a map of relative paths to FileMeta.
func (a *FSAdapter) ScanPath(root string) (map[string]core.FileMeta, error) {
	normRoot := core.NormalizePath(root)
	info, err := os.Stat(normRoot)
	if err != nil {
		return nil, err
	}

	result := make(map[string]core.FileMeta)

	if !info.IsDir() {
		fileName := filepath.Base(normRoot)
		result[fileName] = core.FileMeta{
			RelPath:     fileName,
			Size:        info.Size(),
			ModTimeUnix: info.ModTime().Unix(),
			IsDirectory: false,
		}
		return result, nil
	}

	err = filepath.WalkDir(normRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if d.IsDir() {
			return nil
		}

		rel, relErr := filepath.Rel(normRoot, path)
		if relErr != nil {
			return relErr
		}

		entryInfo, infoErr := d.Info()
		if infoErr != nil {
			return infoErr
		}

		unifiedRel := filepath.ToSlash(rel)
		result[unifiedRel] = core.FileMeta{
			RelPath:     unifiedRel,
			Size:        entryInfo.Size(),
			ModTimeUnix: entryInfo.ModTime().Unix(),
			IsDirectory: false,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// CopyFileAtomic safely copies a file using a temporary destination and atomic rename.
func (a *FSAdapter) CopyFileAtomic(src, dst string) error {
	normSrc := core.NormalizePath(src)
	normDst := core.NormalizePath(dst)

	srcFile, err := os.Open(normSrc)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstDir := filepath.Dir(normDst)
	if dstDir != "" && dstDir != "." {
		if err := os.MkdirAll(dstDir, DefaultDirPermission); err != nil {
			return err
		}
	}

	tmpDst := normDst + TempFileSuffix
	tmpFile, err := os.OpenFile(tmpDst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(tmpFile, srcFile)
	closeErr := tmpFile.Close()

	if copyErr != nil {
		_ = os.Remove(tmpDst)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(tmpDst)
		return closeErr
	}

	// Preserve modification time
	_ = os.Chtimes(tmpDst, srcInfo.ModTime(), srcInfo.ModTime())

	// Remove target if exists on Windows before rename
	if err := os.Rename(tmpDst, normDst); err != nil {
		_ = os.Remove(normDst)
		if renameErr := os.Rename(tmpDst, normDst); renameErr != nil {
			_ = os.Remove(tmpDst)
			return errors.New("failed to atomic rename: " + renameErr.Error())
		}
	}

	return nil
}
