// Package dump provides functions to dump messages and logs in various formats
//
// For testing and debugging only, no performance critical.
package dump

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/relex/gotils/logger"
)

// PrintFileOrDirectories prints log records from a list of files or directories of files
func PrintFileOrDirectories(pathList []string, ignoreError bool) error {
	bufWriter := bufio.NewWriterSize(os.Stdout, 1048576)
	defer bufWriter.Flush()
	for _, path := range pathList {
		if err := printChunkFilesInDir(path, bufWriter, ignoreError); err != nil {
			return err
		}
	}
	return nil
}

func printChunkFilesInDir(root string, bufWriter *bufio.Writer, ignoreError bool) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, walkErr error) error {
		if walkErr != nil {
			if ignoreError {
				logger.Errorf("failed to walk %s: %v", path, walkErr)
				return nil
			} else {
				return walkErr
			}
		}
		if info.IsDir() {
			return nil
		}
		if decErr := PrintChunkFileInJSON(path, false, bufWriter); decErr != nil {
			if ignoreError {
				logger.Errorf("failed to print %s: %v", path, decErr)
			} else {
				return fmt.Errorf("failed to print %s: %w", path, decErr)
			}
		}
		return nil
	})
}
