package log

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestRingWriterSmallSize(t *testing.T) {
	w := NewRingWriter(os.Stderr, 16, 0)
	for i := 0; i < 10; i++ {
		fmt.Fprintf(w, "%s, %d during ring writer 1k buff size\n", timeNow(), i)
	}
}

func BenchmarkRingWriter(b *testing.B) {
	w := NewRingWriter(ioutil.Discard, 1000, 0)

	b.SetParallelism(1000)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		p := []byte(`{"time":"2019-07-10T05:35:54.277Z","level":"info","caller":"pretty.go:42","error":"i am test error","foo":"bar","n":42,"a":[1,2,3],"obj":{"a":[1], "b":{}},"message":"hello json writer"}`)
		for b.Next() {
			w.Write(p)
		}
	})
}
