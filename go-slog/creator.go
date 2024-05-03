package _go_slog

import (
	"io"
	"log/slog"

	"github.com/madkins23/go-slog/infra"
	"github.com/phuslu/log"
)

// creatorPhusluSlog returns a PhusluSlog object for the [phuslu/slog] handler.
func creatorPhusluSlog() infra.Creator {
	return infra.NewCreator("phuslu/slog", creatorPhusluSloghandlerFn, nil,
		`^phuslu/slog^ is a wrapper around the pre-existing ^phuslu/log^ logging library.`,
		map[string]string{
			"phuslu/log": "https://github.com/phuslu/log",
		})
}

func creatorPhusluSloghandlerFn(w io.Writer, options *slog.HandlerOptions) slog.Handler {
	return log.SlogNewJSONHandler(w, options)
}
