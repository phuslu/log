package log

import (
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

// FormatterArgs is a parsed sturct from json input
type FormatterArgs struct {
	Time      string // "2019-07-10T05:35:54.277Z"
	Level     string // "info"
	Caller    string // "prog.go:42"
	Goid      string // "123"
	Message   string // "a structure message"
	Stack     string // "<stack string>"
	KeyValues []struct {
		Key   string // "foo"
		Value string // "bar"
	}
}

// parseFormatterArgs extracts json string to json items
func parseFormatterArgs(json []byte, args *FormatterArgs) {
	var keys bool
	var i int
	var key, value []byte
	_ = json[len(json)-1] // remove bounds check
	for ; i < len(json); i++ {
		if json[i] == '{' {
			i++
			keys = true
			break
		} else if json[i] == '[' {
			i++
			break
		}
		if json[i] > ' ' {
			return
		}
	}
	var str []byte
	var typ byte
	var ok bool
	for ; i < len(json); i++ {
		if keys {
			if json[i] != '"' {
				continue
			}
			i, str, _, ok = jsonParseString(json, i+1)
			if !ok {
				return
			}
			key = str[1 : len(str)-1]
		}
		for ; i < len(json); i++ {
			if json[i] <= ' ' || json[i] == ',' || json[i] == ':' {
				continue
			}
			break
		}
		i, typ, value, ok = jsonParseAny(json, i, true)
		if !ok {
			return
		}
		if args.Time == "" {
			switch typ {
			case 's', 'S':
				args.Time = b2s(value[1 : len(value)-1])
			default:
				args.Time = b2s(value)
			}
			continue
		}
		switch b2s(key) {
		case "level":
			switch typ {
			case 's', 'S':
				args.Level = b2s(value[1 : len(value)-1])
			default:
				args.Level = b2s(value)
			}
		case "goid":
			switch typ {
			case 's', 'S':
				args.Goid = b2s(value[1 : len(value)-1])
			default:
				args.Goid = b2s(value)
			}
		case "caller":
			switch typ {
			case 's', 'S':
				args.Caller = b2s(value[1 : len(value)-1])
			default:
				args.Caller = b2s(value)
			}
		case "stack":
			switch typ {
			case 'S':
				args.Stack = b2s(jsonUnescape(value[1:len(value)-1], value[:0]))
			case 's':
				args.Stack = b2s(value[1 : len(value)-1])
			default:
				args.Stack = b2s(value)
			}
		case "message", "msg":
			switch typ {
			case 'S':
				value = jsonUnescape(value[1:len(value)-1], value[:0])
				if len(value) > 0 && value[len(value)-1] == '\n' {
					args.Message = b2s(value[1 : len(value)-1])
				} else {
					args.Message = b2s(value)
				}
			case 's':
				args.Message = b2s(value[1 : len(value)-1])
			default:
				args.Message = b2s(value)
			}
		default:
			switch typ {
			case 'S':
				value = jsonUnescape(value[1:len(value)-1], value[:0])
			case 's':
				value = value[1 : len(value)-1]
			}
			args.KeyValues = append(args.KeyValues, struct {
				Key, Value string
			}{b2s(key), b2s(value)})
		}
	}

	if len(args.Level) == 0 {
		args.Level = "????"
	}
}

func jsonParseString(json []byte, i int) (int, []byte, bool, bool) {
	var s = i
	_ = json[len(json)-1] // remove bounds check
	for ; i < len(json); i++ {
		if json[i] > '\\' {
			continue
		}
		if json[i] == '"' {
			return i + 1, json[s-1 : i+1], false, true
		}
		if json[i] == '\\' {
			i++
			for ; i < len(json); i++ {
				if json[i] > '\\' {
					continue
				}
				if json[i] == '"' {
					// look for an escaped slash
					if json[i-1] == '\\' {
						n := 0
						for j := i - 2; j > 0; j-- {
							if json[j] != '\\' {
								break
							}
							n++
						}
						if n%2 == 0 {
							continue
						}
					}
					return i + 1, json[s-1 : i+1], true, true
				}
			}
			break
		}
	}
	return i, json[s-1:], false, false
}

// jsonParseAny parses the next value from a json string.
// A Result is returned when the hit param is set.
// The return values are (i int, res Result, ok bool)
func jsonParseAny(json []byte, i int, hit bool) (int, byte, []byte, bool) {
	var typ byte
	var val []byte
	_ = json[len(json)-1] // remove bounds check
	for ; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			i, val = jsonParseSquash(json, i)
			if hit {
				typ = 'o'
			}
			return i, typ, val, true
		}
		if json[i] <= ' ' {
			continue
		}
		switch json[i] {
		case '"':
			i++
			var vesc bool
			var ok bool
			i, val, vesc, ok = jsonParseString(json, i)
			typ = 's'
			if !ok {
				return i, typ, val, false
			}
			if hit && vesc {
				typ = 'S'
			}
			return i, typ, val, true
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i, val = jsonParseNumber(json, i)
			if hit {
				typ = 'n'
			}
			return i, typ, val, true
		case 't', 'f', 'n':
			vc := json[i]
			i, val = jsonParseLiteral(json, i)
			if hit {
				switch vc {
				case 't':
					typ = 't'
				case 'f':
					typ = 'f'
				}
				return i, typ, val, true
			}
		}
	}
	return i, typ, val, false
}

