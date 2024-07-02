package log

import (
	"testing"
	"time"
)

func TestWalltime(t *testing.T) {
	sec, nsec := walltime()
	t.Log(time.Unix(sec, int64(nsec)))
}
