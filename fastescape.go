package log

// simdEscapeThreshold keeps values shorter than two vectors on the scalar path.
const simdEscapeThreshold = 31

var escapes = [256]bool{
	'"':  true,
	'<':  true,
	'\'': true,
	'\\': true,
	'\b': true,
	'\f': true,
	'\n': true,
	'\r': true,
	'\t': true,
}

func appendEscapedBytes(dst, b []byte) []byte {
	n := len(b)
	j := 0
	if n > 0 {
		// Hint the compiler to remove bounds checks in the loop below.
		_ = b[n-1]
	}
	for i := 0; i < n; i++ {
		switch b[i] {
		case '"':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', '"')
			j = i + 1
		case '\\':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', '\\')
			j = i + 1
		case '\n':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', 'n')
			j = i + 1
		case '\r':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', 'r')
			j = i + 1
		case '\t':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', 't')
			j = i + 1
		case '\f':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '0', 'c')
			j = i + 1
		case '\b':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '0', '8')
			j = i + 1
		case '<':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '3', 'c')
			j = i + 1
		case '\'':
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '2', '7')
			j = i + 1
		case 0:
			dst = append(dst, b[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '0', '0')
			j = i + 1
		}
	}
	return append(dst, b[j:]...)
}

func appendEscapedString(dst []byte, s string) []byte {
	n := len(s)
	j := 0
	if n > 0 {
		// Hint the compiler to remove bounds checks in the loop below.
		_ = s[n-1]
	}
	for i := 0; i < n; i++ {
		switch s[i] {
		case '"':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', '"')
			j = i + 1
		case '\\':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', '\\')
			j = i + 1
		case '\n':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', 'n')
			j = i + 1
		case '\r':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', 'r')
			j = i + 1
		case '\t':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', 't')
			j = i + 1
		case '\f':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '0', 'c')
			j = i + 1
		case '\b':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '0', '8')
			j = i + 1
		case '<':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '3', 'c')
			j = i + 1
		case '\'':
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '2', '7')
			j = i + 1
		case 0:
			dst = append(dst, s[j:i]...)
			dst = append(dst, '\\', 'u', '0', '0', '0', '0')
			j = i + 1
		}
	}
	return append(dst, s[j:]...)
}

// appendLoggerString appends a string using the scalar escape scan.
func appendLoggerString(dst []byte, s string) []byte {
	for _, c := range []byte(s) {
		if escapes[c] {
			return appendEscapedString(dst, s)
		}
	}
	return append(dst, s...)
}

// appendLoggerBytes appends bytes using the scalar escape scan.
func appendLoggerBytes(dst, b []byte) []byte {
	for _, c := range b {
		if escapes[c] {
			return appendEscapedBytes(dst, b)
		}
	}
	return append(dst, b...)
}

// appendLoggerString2 appends a string after the SIMD escape scan.
func appendLoggerString2(dst []byte, s string) []byte {
	if needEscapeSIMD(s) {
		return appendEscapedString(dst, s)
	}
	return append(dst, s...)
}

// appendLoggerBytes2 appends bytes after the SIMD escape scan.
func appendLoggerBytes2(dst, b []byte) []byte {
	if needEscapeSIMD(b2s(b)) {
		return appendEscapedBytes(dst, b)
	}
	return append(dst, b...)
}
