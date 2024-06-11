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

func TestAsyncWriterSize(t *testing.T) {
	writer1 := &FileWriter{
		Filename: "async_file_test1.log",
	}

	writer2 := &AsyncWriter{
		ChannelSize:    4096,
		WritevDisabled: false,
		DiscardOnFull:  false,
		Writer: &FileWriter{
			Filename: "async_file_test2.log",
		},
	}

	logger := Logger{
		Writer: &MultiEntryWriter{
			writer1,
			writer2,
		},
	}

	for i := 0; i < 100000; i++ {
		logger.Info().Msg("hello file writer")
	}

	if err := writer1.Close(); err != nil {
		t.Errorf("file writer close error: %+v", err)
	}

	if err := writer2.Close(); err != nil {
		t.Errorf("async file writer close error: %+v", err)
	}

	fi1, err := os.Stat(writer1.Filename)
	if err != nil {
		t.Errorf("file writer stat error: %+v", err)
	}

	fi2, err := os.Stat(writer2.Writer.(*FileWriter).Filename)
	if err != nil {
		t.Errorf("async file writer stat error: %+v", err)
	}

	if fi1.Size() != fi2.Size() {
		t.Errorf("filesize not equal: %v != %v", fi1.Size(), fi2.Size())
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
			WritevDisabled: false,
			DiscardOnFull:  false,
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
