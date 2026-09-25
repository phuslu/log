package log

import (
	"math/rand"
	"strings"
	"testing"
)

// stringReference is the original scalar implementation, used as the
// reference for the vectorized scan: the two must produce identical bytes for
// every input.
func stringReference(dst []byte, s string) []byte {
	for _, c := range []byte(s) {
		if escapes[c] {
			return appendEscapedString(dst, s)
		}
	}
	return append(dst, s...)
}

// TestStringDifferential checks string and byte fields against the original
// scalar implementation. It covers every byte, both sides of the dispatch
// threshold and vector boundaries, and random data.
func TestStringDifferential(t *testing.T) {
	interesting := []byte("\"\\<'\b\f\n\r\t\x00\x0babc")
	corpus := []string{"", strings.Repeat("x", 15), strings.Repeat("x", 16), strings.Repeat("x", 17)}
	for i := range 256 {
		for _, n := range []int{16, 32, 64} {
			corpus = append(corpus, strings.Repeat(string([]byte{byte(i)}), n))
		}
	}

	random := rand.New(rand.NewSource(20260913))
	for _, n := range []int{0, 1, 2, 15, 16, 17, 23, 24, simdEscapeThreshold - 1, simdEscapeThreshold, simdEscapeThreshold + 1, 33, 63, 64, 65, 127, 128, 1000} {
		for k := range 200 {
			b := make([]byte, n)
			for i := range b {
				switch k % 4 {
				case 0:
					b[i] = byte(random.Intn(256))
				case 1:
					b[i] = 'a'
				case 2:
					b[i] = interesting[random.Intn(len(interesting))]
				default:
					b[i] = byte(random.Intn(0x20))
				}
			}
			corpus = append(corpus, string(b))
		}
	}
	corpus = append(corpus, strings.Repeat("x", 1000)+"\n", strings.Repeat("x", 16)+`"`)

	got := new(Entry)
	want := new(Entry)
	for _, s := range corpus {
		got.buf = got.buf[:0]
		want.buf = want.buf[:0]
		want.buf = append(want.buf, `,"value":"`...)
		want.buf = stringReference(want.buf, s)
		want.buf = append(want.buf, '"')
		got.Str("value", s)
		if string(got.buf) != string(want.buf) {
			t.Fatalf("Entry.Str(%q) = %q, want %q", s, got.buf, want.buf)
		}
		got.buf = got.buf[:0]
		got.Bytes("value", []byte(s))
		if string(got.buf) != string(want.buf) {
			t.Fatalf("Entry.Bytes(%q) = %q, want %q", s, got.buf, want.buf)
		}
	}
}

// Check the emitters against literal output, independently of the escape
// table and of stringReference, which also calls appendEscapedString.
func TestAppendEscaped(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain", "ordinary text", "ordinary text"},
		{"quotes", "a\"b\\c", `a\"b\\c`},
		{"html", "<'>&", `\u003c\u0027>&`},
		{"unicode", "你好🙂\u2028\u2029", "你好🙂\u2028\u2029"},
		{"raw_bytes", "\xff\x80\n\xfe", "\xff\x80\\n\xfe"},
		{
			"controls",
			"\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f" +
				"\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f",
			`\u0000\u0001\u0002\u0003\u0004\u0005\u0006\u0007\u0008\t\n\u000b\u000c\r\u000e\u000f` +
				`\u0010\u0011\u0012\u0013\u0014\u0015\u0016\u0017\u0018\u0019\u001a\u001b\u001c\u001d\u001e\u001f`,
		},
		{
			"sparse",
			"\n" + strings.Repeat("x", 4096) + "\x00tail\t",
			`\n` + strings.Repeat("x", 4096) + `\u0000tail\t`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const prefix = "prefix:"
			for _, capacity := range []int{len(prefix), len(prefix) + len(tc.want)} {
				dst := make([]byte, len(prefix), capacity)
				copy(dst, prefix)
				if got := appendEscapedString(dst, tc.in); string(got) != prefix+tc.want {
					t.Errorf("appendEscapedString = %q, want %q", got, prefix+tc.want)
				}
				dst = make([]byte, len(prefix), capacity)
				copy(dst, prefix)
				if got := appendEscapedBytes(dst, []byte(tc.in)); string(got) != prefix+tc.want {
					t.Errorf("appendEscapedBytes = %q, want %q", got, prefix+tc.want)
				}
			}
		})
	}
}