func jsonParseSquash(json []byte, i int) (int, []byte) {
	// expects that the lead character is a '[' or '{' or '('
	// squash the value, ignoring all nested arrays and objects.
	// the first '[' or '{' or '(' has already been read
	s := i
	i++
	depth := 1
	_ = json[len(json)-1] // remove bounds check
	for ; i < len(json); i++ {
		if json[i] >= '"' && json[i] <= '}' {
			switch json[i] {
			case '"':
				i++
				s2 := i
				for ; i < len(json); i++ {
					if json[i] > '\\' {
						continue
					}
					if json[i] == '"' {
						// look for an escaped slash
						if json[i-1] == '\\' {
							n := 0
							for j := i - 2; j > s2-1; j-- {
								if json[j] != '\\' {
									break
								}
								n++
							}
							if n%2 == 0 {
								continue
							}
						}
						break
					}
				}
			case '{', '[', '(':
				depth++
			case '}', ']', ')':
				depth--
				if depth == 0 {
					i++
					return i, json[s:i]
				}
			}
		}
	}
	return i, json[s:]
}

func jsonParseNumber(json []byte, i int) (int, []byte) {
	var s = i
	i++
	_ = json[len(json)-1] // remove bounds check
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ']' ||
			json[i] == '}' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

func jsonParseLiteral(json []byte, i int) (int, []byte) {
	var s = i
	i++
	_ = json[len(json)-1] // remove bounds check
	for ; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

// jsonUnescape unescapes a string
func jsonUnescape(json, str []byte) []byte {
	_ = json[len(json)-1] // remove bounds check
	for i := 0; i < len(json); i++ {
		switch {
		default:
			str = append(str, json[i])
		case json[i] < ' ':
			return str
		case json[i] == '\\':
			i++
			if i >= len(json) {
				return str
			}
			switch json[i] {
			default:
				return str
			case '\\':
				str = append(str, '\\')
			case '/':
				str = append(str, '/')
			case 'b':
				str = append(str, '\b')
			case 'f':
				str = append(str, '\f')
			case 'n':
				str = append(str, '\n')
			case 'r':
				str = append(str, '\r')
			case 't':
				str = append(str, '\t')
			case '"':
				str = append(str, '"')
			case 'u':
				if i+5 > len(json) {
					return str
				}
				m, _ := strconv.ParseUint(b2s(json[i+1:i+5]), 16, 64)
				r := rune(m)
				i += 5
				if utf16.IsSurrogate(r) {
					// need another code
					if len(json[i:]) >= 6 && json[i] == '\\' &&
						json[i+1] == 'u' {
						// we expect it to be correct so just consume it
						m, _ = strconv.ParseUint(b2s(json[i+2:i+6]), 16, 64)
						r = utf16.DecodeRune(r, rune(m))
						i += 6
					}
				}
				// provide enough space to encode the largest utf8 possible
				str = append(str, 0, 0, 0, 0, 0, 0, 0, 0)
				n := utf8.EncodeRune(str[len(str)-8:], r)
				str = str[:len(str)-8+n]
				i-- // backtrack index by one
			}
		}
	}
	return str
}
