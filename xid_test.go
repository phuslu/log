package log

import (
	"encoding"
	"encoding/json"
	"os"
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

	for range 10 {
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

	if want := uint16(os.Getpid()); x.Pid() != want || y.Pid() != want {
		t.Errorf("XID.Pid want=%d got=%d,%d", want, x.Pid(), y.Pid())
	}

	if y.Counter()-x.Counter() != 1 {
		t.Errorf("XID.Counter not correct")
	}
}

// TestXIDPidCounter guards the packing of the pid and counter fields: the high
// bits of the 32-bit counter must not leak into the pid bytes.
func TestXIDPidCounter(t *testing.T) {
	saved := counter
	t.Cleanup(func() { counter = saved })

	counters := []uint32{0, 1, 0x00ffffff, 0x01000000, 0x00fedcba, 0xffffffff, 0xdeadbeef}
	for _, i := range counters {
		counter = i - 1 // NewXIDWithTime bumps counter to i; 0xffffffff+1 wraps to 0.
		x := NewXIDWithTime(0)
		if got, want := x.Pid(), uint16(pid); got != want {
			t.Errorf("counter=%#x: XID.Pid() = %#x, want %#x", i, got, want)
		}
		if got, want := x.Counter(), i&0xffffff; got != want {
			t.Errorf("counter=%#x: XID.Counter() = %#x, want %#x", i, got, want)
		}
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
