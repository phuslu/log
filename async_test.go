package log

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestAsyncWriterZero(t *testing.T) {
	w := &AsyncWriter{
		ChannelSize: 0,
		Writer:      IOWriter{os.Stderr},
	}
	for i := 0; i < 10; i++ {
		wlprintf(w, InfoLevel, "%s, %d during async writer 1k buff size\n", timeNow(), i)
	}
	if err := w.Close(); err != nil {
		t.Errorf("async close error: %+v", err)
	}
}

func TestAsyncWriterSmall(t *testing.T) {
	w := &AsyncWriter{
		ChannelSize: 5,
		Writer:      IOWriter{os.Stderr},
	}
	for i := 0; i < 10; i++ {
		wlprintf(w, InfoLevel, "%s, %d during async writer 1k buff size\n", timeNow(), i)
	}
	if err := w.Close(); err != nil {
		t.Errorf("async close error: %+v", err)
	}
}

func BenchmarkAsyncWriter(b *testing.B) {
	logger := Logger{
		Writer: &AsyncWriter{
			ChannelSize: 100,
			Writer:      IOWriter{ioutil.Discard},
		},
	}
	b.SetParallelism(1000)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			logger.Info().Msg("hello async writer")
		}
	})
}
