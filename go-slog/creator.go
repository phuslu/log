package _go_slog

import (
	"io"
	"log/slog"

	"github.com/madkins23/go-slog/infra"
	"github.com/phuslu/log"
)

// creatorSlogJson returns a Creator object for the [slog/JSONHandler] handler.
//
// [slog/JSONHandler]: https://pkg.go.dev/log/slog#JSONHandler
func creatorSlogJson() infra.Creator {
	return infra.NewCreator("slog/JSONHandler", creatorSlogJsonHandlerFn, nil,
		`^slog/JSONHandler^ is the JSON handler provided with the ^slog^ library.
		It is fast and as a part of the Go distribution it is used
		along with published documentation as a model for ^slog.Handler^ behavior.`,
		map[string]string{
			"slog/JSONHandler": "https://pkg.go.dev/log/slog#JSONHandler",
		})
}

func creatorSlogJsonHandlerFn(w io.Writer, options *slog.HandlerOptions) slog.Handler {
	return slog.NewJSONHandler(w, options)
}

// creatorPhusluSlog returns a Creator object for the [phuslu/slog] handler.
//
// [phuslu/slog]: https://github.com/phuslu/log/blob/master/slog.go
func creatorPhusluSlog() infra.Creator {
	return infra.NewCreator("phuslu/slog", creatorPhusluSlogHandlerFn, nil,
		`^phuslu/slog^ is a wrapper around the pre-existing ^phuslu/log^ logging library.`,
		map[string]string{
			"phuslu/log": "https://github.com/phuslu/log",
		})
}

func creatorPhusluSlogHandlerFn(w io.Writer, options *slog.HandlerOptions) slog.Handler {
	return log.SlogNewJSONHandler(w, options)
}
