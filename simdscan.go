//go:build gc && (amd64 || arm64)

package log

// needEscapeBlocks reports whether the first len(b) bytes contain a byte that
// Entry.escapes rewrites. len(b) must be a positive multiple of 16; the caller
// handles the tail. It is implemented in assembly per architecture, and the
// vector predicate is 0x08..0x0d plus '"', '\”, '<' and '\\', a superset of
// the escapes table that only adds 0x0b, which escapes copies verbatim.
//
//go:noescape
func needEscapeBlocks(b string) bool

// needEscapeSIMD reports whether s contains a byte that Entry.escapes
// rewrites.
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
