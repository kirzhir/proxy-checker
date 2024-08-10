package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestColorize(t *testing.T) {
	coloredText := colorize(green, "info")
	expected := "\033[92minfo\033[0m"
	if coloredText != expected {
		t.Errorf("expected %s, got %s", expected, coloredText)
	}
}

func TestPrettyHandler_Handle(t *testing.T) {
	var buf bytes.Buffer
	handler := NewPrettyHandler(&buf, &slog.HandlerOptions{})

	tests := []struct {
		level   slog.Level
		message string
		color   int
	}{
		{slog.LevelInfo, "Info message", green},
		{slog.LevelError, "Error message", red},
		{slog.LevelDebug, "Debug message", cyan},
		{slog.LevelWarn, "Warn message", yellow},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			buf.Reset()
			record := slog.NewRecord(time.Now(), tt.level, tt.message, 0)
			record.AddAttrs(slog.String("key", "value"))

			err := handler.Handle(context.Background(), record)
			if err != nil {
				t.Fatalf("Handle failed: %v", err)
			}

			output := buf.String()
			if !contains(output, tt.message) {
				t.Errorf("expected output to contain log message %s, got: %s", tt.message, output)
			}

			coloredLevel := colorize(tt.color, tt.level.String()+":")
			if !contains(output, coloredLevel) {
				t.Errorf("expected output to contain colored level %s, got: %s", coloredLevel, output)
			}

			if !contains(output, "key") || !contains(output, "value") {
				t.Errorf("expected output to contain attributes, got: %s", output)
			}
		})
	}
}

func TestPrettyHandler_WithAttrs(t *testing.T) {
	var buf bytes.Buffer

	handler := NewPrettyHandler(&buf, &slog.HandlerOptions{})
	handlerWithAttrs := handler.WithAttrs([]slog.Attr{slog.String("globalKey", "globalValue")})

	record := slog.NewRecord(time.Now(), slog.LevelWarn, "Test message", 0)
	record.AddAttrs(slog.String("key1", "value1"))

	err := handlerWithAttrs.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	output := buf.String()
	if !contains(output, "globalKey") || !contains(output, "globalValue") {
		t.Errorf("expected output to contain global attributes, got: %s", output)
	}

	if !contains(output, "key1") || !contains(output, "value1") {
		t.Errorf("expected output to contain log attributes, got: %s", output)
	}
}

func contains(str, substr string) bool {
	return bytes.Contains([]byte(str), []byte(substr))
}
