package log

import (
	"runtime"
	"testing"
)

func TestPcFileLine(t *testing.T) {
	var pc uintptr
	caller1(0, &pc, 1, 1)
	file1, line1 := pcFileLine(pc)
	_, file2, line2, _ := runtime.Caller(0)

	if file1 != file2 {
		t.Errorf("pcFileLine file error: %q != %q", file1, file2)
	}

	if int(line1)+1 != line2 && int(line1)+2 != line2 {
		t.Errorf("pcFileLine line error: %d+2 != %d", line1, line2)
	}
}

func TestPcFileLineName(t *testing.T) {
	var pc uintptr
	caller1(0, &pc, 1, 1)
	file1, line1, name1 := pcFileLineName(pc)
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

	if int(line1)+2 != line2 && int(line1)+3 != line2 {
		t.Errorf("pcFileLine line error: %d+3 != %d", line1, line2)
	}
}

// inlineCaller is small enough for the compiler to inline it into the test, so
// the pc it records is a pc inside inlined code.
func inlineCaller(pc *uintptr) {
	caller1(0, pc, 1, 1)
}

func TestPcFileLineNameInlined(t *testing.T) {
	var pc uintptr
	inlineCaller(&pc)

	file1, line1, name1 := pcFileLineName(pc)
	file2, line2 := pcFileLine(pc)
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()

	if name1 != frame.Function {
		t.Errorf("pcFileLineName name error: %q != %q", name1, frame.Function)
	}
	if file1 != frame.File {
		t.Errorf("pcFileLineName file error: %q != %q", file1, frame.File)
	}
	if int(line1) != frame.Line {
		t.Errorf("pcFileLineName line error: %d != %d", line1, frame.Line)
	}
	if file2 != frame.File {
		t.Errorf("pcFileLine file error: %q != %q", file2, frame.File)
	}
	if int(line2) != frame.Line {
		t.Errorf("pcFileLine line error: %d != %d", line2, frame.Line)
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

func BenchmarkPcFileLineName(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var pc uintptr
		caller1(1, &pc, 1, 1)
		pcFileLineName(pc)
	}
}
