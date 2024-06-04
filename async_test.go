package log

import (
	"io"
	"os"
	"testing"
)

func TestAsyncWriterZero(t *testing.T) {
	w := &AsyncWriter{
		ChannelSize: 0,
		Writer:      IOWriter{os.Stderr},
	}
	for i := 0; i < 10; i++ {
		_, _ = wlprintf(w, InfoLevel, "%s, %d during async writer 1k buff size\n", timeNow(), i)
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
		_, _ = wlprintf(w, InfoLevel, "%s, %d during async writer 1k buff size\n", timeNow(), i)
	}
	if err := w.Close(); err != nil {
		t.Errorf("async close error: %+v", err)
	}
}

func BenchmarkSyncFileWriter(b *testing.B) {
	logger := Logger{
		Writer: &FileWriter{
			Filename: "sync_file_test.log",
		},
	}
	defer logger.Writer.(io.Closer).Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			logger.Info().Msg("hello file writer")
		}
	})
}

func BenchmarkAsyncFileWriter(b *testing.B) {
	logger := Logger{
		Writer: &AsyncWriter{
			ChannelSize:    4096,
			WritevDisabled: true,
			Writer: &FileWriter{
				Filename: "async_file_test.log",
			},
		},
	}
	defer logger.Writer.(io.Closer).Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			logger.Info().Msg("hello file writer")
		}
	})
}

func BenchmarkAsyncFileWriterWriteV(b *testing.B) {
	logger := Logger{
		Writer: &AsyncWriter{
			ChannelSize:    4096,
			WritevDisabled: false,
			Writer: &FileWriter{
				Filename: "async_file_test2.log",
			},
		},
	}
	defer logger.Writer.(io.Closer).Close()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(b *testing.PB) {
		for b.Next() {
			logger.Info().Msg("hello file writer")
		}
	})
}
