package log

import (
	"testing"
)

func TestJsonParse(t *testing.T) {
	s := `{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"这是一个🌐哦\n","foo":"bar","n":42,"t":true,"f":false,"o":null,"a":[1,2,3],"obj":{"a":[1,2], "b":{"c":3}},"message":"hello json console color writer\t123"}`

	results, err := jsonParse([]byte(s))
	if err != nil {
		t.Errorf("jsonParse error: %+v", err)
	}

	for _, v := range results {
		t.Logf("%c %s", v.Type, v.Value)
	}
}