func benchmarkNeedEscapeSIMD(b *testing.B, s string) {
	b.SetBytes(int64(len(s)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if needEscapeSIMD(s) {
			b.Fatal("unexpected escape")
		}
	}
}

func BenchmarkNeedEscapeSIMD(b *testing.B) {
	for _, c := range []struct {
		name string
		n    int
	}{
		{"16", 16},
		{"64", 64},
		{"200", 200},
		{"1024", 1024},
		{"4096", 4096},
	} {
		b.Run(c.name, func(b *testing.B) {
			benchmarkNeedEscapeSIMD(b, strings.Repeat("x", c.n))
		})
	}
}

func BenchmarkStringEscaped(b *testing.B) {
	s := strings.Repeat("x", 200) + "\n"
	e := new(Entry)
	b.SetBytes(int64(len(s)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.buf = e.buf[:0]
		e.buf = appendLoggerString2(e.buf, s)
	}
}

// BenchmarkEscapeFields tracks the actual callers around the scalar/SIMD
// crossover; benchmarking a function value would hide scalar inlining.
func BenchmarkEscapeFields(b *testing.B) {
	for _, c := range []struct {
		name string
		n    int
	}{
		{"3", 3}, {"16", 16}, {"31", 31}, {"32", 32}, {"200", 200},
	} {
		s := strings.Repeat("x", c.n)
		data := []byte(s)
		b.Run(c.name+"/Str", func(b *testing.B) {
			e := &Entry{buf: make([]byte, 0, len(s)+16)}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				e.buf = e.buf[:0]
				e.Str("value", s)
			}
		})
		b.Run(c.name+"/Bytes", func(b *testing.B) {
			e := &Entry{buf: make([]byte, 0, len(data)+16)}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				e.buf = e.buf[:0]
				e.Bytes("value", data)
			}
		})
	}
}

// Exercise both sparse escapes, where skipping ordinary bytes dominates,
// and dense escapes, where the dispatch and output cost dominate.
func BenchmarkEscapedFields(b *testing.B) {
	for _, tc := range []struct {
		name string
		s    string
	}{
		{"newline16", strings.Repeat("x", 15) + "\n"},
		{"newline200", strings.Repeat("x", 199) + "\n"},
		{"newline4096", strings.Repeat("x", 4095) + "\n"},
		{"quote200", strings.Repeat("x", 199) + "\""},
		{"slash200", strings.Repeat("x", 199) + "\\"},
		{"nul200", strings.Repeat("x", 199) + "\x00"},
		{"json", strings.Repeat("{\"field\":\"value\"}\n", 11)},
		{"newlines", strings.Repeat("\n", 200)},
		{"quotes", strings.Repeat("\"\\", 100)},
		{"html", strings.Repeat("<'", 100)},
		{"backspace", strings.Repeat("\b\f", 100)},
		{"controls", strings.Repeat("\x00\x01\x1b\x1f", 50)},
	} {
		b.Run(tc.name+"/Str", func(b *testing.B) {
			e := Entry{buf: make([]byte, 0, len(tc.s)*6+16)}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				e.buf = e.buf[:0]
				e.Str("value", tc.s)
			}
		})
		b.Run(tc.name+"/Bytes", func(b *testing.B) {
			data := []byte(tc.s)
			e := Entry{buf: make([]byte, 0, len(data)*6+16)}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				e.buf = e.buf[:0]
				e.Bytes("value", data)
			}
		})
	}
}
