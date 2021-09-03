package dump

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/relex/fluentlib/testdata"
	"github.com/stretchr/testify/assert"
)

var extPattern = regexp.MustCompile(`.ff$`)

func TestGenerateExpectedOutputs(t *testing.T) {
	if !testdata.IsTestGenerationMode() {
		return
	}

	t.Log("regenerate log outputs...")

	inFiles, globErr := filepath.Glob("../testdata/*.ff")
	assert.Nil(t, globErr)

	for _, fn := range inFiles {
		wrt := &bytes.Buffer{}
		assert.Nil(t, PrintChunkFileInJSON(fn, true, wrt))

		expectedFn := extPattern.ReplaceAllString(fn, ".json")
		t.Logf("regenerate %s", expectedFn)
		assert.Nil(t, ioutil.WriteFile(expectedFn, wrt.Bytes(), 0644), expectedFn)
	}
}

func TestPrintChunkFilesInJSON(t *testing.T) {
	if testdata.IsTestGenerationMode() {
		return
	}
	inFiles, globErr := filepath.Glob("../testdata/*.ff")
	assert.Nil(t, globErr)

	for _, fn := range inFiles {
		expectedFn := extPattern.ReplaceAllString(fn, ".json")
		expected, readErr := ioutil.ReadFile(expectedFn)
		assert.Nil(t, readErr, expectedFn)

		wrt := &bytes.Buffer{}
		assert.Nil(t, PrintChunkFileInJSON(fn, true, wrt))
		assert.Equal(t, string(expected), wrt.String(), fn)
	}
}
