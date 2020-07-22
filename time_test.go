package log

import (
	"testing"
)

func TestTimestampMS(t *testing.T) {
	now := timeNow()
	ts1 := now.Unix()*1000 + int64(now.Nanosecond())/1000000
	ts2 := TimestampMS()

	if ts1 != ts2 {
		t.Errorf("test TestTimestampMS failed, %d != %d", ts1, ts2)
	}
}
