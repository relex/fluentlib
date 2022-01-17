package forwardprotocol

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResolveEventPath(t *testing.T) {
	event := EventEntry{
		Time: EventTime{
			time.Date(2022, 1, 14, 10, 30, 55, 123, time.UTC),
		},
		Record: map[string]interface{}{
			"msg": "Hello",
			"http": map[string]interface{}{
				"statusCode": "500",
				"userAgent":  "Chrome",
			},
			"alert": map[string]interface{}{
				"notMap": 100,
			},
		},
	}

	type test struct {
		path []string
		res  interface{}
		err  string
	}

	testList := []test{
		{[]string{"msg"}, "Hello", ""},
		{[]string{"http", "statusCode"}, "500", ""},
		{[]string{"http", "userAgent2"}, nil, "failed to resolve [http userAgent2] at step 2: 'userAgent2' does not exist"},
		{[]string{"http", "notHere", "name"}, nil, "failed to resolve [http notHere name] at step 2: 'notHere' does not exist"},
		{[]string{"alert", "notMap", "id"}, nil, "failed to resolve [alert notMap id] at step 2: 'notMap' is not a map[string]interface{}: type=int value=100"},
	}

	for i, test := range testList {
		title := fmt.Sprintf("test[%d] path=%s", i, test.path)
		result, err := event.ResolvePath(test.path...)
		assert.Equal(t, test.res, result, title)
		if len(test.err) > 0 {
			assert.EqualError(t, err, test.err, title)
		} else {
			assert.Nil(t, err, title)
		}
	}
}
