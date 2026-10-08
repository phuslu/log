package log

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

func TestAsyncWriterFanout(t *testing.T) {
	for _, name := range []string{"single", "two-async", "async-sync", "sync-async", "repeated-async", "multi-level", "nested-async"} {
		t.Run(name, func(t *testing.T) {
			var buffers [2]bytes.Buffer
			first := &AsyncWriter{Writer: IOWriter{&buffers[0]}}
			second := &AsyncWriter{Writer: IOWriter{&buffers[1]}}
			var writer Writer
			level := InfoLevel
			counts := []int{32, 32}
			switch name {
			case "single":
				writer = first
				counts[1] = 0
			case "two-async":
				writer = &MultiEntryWriter{first, second}
			case "async-sync":
				writer = &MultiEntryWriter{first, IOWriter{&buffers[1]}}
			case "sync-async":
				writer = &MultiEntryWriter{IOWriter{&buffers[0]}, second}
			case "repeated-async":
				writer = &MultiEntryWriter{first, first}
				counts = []int{64, 0}
			case "multi-level":
				writer = &MultiLevelWriter{ErrorWriter: first, InfoWriter: second}
				level = ErrorLevel
			case "nested-async":
				writer = &AsyncWriter{Writer: &MultiEntryWriter{first, second}}
			}
			logger := Logger{Writer: writer}
			for i := range 32 {
				logger.WithLevel(level).Int("sequence", i).Msg("fanout message")
			}
			if err := writer.(io.Closer).Close(); err != nil {
				t.Fatal(err)
			}
			for i, count := range counts {
				decoder := json.NewDecoder(&buffers[i])
				for j := range count {
					var record struct {
						Message  string `json:"message"`
						Sequence int    `json:"sequence"`
						Level    string `json:"level"`
					}
					if err := decoder.Decode(&record); err != nil {
						t.Fatalf("sink %d entry %d: %v", i, j, err)
					}
					sequence := j
					if name == "repeated-async" {
						sequence /= 2
					}
					if record.Message != "fanout message" || record.Sequence != sequence || record.Level != level.String() {
						t.Fatalf("sink %d entry %d: %+v", i, j, record)
					}
				}
				var extra any
				if err := decoder.Decode(&extra); err != io.EOF {
					t.Fatalf("sink %d has unexpected extra data: %v, %v", i, extra, err)
				}
			}
		})
	}
}

func TestAsyncWriterFanoutRejectedEntry(t *testing.T) {
	for _, name := range []string{"closed", "full"} {
		t.Run(name, func(t *testing.T) {
			hw := newHoldingWriter()
			async := &AsyncWriter{Writer: hw, ChannelSize: 1, DiscardOnFull: true}
			wantErr := ErrAsyncWriterClosed
			if name == "closed" {
				if err := async.Close(); err != nil {
					t.Fatal(err)
				}
			} else {
				fillAsyncWriter(t, async, hw, 1)
				wantErr = ErrAsyncWriterFull
				defer func() {
					close(hw.release)
					if err := async.Close(); err != nil {
						t.Error(err)
					}
				}()
			}
			var buffer bytes.Buffer
			multi := &MultiEntryWriter{async, IOWriter{&buffer}}
			var n int
			var writeErr error
			logger := Logger{Writer: WriterFunc(func(e *Entry) (int, error) {
				n, writeErr = multi.WriteEntry(e)
				return n, writeErr
			})}
			logger.Info().Msg("still written to sibling")
			if writeErr != wantErr {
				t.Fatalf("WriteEntry error = %v, want %v", writeErr, wantErr)
			}
			if n != buffer.Len() || n == 0 {
				t.Fatalf("WriteEntry n = %d, sibling bytes = %d", n, buffer.Len())
			}
			var record struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(buffer.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if record.Message != "still written to sibling" {
				t.Fatalf("sibling message = %q", record.Message)
			}
		})
	}
}
