package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"
	"strconv"
)

const (
	gray   = 90
	red    = 91
	green  = 92
	yellow = 93
	cyan   = 96
)

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, "\033[0m")
}

type PrettyHandler struct {
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	return &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts),
		l:       stdLog.New(out, "", 0),
	}
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	fields := make(map[string]interface{}, r.NumAttrs())

	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()

		return true
	})

	for _, a := range h.attrs {
		fields[a.Key] = a.Value.Any()
	}

	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	colorCode := code(r.Level)

	h.l.Println(
		r.Time.Format("[15:05:05.000]"),
		colorize(colorCode, level),
		colorize(colorCode, r.Message),
		colorize(gray, string(b)),
	)

	return nil
}

func code(level slog.Level) int {
	switch level {
	case slog.LevelDebug:
		return cyan
	case slog.LevelInfo:
		return green
	case slog.LevelWarn:
		return yellow
	case slog.LevelError:
		return red
	}

	return 0
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}
