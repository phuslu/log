package log

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestGoid(t *testing.T) {
	id := goid()
	a := fmt.Sprintf("goroutine %d ", id)
	data := make([]byte, 1024)
	b := data[:runtime.Stack(data, false)]

	if !strings.HasPrefix(string(b), a) {
		t.Errorf("runtime.Stack return %s, does contains goid %d", b, id)
	}
}
