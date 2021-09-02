// Package dump provides functions to dump messages and logs in various formats
//
// For testing and debugging only, no performance critical.
package dump

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/relex/gotils/logger"
)

// PrintFileOrDirectories prints log records from a list of files or directories of files (no nesting)
func PrintFileOrDirectories(pathList []string) {
	bufWriter := bufio.NewWriterSize(os.Stdout, 1048576)
	defer bufWriter.Flush()
	for _, path := range pathList {
		stat, statErr := os.Stat(path)
		if statErr != nil {
			logger.Errorf("input '%s' is not accessible: %v", path, statErr)
			continue
		}
		if stat.IsDir() {
			fileList, err := ioutil.ReadDir(path)
			if err != nil {
				panic(err)
			}
			for _, file := range fileList {
				PrintChunkFileInJSON(filepath.Join(path, file.Name()), false, bufWriter)
			}
		} else {
			PrintChunkFileInJSON(path, false, bufWriter)
		}
	}
}
