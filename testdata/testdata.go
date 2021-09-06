// Package testdata provides access to shared sample logs for testing
package testdata

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

var inputExtPattern = regexp.MustCompile(`.ff$`)

var absoluteDirPath string

func init() {
	_, thisFile, _, _ := runtime.Caller(0)
	absoluteDirPath = filepath.Dir(thisFile)
}

func ListInputFiles(t *testing.T) []string {
	pwd, pwdErr := os.Getwd()
	if pwdErr != nil {
		t.Fatal("failed to get current dir: %v", pwdErr)
	}
	pattern := absoluteDirPath + "/*.ff"

	inFiles, globErr := filepath.Glob(pattern)
	if globErr != nil {
		t.Fatal("failed to scan test files at path %s: %v", pwd+"/"+pattern, globErr)
	}
	if len(inFiles) == 0 {
		t.Fatal("failed to find test files at path %s: no match", pwd+"/"+pattern)
	}
	return inFiles
}

func GetOutputFilename(t *testing.T, fn string) string {
	outFn := inputExtPattern.ReplaceAllString(fn, ".json")
	if outFn == fn {
		t.Fatalf("invalid input filename %s", fn)
	}
	return outFn
}
