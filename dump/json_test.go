package dump

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/relex/fluentlib/testdata"
	"github.com/relex/fluentlib/util"
	"github.com/stretchr/testify/assert"
)

func TestGenerateExpectedOutputs(t *testing.T) {
	if !util.IsTestGenerationMode() {
		return
	}

	t.Log("regenerate log outputs...")

	for _, fn := range testdata.ListInputFiles(t) {
		wrt := &bytes.Buffer{}
		assert.Nil(t, PrintChunkFileInJSON(fn, true, wrt))

		expectedFn := testdata.GetOutputFilename(t, fn)
		t.Logf("regenerate %s", expectedFn)
		assert.Nil(t, ioutil.WriteFile(expectedFn, wrt.Bytes(), 0644), expectedFn)
	}
}

func TestPrintChunkFilesInJSON(t *testing.T) {
	if util.IsTestGenerationMode() {
		return
	}

	for _, fn := range testdata.ListInputFiles(t) {
		expectedFn := testdata.GetOutputFilename(t, fn)
		expected, readErr := ioutil.ReadFile(expectedFn)
		assert.Nil(t, readErr, expectedFn)

		wrt := &bytes.Buffer{}
		assert.Nil(t, PrintChunkFileInJSON(fn, true, wrt))
		assert.Equal(t, string(expected), wrt.String(), fn)
	}
}
