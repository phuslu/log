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
	for i := 0; i < 256; i++ {
		for _, n := range []int{16, 32, 64} {
			corpus = append(corpus, strings.Repeat(string([]byte{byte(i)}), n))
		}
	}

	random := rand.New(rand.NewSource(20260913))
	for _, n := range []int{0, 1, 2, 15, 16, 17, 23, 24, simdEscapeThreshold - 1, simdEscapeThreshold, simdEscapeThreshold + 1, 33, 63, 64, 65, 127, 128, 1000} {
		for k := 0; k < 200; k++ {
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
