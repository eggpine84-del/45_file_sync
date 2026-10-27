package core

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	PathSeparatorSlash        = "/"
	PathSeparatorBackslash    = "\\"
	PathEmptyErrMsg           = "source or target path cannot be empty"
	PathIdenticalErrMsg       = "source and target paths cannot be identical"
	PathNestingSourceInTarget = "source path cannot be inside target path"
	PathNestingTargetInSource = "target path cannot be inside source path"
	PathNoSourcesErrMsg       = "at least one source path is required"
	PathNoTargetsErrMsg       = "at least one target path is required"
)

// NormalizePath standardizes separators and cleans up trailing slashes.
func NormalizePath(p string) string {
	cleaned := strings.TrimSpace(p)
	if cleaned == "" {
		return ""
	}
	cleaned = filepath.Clean(cleaned)
	unified := filepath.ToSlash(cleaned)
	return strings.TrimRight(unified, PathSeparatorSlash)
}


// ValidatePathPair ensures two paths do not cause recursion, collision, or invalid states.
func ValidatePathPair(source, target string) error {
	normSrc := NormalizePath(source)
	normDst := NormalizePath(target)

	if normSrc == "" || normDst == "" {
		return errors.New(PathEmptyErrMsg)
	}

	if normSrc == normDst {
		return errors.New(PathIdenticalErrMsg)
	}

	prefixSrc := normSrc + PathSeparatorSlash
	prefixDst := normDst + PathSeparatorSlash

	if strings.HasPrefix(normDst, prefixSrc) {
		return errors.New(PathNestingTargetInSource)
	}

	if strings.HasPrefix(normSrc, prefixDst) {
		return errors.New(PathNestingSourceInTarget)
	}

	return nil
}

// ValidateBatch validates all source-target pairs within a preset batch.
func ValidateBatch(sources []string, targets []string) error {
	if len(sources) == 0 {
		return errors.New(PathNoSourcesErrMsg)
	}
	if len(targets) == 0 {
		return errors.New(PathNoTargetsErrMsg)
	}

	for _, src := range sources {
		for _, dst := range targets {
			if err := ValidatePathPair(src, dst); err != nil {
				return err
			}
		}
	}

	return nil
}
