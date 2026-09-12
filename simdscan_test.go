package log

import (
	"math/rand"
	"strings"
	"testing"
)

// stringReference is the original scalar implementation of Entry.string, used
// as the reference for the vectorized scan: the two must produce identical
// bytes for every input.
func stringReference(e *Entry, s string) {
	for _, c := range []byte(s) {
		if escapes[c] {
			e.escapes(s)
			return
		}
	}
	e.buf = append(e.buf, s...)
}

// TestStringDifferential checks Entry.string against its scalar reference over
// a corpus that covers every byte, every length class around the vector width
// and random data, because the vector kernel has its own predicate that must
// not drift from the escapes table.
func TestStringDifferential(t *testing.T) {
	interesting := []byte("\"\\<'\b\f\n\r\t\x00\x0babc")
	corpus := []string{"", strings.Repeat("x", 15), strings.Repeat("x", 16), strings.Repeat("x", 17)}
	for i := 0; i < 256; i++ {
		corpus = append(corpus, strings.Repeat(string(rune(byte(i))), 16))
	}

	random := rand.New(rand.NewSource(20260913))
	for _, n := range []int{0, 1, 2, 15, 16, 17, 31, 32, 33, 63, 64, 65, 127, 128, 1000} {
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
		got.string(s)
		stringReference(want, s)
		if string(got.buf) != string(want.buf) {
			t.Fatalf("Entry.string(%q) = %q, want %q", s, got.buf, want.buf)
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
		e.string(s)
	}
}
