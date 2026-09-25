//go:build amd64 || arm64

package log

import (
	"strings"
	"testing"
)

// TestNeedEscapeBlocks checks the vector kernel against the escapes table byte
// by byte: for every byte value it feeds a full vector of that byte, which
// covers the range test, every equality test and the table lookups.
func TestNeedEscapeBlocks(t *testing.T) {
	for b := 0; b < 256; b++ {
		s := strings.Repeat(string([]byte{byte(b)}), 16)
		if got, want := needEscapeBlocks(s), escapes[byte(b)]; got != want {
			t.Errorf("needEscapeBlocks(%#02x x16) = %v, want %v", b, got, want)
		}
	}

	// A byte in the last byte of an otherwise clean vector must still be seen.
	for b := 0; b < 256; b++ {
		s := strings.Repeat("x", 15) + string([]byte{byte(b)})
		if got, want := needEscapeBlocks(s), escapes[byte(b)]; got != want {
			t.Errorf("needEscapeBlocks(x..x,%#02x) = %v, want %v", b, got, want)
		}
	}
}

func BenchmarkNeedEscapeBlocks(b *testing.B) {
	for _, c := range []struct {
		name string
		n    int
	}{
		{"192", 192},
		{"1024", 1024},
		{"4096", 4096},
	} {
		b.Run(c.name, func(b *testing.B) {
			s := strings.Repeat("x", c.n)
			b.SetBytes(int64(len(s)))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if needEscapeBlocks(s) {
					b.Fatal("unexpected escape")
				}
			}
		})
	}
}
