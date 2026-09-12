//go:build gc && (amd64 || arm64)

package log

const useSIMDEscape = true

// needEscapeBlocks reports whether b contains a byte set in the escapes table.
// len(b) must be a positive multiple of 16; the caller handles the tail.
// The vector predicate matches the table exactly, including the exclusion of
// 0x00 and 0x0b.
//
//go:noescape
func needEscapeBlocks(b string) bool

// needEscapeSIMD reports whether s contains a byte set in the escapes table.
func needEscapeSIMD(s string) bool {
	if n := len(s) &^ 15; n > 0 && needEscapeBlocks(s[:n]) {
		return true
	}
	for _, c := range []byte(s[len(s)&^15:]) {
		if escapes[c] {
			return true
		}
	}
	return false
}
