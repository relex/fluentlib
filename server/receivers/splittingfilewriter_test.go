package receivers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/relex/fluentlib/protocol/forwardprotocol"
	"github.com/stretchr/testify/assert"
)

var splitInputs = []forwardprotocol.Message{
	{
		Tag: "my-app",
		Entries: []forwardprotocol.EventEntry{
			{
				Time: forwardprotocol.EventTime{Time: time.Date(2022, 1, 14, 10, 30, 55, 123, time.UTC)},
				Record: map[string]interface{}{
					"role": "Salesman",
					"msg":  "Log S 1",
				},
			},
			{
				Time: forwardprotocol.EventTime{Time: time.Date(2022, 1, 14, 10, 30, 55, 123, time.UTC)},
				Record: map[string]interface{}{
					"role": "Salesman",
					"msg":  "Log S 2",
				},
			},
			{
				Time: forwardprotocol.EventTime{Time: time.Date(2022, 1, 14, 10, 30, 55, 123, time.UTC)},
				Record: map[string]interface{}{
					"role": "Customer",
					"msg":  "Log C 1",
				},
			},
		},
		Option: forwardprotocol.TransportOption{},
	},
	{
		Tag: "my-app",
		Entries: []forwardprotocol.EventEntry{
			{
				Time: forwardprotocol.EventTime{Time: time.Date(2022, 1, 14, 10, 30, 55, 123, time.UTC)},
				Record: map[string]interface{}{
					"role": "Customer",
					"msg":  "Log C 2",
				},
			},
		},
		Option: forwardprotocol.TransportOption{},
	},
}

var splitOutputs = map[string]string{
	"split-my-app-Customer.json": `[
[
  "my-app",
  1642156255,
  {
    "msg": "Log C 1",
    "role": "Customer"
  }
],
[
  "my-app",
  1642156255,
  {
    "msg": "Log C 2",
    "role": "Customer"
  }
]
]
`,
	"split-my-app-Salesman.json": `[
[
  "my-app",
  1642156255,
  {
    "msg": "Log S 1",
    "role": "Salesman"
  }
],
[
  "my-app",
  1642156255,
  {
    "msg": "Log S 2",
    "role": "Salesman"
  }
]
]
`,
}

func TestSplittingFileWriter(t *testing.T) {
	dirPath, dirErr := os.MkdirTemp("", "fluentlib-split-test-*")
	assert.Nil(t, dirErr)
	defer os.RemoveAll(dirPath)

	recv := NewSplittingFileWriter([]string{"role"}, filepath.Join(dirPath, "split-%s.json"), false)
	for i, msg := range splitInputs {
		assert.Nil(t, recv.Accept(ClientMessage{
			ConnectionID: int64(i),
			Message:      msg,
		}), "input[%d]", i)
	}

	assert.Nil(t, recv.End())

	dir, dirErr := os.Open(dirPath)
	assert.Nil(t, dirErr)
	defer dir.Close()

	fnList, lsErr := dir.Readdirnames(-1)
	assert.Nil(t, lsErr)

	sort.Strings(fnList)
	assert.Equal(t, []string{
		"split-my-app-Customer.json",
		"split-my-app-Salesman.json",
	}, fnList)

	for _, fn := range fnList {
		fdata, ferr := ioutil.ReadFile(filepath.Join(dirPath, fn))
		assert.Nil(t, ferr, "output %s", fn)
		assert.Equal(t, splitOutputs[fn], string(fdata), "output %s", fn)
	}
}
