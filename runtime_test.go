package log

import (
	"runtime"
	"testing"
)

func TestPcFileLine(t *testing.T) {
	var pcs [2]uintptr
	runtime.Callers(1, pcs[:])
	file1, line1 := pcFileLine(pcs[0])
	_, file2, line2, _ := runtime.Caller(0)

	if file1 != file2 {
		t.Errorf("pcFileLine file error: %q != %q", file1, file2)
	}

	if int(line1)+2 != line2 {
		t.Errorf("pcFileLine line error: %d+2 != %d", line1, line2)
	}
}

func TestPcNameFileLine(t *testing.T) {
	var pcs [2]uintptr
	runtime.Callers(1, pcs[:])
	name1, file1, line1 := pcNameFileLine(pcs[0])
	t.Log(name1, file1, line1)
	pc, file2, line2, _ := runtime.Caller(0)
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	name2 := frame.Function
	t.Log(name2, file2, line2)

	if name1 != name2 {
		t.Errorf("pcFileLine file error: %q != %q", name1, name2)
	}

	if file1 != file2 {
		t.Errorf("pcFileLine file error: %q != %q", file1, file2)
	}

	if int(line1)+3 != line2 {
		t.Errorf("pcFileLine line error: %d+3 != %d", line1, line2)
	}
}

func BenchmarkPcFileLine(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var pc uintptr
		caller1(1, &pc, 1, 1)
		pcFileLine(pc)
	}
}

func BenchmarkPcNameFileLine(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var pc uintptr
		caller1(1, &pc, 1, 1)
		pcNameFileLine(pc)
	}
}
