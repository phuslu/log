package log

import (
	"encoding"
	"encoding/json"
	"testing"
	"time"
)

func TestXIDParse(t *testing.T) {
	_, err := ParseXID("ab")
	if err == nil {
		t.Errorf("ParseXID should error")
	}

	_, err = ParseXID("\x012345678901234567890")
	if err == nil {
		t.Errorf("ParseXID should error")
	}

	for i := 0; i < 10; i++ {
		x := NewXID()
		got, _ := ParseXID(x.String())
		want := x
		if got != want {
			t.Errorf("ParseXID(x) want=%+v got=%+v", want, got)
		}
	}
}

func TestXIDTime(t *testing.T) {
	x := NewXID()
	time.Sleep(1 * time.Second)
	y := NewXID()

	if y.Time().Sub(x.Time()) != 1*time.Second {
		t.Errorf("XID.Time not correct")
	}

	if string(x.Machine()) != string(y.Machine()) {
		t.Errorf("XID.Machine not correct")
	}

	if x.Pid() != y.Pid() {
		t.Errorf("XID.Pid not correct")
	}

	if y.Counter()-x.Counter() != 1 {
		t.Errorf("XID.Counter not correct")
	}
}

func TestXIDMarshalJSON(t *testing.T) {
	s := struct {
		XID XID `json:"id"`
	}{}
	copy(s.XID[:], "012345678912")

	data, err := json.Marshal(s)
	if err != nil {
		t.Errorf("json.Marshal(s) err: %+v", err)
	}

	got := string(data)
	want := `{"id":"60oj4cpk6kr3ee1p64p0"}`
	if got != want {
		t.Errorf("json.Marshal(s) want=%+v got=%+v", want, got)
	}

	err = json.Unmarshal(data, &s)
	if err != nil {
		t.Errorf("json.Marshal(s) err: %+v", err)
	}

	got = string(s.XID[:])
	want = "012345678912"
	if got != want {
		t.Errorf("json.Marshal(s) want=%#v got=%#v", want, got)
	}
}

func TestXIDMarshalJSONNull(t *testing.T) {
	s := struct {
		XID XID `json:"id"`
	}{}

	data, err := json.Marshal(s)
	if err != nil {
		t.Errorf("json.Marshal(s) err: %+v", err)
	}

	got := string(data)
	want := `{"id":null}`
	if got != want {
		t.Errorf("json.Marshal(s) want=%+v got=%+v", want, got)
	}

	err = json.Unmarshal(data, &s)
	if err != nil {
		t.Errorf("json.Marshal(s) err: %+v", err)
	}

	got = string(s.XID[:])
	want = "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
	if got != want {
		t.Errorf("json.Marshal(s) want=%#v got=%#v", want, got)
	}
}

func TestXIDMarshalText(t *testing.T) {
	x := NewXID()

	var m encoding.TextMarshaler = x
	text, err := m.MarshalText()
	if err != nil {
		t.Errorf("xid.MarshalText() err: %+v", err)
	}

	var y XID
	var u encoding.TextUnmarshaler = &y
	err = u.UnmarshalText(text)
	if err != nil {
		t.Errorf("xid.UnmarshalText() err: %+v", err)
	}

	if x != y {
		t.Error("MarshalText()/UnmarshalText mismatched")
	}
}
