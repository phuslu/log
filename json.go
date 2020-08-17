// json.go extracted from https://github.com/tidwall/gjson/blob/master/gjson.go

package log

import (
	"strconv"
	"unicode/utf16"
	"unicode/utf8"
)

type jsonType int

const (
	jsonNull   jsonType = 0
	jsonFalse  jsonType = 1
	jsonTrue   jsonType = 2
	jsonString jsonType = 3
	jsonNumber jsonType = 4
	jsonJSON   jsonType = 5
)

type jsonResult struct {
	// Type is the json type
	Type jsonType
	// Raw is the raw json
	Raw string
	// Str is the json string/number
	Str string
}

func parseJson(data string) (results []jsonResult, err error) {
	var keys bool
	var i int
	var key, value jsonResult
	for ; i < len(data); i++ {
		if data[i] == '{' {
			i++
			key.Type = jsonString
			keys = true
			break
		} else if data[i] == '[' {
			i++
			break
		}
		if data[i] > ' ' {
			return
		}
	}
	var str string
	var vesc bool
	var ok bool
	for ; i < len(data); i++ {
		if keys {
			if data[i] != '"' {
				continue
			}
			i, str, vesc, ok = parseJsonString(data, i+1)
			if !ok {
				return
			}
			if vesc {
				key.Str = unescapeJsonString(str[1 : len(str)-1])
			} else {
				key.Str = str[1 : len(str)-1]
			}
			key.Raw = str
		}
		for ; i < len(data); i++ {
			if data[i] <= ' ' || data[i] == ',' || data[i] == ':' {
				continue
			}
			break
		}
		i, value, ok = parseJsonAny(data, i, true)
		if !ok {
			return
		}
		results = append(results, key, value)
	}
	return
}

func parseJsonString(json string, i int) (int, string, bool, bool) {
	var s = i
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

// unescapeJsonString unescapes a string
func unescapeJsonString(json string) string {
	var str = make([]byte, 0, len(json))
	for i := 0; i < len(json); i++ {
		switch {
		default:
			str = append(str, json[i])
		case json[i] < ' ':
			return string(str)
		case json[i] == '\\':
			i++
			if i >= len(json) {
				return string(str)
			}
			switch json[i] {
			default:
				return string(str)
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
					return string(str)
				}
				m, _ := strconv.ParseUint(json[i+1:4], 16, 64)
				r := rune(m)
				i += 5
				if utf16.IsSurrogate(r) {
					// need another code
					if len(json[i:]) >= 6 && json[i] == '\\' &&
						json[i+1] == 'u' {
						// we expect it to be correct so just consume it
						m, _ = strconv.ParseUint(json[i+2:4], 16, 64)
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
	return string(str)
}

// parseJsonAny parses the next value from a json string.
// A Result is returned when the hit param is set.
// The return values are (i int, res Result, ok bool)
func parseJsonAny(json string, i int, hit bool) (int, jsonResult, bool) {
	var res jsonResult
	var val string
	for ; i < len(json); i++ {
		if json[i] == '{' || json[i] == '[' {
			i, val = parseJsonSquash(json, i)
			if hit {
				res.Raw = val
				res.Type = jsonJSON
			}
			return i, res, true
		}
		if json[i] <= ' ' {
			continue
		}
		switch json[i] {
		case '"':
			i++
			var vesc bool
			var ok bool
			i, val, vesc, ok = parseJsonString(json, i)
			if !ok {
				return i, res, false
			}
			if hit {
				res.Type = jsonString
				res.Raw = val
				if vesc {
					res.Str = unescapeJsonString(val[1 : len(val)-1])
				} else {
					res.Str = val[1 : len(val)-1]
				}
			}
			return i, res, true
		case '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			i, val = parseJsonNumber(json, i)
			if hit {
				res.Raw = val
				res.Type = jsonNumber
				res.Str = val
			}
			return i, res, true
		case 't', 'f', 'n':
			vc := json[i]
			i, val = parseJsonLiteral(json, i)
			if hit {
				res.Raw = val
				switch vc {
				case 't':
					res.Type = jsonTrue
				case 'f':
					res.Type = jsonFalse
				}
				return i, res, true
			}
		}
	}
	return i, res, false
}

func parseJsonSquash(json string, i int) (int, string) {
	// expects that the lead character is a '[' or '{' or '('
	// squash the value, ignoring all nested arrays and objects.
	// the first '[' or '{' or '(' has already been read
	s := i
	i++
	depth := 1
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

func parseJsonNumber(json string, i int) (int, string) {
	var s = i
	i++
	for ; i < len(json); i++ {
		if json[i] <= ' ' || json[i] == ',' || json[i] == ']' ||
			json[i] == '}' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}

func parseJsonLiteral(json string, i int) (int, string) {
	var s = i
	i++
	for ; i < len(json); i++ {
		if json[i] < 'a' || json[i] > 'z' {
			return i, json[s:i]
		}
	}
	return i, json[s:]
}
