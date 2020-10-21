package log

import (
	"testing"
)

func TestJsonParse(t *testing.T) {
	var jsons = []string{
		`{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"这是一个🌐哦\n","foo":"bar","n":42,"t":true,"f":false,"o":null,"a":[1,2,3],"obj":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer\t123"}`,
		`[1, "a", "b"]`,
	}

	for _, s := range jsons {
		results := appendJsonItems(nil, []byte(s))
		for _, v := range results {
			t.Logf("%c [%s]", v.Type, v.Value)
		}
	}
}
