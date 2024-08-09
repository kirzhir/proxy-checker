package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"
)

type PrettyHandler struct {
	slog.Handler
	l      *stdLog.Logger
	attrs  []slog.Attr
	colors map[slog.Level]string
}

func NewPrettyHandler(out io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	return &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts),
		l:       stdLog.New(out, "", 0),
		colors: map[slog.Level]string{
			slog.LevelDebug: "\033[36m", // Gray
			slog.LevelInfo:  "\033[92m", // Cyan
			slog.LevelWarn:  "\033[93m", // Yellow
			slog.LevelError: "\033[91m", // Red
		},
	}
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	colorCode := h.colors[r.Level]
	resetCode := "\033[0m"

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

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := fmt.Sprintf("%s%s", colorCode, r.Message)

	h.l.Println(
		timeStr,
		level,
		msg,
		fmt.Sprintf("%s%s", resetCode, string(b)),
	)

	return nil
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   attrs,
	}
}
